package main

import (
	"os/exec"
	"time"

	"golang.org/x/net/context"

	"github.com/ian-kent/go-log/log"
	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/configuration"

	"gopkg.in/olivere/elastic.v5"
)

func runIndexer(force bool, devMode bool) {
	log.Info("Starting indexing")

	var args = []string{
		"-s", configuration.Current.ElasticSearchUrl,
		"-l", configuration.Current.LocationLookupUrl,
		"-r", configuration.Current.RedisUrl,
		"-a", common.AliasPathOverride}

	if force {
		args = append(args, "--reindex")
	}

	if devMode {
		args = append(args, "-i")
		args = append(args, "dev-")
	}

	countPaths := 0
	common.VisitAllPaths(func(alias common.AliasDocument) {
		countPaths++
		timeAndRunIndexer(args, alias.Path)
	})

	// If the repository is empty and there's a default path, index the default path
	if countPaths == 0 {
		if len(configuration.Current.DefaultIndexPath) > 0 {
			log.Info("Run the indexer for the default path: '%v'", configuration.Current.DefaultIndexPath)
			timeAndRunIndexer(args, configuration.Current.DefaultIndexPath)
		} else {
			log.Warn("No paths to index")
		}
	}

	// Re-load the aliases - at least update the last indexed timestamp
	client, err := elastic.NewSimpleClient(
		elastic.SetURL(common.ElasticSearchServer),
		elastic.SetSniff(false))

	if err != nil {
		log.Fatalf("Unable to connect to '%s': %s", common.ElasticSearchServer, err.Error())
	}
	_, err = client.Refresh(common.MediaIndexName).Do(context.TODO())
	if err != nil {
		log.Error("Failed refreshing '%s': %s", common.MediaIndexName, err)
	}
	common.InitializeAliases(client)
}

func timeAndRunIndexer(args []string, path string) {

	startTime := time.Now()

	pathAndArgs := append(args, "-p")
	pathAndArgs = append(pathAndArgs, path)
	err := exec.Command(common.IndexerPath, pathAndArgs...).Run()
	duration := time.Now().Sub(startTime).Seconds()
	log.Info("Finished indexing '%v' in %1.1f seconds", path, duration)

	if err != nil {
		log.Error("Failed executing indexer for '%s': %s", path, err.Error())
	}
}
