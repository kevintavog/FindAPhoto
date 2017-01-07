package helpers

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/indexer/steps"

	"github.com/ian-kent/go-log/log"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
)

var DuplicateCheckFailed int64
var DuplicatesIgnored int64

type expiringItem struct {
	insertTime time.Time
	signature  string
}

const expirationCheckLengthSeconds = 5
const expireDurationSeconds = 20

var recentItems = make(map[string]string)
var recentItemsLock sync.Mutex
var itemList = list.New()
var nextExpirationCheck = time.Now().UTC().Add(expirationCheckLengthSeconds * time.Second)

// Check for duplicates. To handle the delay from indexing to being available for searching,
// track everything for the last few seconds ('expireDurationSeconds')
func IsDuplicate(client *elastic.Client, signature string, aliasedPath string, markAsIndexed bool) bool {

	if shouldCheckExpiredItems() {
		checkAndRemoveExpiredItems()
	}

	// Is this a recent item?
	if isRecent, original := isRecentItem(signature); isRecent {
		atomic.AddInt64(&DuplicatesIgnored, 1)
		AddDuplicateToIndex(client, aliasedPath, original)
		return true
	}

	signatureExistsResult, err := client.Search().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Query(elastic.NewTermQuery("signature", signature)).
		Pretty(true).
		Do(context.TODO())
	if err != nil {
		atomic.AddInt64(&DuplicateCheckFailed, 1)
		log.Error("Error checking signature existence for '%s': %s", aliasedPath, err.Error())
		return false
	}

	signatureExists := signatureExistsResult.TotalHits() > 0

	pathSearchResult, err := client.Search().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Query(elastic.NewTermQuery("_id", aliasedPath)).
		Pretty(true).
		Do(context.TODO())
	if err != nil {
		atomic.AddInt64(&DuplicateCheckFailed, 1)
		log.Error("Error checking document existence for '%s': %s", aliasedPath, err.Error())
		return false
	}
	pathExists := pathSearchResult.TotalHits() > 0

	if signatureExists && !pathExists {
		media := returnFirstMatch(signatureExistsResult)
		if media != nil {
			AddDuplicateToIndex(client, aliasedPath, media.Path)

			fullPath, err := common.FullPathForAliasedPath(media.Path)
			if err != nil {
				log.Error("Failed getting full path from alias: %s: (%s)", media.Path, err)
			} else {
				classifymedia.Enqueue(fullPath, media.Path, media.Tags)
			}
		} else {
			log.Error("Unable to get original path from duplicate search result")
		}
		atomic.AddInt64(&DuplicatesIgnored, 1)
		return true
	}

	if markAsIndexed {
		recentItemsLock.Lock()
		defer recentItemsLock.Unlock()

		original, exists := recentItems[signature]
		if exists {
			AddDuplicateToIndex(client, aliasedPath, original)
			return true
		}
		recentItems[signature] = aliasedPath
		itemList.PushBack(&expiringItem{
			signature:  signature,
			insertTime: time.Now().UTC(),
		})
	}

	return false
}

func isRecentItem(signature string) (bool, string) {
	recentItemsLock.Lock()
	defer recentItemsLock.Unlock()

	original, ok := recentItems[signature]
	return ok, original
}

func shouldCheckExpiredItems() bool {
	return time.Now().UTC().After(nextExpirationCheck)
}

func checkAndRemoveExpiredItems() {
	recentItemsLock.Lock()
	defer recentItemsLock.Unlock()

	if shouldCheckExpiredItems() {
		nextExpirationCheck = time.Now().UTC().Add(expirationCheckLengthSeconds * time.Second)
		removeExpiredItems()
	}
}

func removeExpiredItems() {
	expiredTime := time.Now().UTC().Add(-expireDurationSeconds * time.Second)

	var next *list.Element
	for i := itemList.Front(); i != nil; i = next {
		item := i.Value.(*expiringItem)
		if item.insertTime.After(expiredTime) {
			break
		}

		next = i.Next()

		itemList.Remove(i)
		delete(recentItems, item.signature)
	}
}
