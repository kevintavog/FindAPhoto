package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/lars"
	"github.com/ian-kent/go-log/log"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
	"github.com/kevintavog/findaphoto/findaphotoserver/configuration"
	"github.com/kevintavog/findaphoto/findaphotoserver/controllers/api"
	"github.com/kevintavog/findaphoto/findaphotoserver/controllers/files"
)

func run(devolopmentMode bool, indexOverride string, aliasOverride string) {
	listenPort := 2000
	easyExit := false
	skipMediaClassifier := false
	api.FindAPhotoVersionNumber = versionString()
	log.Info("FindAPhoto %s", api.FindAPhotoVersionNumber)

	if !common.IsExecWorking(common.ExifToolPath, "-ver") {
		log.Fatalf("exiftool isn't usable (path is '%s')", common.ExifToolPath)
	}
	if !common.IsExecWorking(common.FfmpegPath, "-version") {
		log.Fatalf("ffmpeg isn't usable (path is '%s')", common.FfmpegPath)
	}

	if devolopmentMode {
		fmt.Println("*** Using development mode ***")
		// common.MediaIndexName = "dev-" + common.MediaIndexName
		listenPort = 5000
		easyExit = true
		skipMediaClassifier = true
		if len(aliasOverride) > 0 {
			common.AliasPathOverride = aliasOverride
		}
	} else {
		if !common.IsExecWorking(common.IndexerPath, "-v") {
			log.Fatalf("The FindAPhoto Indexer isn't usable (path is '%s')", common.IndexerPath)
		}
		if !common.IsExecWorking(common.MediaClassifierPath, "-v") {
			log.Fatalf("The FindAPhoto Media Classifier isn't usable (path is '%s')", common.MediaClassifierPath)
		}
	}

	if len(indexOverride) > 0 {
		common.MediaIndexName = indexOverride
		fmt.Printf("*** Using index %s ***\n", common.MediaIndexName)
	}

	log.Info("Listening at http://localhost:%d/, For ElasticSearch, using: %s/%s", listenPort, configuration.Current.ElasticSearchUrl, common.MediaIndexName)
	log.Info(" Using %s for OpenStreetMap reverse lookups", configuration.Current.OpenMapUrl)

	common.ElasticSearchServer = configuration.Current.ElasticSearchUrl

	checkElasticServerAndIndex()
	checkOpenMapServer()

	l := configureApplicationGlobals()

	api.ConfigureRouting(l)
	files.ConfigureRouting(l)

	mediaClassifierFunc := func() {
		if !skipMediaClassifier {
			runMediaClassifier(devolopmentMode)
		} else {
			log.Info("Skipping media classifier")
		}
	}

	delayThenIndexFunc := func() {
		if !devolopmentMode {
			time.Sleep(1 * time.Second)
			runIndexer(devolopmentMode)
		}
	}

	startServerFunc := func() {
		go mediaClassifierFunc()
		go delayThenIndexFunc()

		err := http.ListenAndServe(fmt.Sprintf(":%d", listenPort), l.Serve())
		if err != nil {
			log.Fatalf("Failed starting the service: %s", err.Error())
		}
	}

	if !configuration.Current.VipsExists {
		log.Warn("Unable to use the 'vipsthumbnails' command, defaulting to slower slide generation (path is '%s')", common.VipsThumbnailPath)
	}

	if easyExit {
		go startServerFunc()

		fmt.Println("Hit enter to exit")
		var input string
		fmt.Scanln(&input)
	} else {
		startServerFunc()
	}
}

func configureApplicationGlobals() *lars.LARS {
	l := lars.New()
	l.RegisterContext(applicationglobals.NewContext)
	return l
}

func checkElasticServerAndIndex() {
	client, err := elastic.NewSimpleClient(
		elastic.SetURL(common.ElasticSearchServer),
		elastic.SetSniff(false))

	if err != nil {
		log.Fatalf("Unable to connect to '%s': %s", common.ElasticSearchServer, err.Error())
	}

	exists, err := client.IndexExists(common.MediaIndexName).Do(context.TODO())
	if err != nil {
		log.Fatalf("Failed querying index: %s", err.Error())
	}
	if !exists {
		log.Warn("The index '%s' doesn't exist", common.MediaIndexName)
		err = common.CreateFindAPhotoIndex(client)
		if err != nil {
			log.Fatalf("Failed creating index '%s': %+v", common.MediaIndexName, err.Error())
		}
	}

	err = common.InitializeAliases(client)
	if err != nil {
		log.Fatalf("Failed initializing aliases: %s", err.Error())
	}
}

func checkOpenMapServer() {
	url := fmt.Sprintf("%s/nominatim/v1/reverse?key=%s&format=json&lat=%f&lon=%f&addressdetails=1&zoom=18&accept-language=en-us",
		configuration.Current.OpenMapUrl, configuration.Current.OpenMapKey, 47.6216, -122.348133)

	_, err := http.Get(url)
	if err != nil {
		log.Fatalf("The open street map values seem to be wrong, a location lookup failed: %s", err.Error())
	}
}
