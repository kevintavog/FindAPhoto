package main

import (
	"flag"
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

	"runtime/pprof"
)

var memprofile = flag.String("memprofile", "", "write memory profile to this file")

func main() {
	runtime.GOMAXPROCS(4)
	common.InitDirectories("FindAPhoto")
	common.ConfigureLogging(common.LogDirectory, "findaphotoindexer")

	app := cli.App("indexer", "The FindAPhoto indexer")
	app.Spec = "-p -s -a -o -k [-i]"
	indexPrefix := app.StringOpt("i", "", "The prefix for the index")
	alias := app.StringOpt("a alias", "", "The alias (prefix) to use for the path")
	scanPath := app.StringOpt("p path", "", "The path to recursively index")
	server := app.StringOpt("s server", "", "The URL for the ElasticSearch server")
	openStreetMapServer := app.StringOpt("o osm", "", "The URL for the OpenStreetMap server")
	key := app.StringOpt("k key", "", "The OpenStreetMap/MapQuest key")
	app.Action = func() {
		common.MediaIndexName = *indexPrefix + common.MediaIndexName

		log.Info("%s: FindAPhoto scanning %s, alias=%s) and indexing to %s/%s",
			time.Now().Format("2006-01-02"),
			*scanPath,
			*alias,
			*server,
			common.MediaIndexName)
		log.Info("Using %s to resolve locations to placename", *openStreetMapServer)

		common.ElasticSearchServer = *server
		resolveplacename.OpenStreetMapUrl = *openStreetMapServer
		resolveplacename.OpenStreetMapKey = *key

		checkServerAndIndex()

		scanStartTime := time.Now()
		scanner.Scan(*scanPath, *alias)
		scanDuration := time.Now().Sub(scanStartTime).Seconds()
		emitStats(scanDuration)

		if *memprofile != "" {
			log.Info("Emitting memory dump to %v", *memprofile)
			f, err := os.Create(*memprofile)
			if err != nil {
				log.Fatal(err)
			}
			pprof.WriteHeapProfile(f)
			f.Close()
		}
	}

	app.Run(os.Args)
}

func emitStats(seconds float64) {
	filesPerSecond := int64(float64(scanner.SupportedFilesFound) / seconds)

	log.Info("[%01.3f seconds, %d files/second], Scanned %d folders and %d files, found %d supported files.",
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

	log.Info("%d files indexed, %d failed and %d were added due to detected changes",
		indexmedia.IndexedFiles, indexmedia.FailedIndexAttempts, indexmedia.ChangedFiles)
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
}
