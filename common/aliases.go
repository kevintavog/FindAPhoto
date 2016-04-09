package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ian-kent/go-log/log"
	"gopkg.in/olivere/elastic.v3"
)

// Artificially limits the number of aliases - can easily handle more, but the search needs to be updated
const maxAliasCount = 100

const AliasTypeName = "alias"

type AliasDocument struct {
	Path      string    `json:"path"`
	Alias     string    `json:"alias"`
	DateAdded time.Time `json:"datetimeadded"`
}

var aliasAndPath []AliasDocument

// Load all aliases
func InitializeAliases(client *elastic.Client) error {
	log.Info("Loading aliases from '%s'", MediaIndexName)
	return loadAliases(client)
}

func AliasForPath(path string) (string, error) {
	// Return an existing alias for the given path, add a new alias if necessary
	ad := findViaPath(path)
	if ad == nil {
		// re-load in case someone else added it recently
		err := loadAliases(CreateClient())
		if err != nil {
			return "", err
		}
		ad = findViaPath(path)
	}

	if ad != nil {
		return ad.Alias, nil
	}

	// Add a new alias to the index
	err := addNewAlias(path)
	if err != nil {
		return "", err
	}

	ad = findViaPath(path)
	if ad == nil {
		return "", errors.New(fmt.Sprintf("Can't find just added alias for '%s", path))
	}
	return ad.Alias, nil
}

func IsValidAliasedPath(aliased string) bool {
	alias, _ := extactAlias(aliased)
	return IsValidAlias(alias)
}

func IsValidAlias(alias string) bool {
	return findViaAlias(alias) != nil
}

func FullPathForAliasedPath(aliased string) (string, error) {
	// Given an alias, return the associated path
	alias, partialPath := extactAlias(aliased)
	ad := findViaAlias(alias)
	if ad == nil {
		// re-load in case someone else added it recently
		err := loadAliases(CreateClient())
		if err != nil {
			return "", err
		}
		ad = findViaAlias(alias)
	}
	if ad != nil {
		return path.Join(ad.Path, partialPath), nil
	}

	return "", errors.New(fmt.Sprintf("Unable to find path for %s", alias))
}

func PathForAlias(alias string) (string, error) {
	// Given an alias, return the associated path
	ad := findViaAlias(alias)
	if ad == nil {
		// re-load in case someone else added it recently
		err := loadAliases(CreateClient())
		if err != nil {
			return "", err
		}
		ad = findViaAlias(alias)
	}
	if ad != nil {
		return ad.Path, nil
	}

	return "", errors.New(fmt.Sprintf("Unable to find path for %s", alias))
}

func extactAlias(aliasAndPath string) (string, string) {
	pos := strings.Index(aliasAndPath, "\\")
	if pos == -1 {
		return aliasAndPath, ""
	}

	partialPath := strings.Replace(aliasAndPath[pos+1:], "\\", "//", -1)
	return aliasAndPath[0:pos], partialPath
}

func findViaPath(path string) *AliasDocument {
	for _, ad := range aliasAndPath {
		//		log.Error("findViaPath: '%s': '%s'", ad.Alias, ad.Path)
		if strings.EqualFold(path, ad.Path) {
			return &ad
		}
	}

	return nil
}

func findViaAlias(alias string) *AliasDocument {
	// Given an alias, return the associated path
	for _, ad := range aliasAndPath {
		//		log.Error("findViaAlias: '%s': '%s'", ad.Alias, ad.Path)
		if strings.EqualFold(alias, ad.Alias) {
			return &ad
		}
	}

	return nil
}

func loadAliases(client *elastic.Client) error {
	search := client.Search().
		Index(MediaIndexName).
		Type(AliasTypeName).
		Pretty(true).
		Query(elastic.NewMatchAllQuery()).
		Size(maxAliasCount).
		Sort("datetime", false) // Sort by created date, descending
	result, err := search.Do()
	if err != nil {
		return err
	}

	numAliases := result.TotalHits()
	//	log.Error("Load aliases - %d matches found", numAliases)
	if numAliases > maxAliasCount {
		log.Fatal("There are more aliases than can currently be handled: %d", numAliases)
	}

	aliasList := make([]AliasDocument, 0)
	if numAliases > 0 {
		for _, hit := range result.Hits.Hits {
			alias := &AliasDocument{}
			err := json.Unmarshal(*hit.Source, alias)
			if err != nil {
				log.Fatal("Unable to parse alias: %s", err.Error())
			}

			log.Info("Alias '%s' maps to '%s'", alias.Alias, alias.Path)
			aliasList = append(aliasList, *alias)
		}
	}

	aliasAndPath = aliasList
	return nil
}

func addNewAlias(path string) error {
	client := CreateClient()
	err := loadAliases(client)
	if err != nil {
		return err
	}

	var newAliasNumber int
	if len(aliasAndPath) == 0 {
		newAliasNumber = 1
	} else {
		latest := aliasAndPath[0]
		for _, ad := range aliasAndPath {
			if ad.DateAdded.Unix() > latest.DateAdded.Unix() {
				latest = ad
			}
		}
		newAliasNumber, err = strconv.Atoi(latest.Alias)
		if err != nil {
			return err
		}
		newAliasNumber += 1
	}

	ad := &AliasDocument{
		Path:      path,
		Alias:     fmt.Sprintf("%d", newAliasNumber),
		DateAdded: time.Now()}

	log.Warn("Adding alias '%s' for '%s'", ad.Alias, ad.Path)
	response, err := client.Index().
		Index(MediaIndexName).
		Type(AliasTypeName).
		Id(ad.Alias).
		BodyJson(ad).
		Do()
	if err != nil {
		return err
	}
	if !response.Created {
		return errors.New(fmt.Sprintf("Failed creating alias entry for new path '%s'", path))
	}

	aliasAndPath = append(aliasAndPath, *ad)
	return nil
}
