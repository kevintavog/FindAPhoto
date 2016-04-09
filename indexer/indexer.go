package main

import (
	_ "net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/indexer/steps/checkindex"
	"github.com/kevintavog/findaphoto/indexer/steps/checkthumbnail"
	"github.com/kevintavog/findaphoto/indexer/steps/generatethumbnail"
	"github.com/kevintavog/findaphoto/indexer/steps/getexif"
	"github.com/kevintavog/findaphoto/indexer/steps/indexmedia"
	"github.com/kevintavog/findaphoto/indexer/steps/resolveplacename"
	"github.com/kevintavog/findaphoto/indexer/steps/scanner"

	"github.com/ian-kent/go-log/log"
	"github.com/jawher/mow.cli"
	"gopkg.in/olivere/elastic.v3"

	_ "runtime/pprof"
)

func main() {
	common.InitDirectories("FindAPhoto")
	common.ConfigureLogging(common.LogDirectory, "findaphotoindexer")

	// go http.ListenAndServe(":8080", nil)

	app := cli.App("indexer", "The FindAPhoto indexer")
	app.Spec = "-p -s -o -k [-i] [-c]"
	indexPrefix := app.StringOpt("i", "", "The prefix for the index (for development) (optional)")
	scanPath := app.StringOpt("p path", "", "The path to recursively index")
	server := app.StringOpt("s server", "", "The URL for the ElasticSearch server")
	openStreetMapServer := app.StringOpt("o osm", "", "The URL for the OpenStreetMap server")
	cachedLocationsServer := app.StringOpt("c", "", "The URL for the cached location server (optional)")
	key := app.StringOpt("k key", "", "The OpenStreetMap/MapQuest key")
	app.Action = func() {

		common.MediaIndexName = *indexPrefix + common.MediaIndexName

		log.Info("%s: FindAPhoto scanning %s, and indexing to %s/%s; using %d/%d CPU's",
			time.Now().Format("2006-01-02"),
			*scanPath,
			*server,
			common.MediaIndexName,
			runtime.NumCPU(),
			runtime.GOMAXPROCS(0))
		log.Info("Using %s to resolve locations to placename", *openStreetMapServer)

		common.ElasticSearchServer = *server
		resolveplacename.OpenStreetMapUrl = *openStreetMapServer
		resolveplacename.OpenStreetMapKey = *key
		resolveplacename.CachedLocationsUrl = *cachedLocationsServer

		checkServerAndIndex()
		alias, err := common.AliasForPath(*scanPath)
		if err != nil {
			log.Fatalf("Unable to get alias for '%s': %s", *scanPath, err.Error())
		}

		scanStartTime := time.Now()
		scanner.Scan(*scanPath, alias)
		scanDuration := time.Now().Sub(scanStartTime).Seconds()
		emitStats(scanDuration)
	}

	app.Run(os.Args)
}

func emitStats(seconds float64) {
	filesPerSecond := float64(scanner.SupportedFilesFound) / seconds

	log.Info("[%01.3f seconds, %01.2f files/second], Scanned %d folders and %d files, found %d supported files.",
		seconds, filesPerSecond,
		scanner.DirectoriesScanned, scanner.FilesScanned, scanner.SupportedFilesFound)

	log.Info("%d failed repository checks, %d badly formatted json responses, %d failed signatures",
		checkindex.BadJson, checkindex.CheckFailed, checkindex.SignatureGenerationFailed)

	log.Info("%d exiftool invocations, %d failed",
		getexif.ExifToolInvocations, getexif.ExifToolFailed)

	log.Info("%d locations lookup attempts, %d location lookup failures, %d server errors, %d other failures",
		resolveplacename.PlacenameLookups, resolveplacename.FailedLookups, resolveplacename.ServerErrors, resolveplacename.Failures)

	log.Info("%d image thumbnails created, %d failed; %d video thumbnails created, %d failed; %d failed thumbnail checks",
		generatethumbnail.GeneratedImage, generatethumbnail.FailedImage, generatethumbnail.GeneratedVideo, generatethumbnail.FailedVideo, checkthumbnail.FailedChecks)

	log.Info("%d files indexed, %d failed and %d added due to detected changes",
		indexmedia.IndexedFiles, indexmedia.FailedIndexAttempts, indexmedia.ChangedFiles)

	log.Info("%d media scanned, %d removed from the index",
		scanner.MediaScanned, scanner.MediaRemoved)
}

func checkServerAndIndex() {
	client, err := elastic.NewSimpleClient(
		elastic.SetURL(common.ElasticSearchServer),
		elastic.SetSniff(false))

	if err != nil {
		log.Fatal("Unable to connect to '%s': %s", common.ElasticSearchServer, err.Error())
	}

	exists, err := client.IndexExists(common.MediaIndexName).Do()
	if err != nil {
		log.Fatal("Failed querying index: %s", err.Error())
	}
	if !exists {
		log.Warn("The index '%s' doesn't exist", common.MediaIndexName)
		err = common.CreateFindAPhotoIndex(client)
		if err != nil {
			log.Fatal("Failed creating index '%s': %+v", common.MediaIndexName, err.Error())
		}
	}

	err = common.InitializeAliases(client)
	if err != nil {
		log.Fatal("Failed initializing aliases: %s", err.Error())
	}
}
