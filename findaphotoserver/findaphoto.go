package main

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/go-playground/lars"
	"github.com/ian-kent/go-log/appenders"
	"github.com/ian-kent/go-log/layout"
	"github.com/ian-kent/go-log/log"
	"github.com/jawher/mow.cli"
	"gopkg.in/olivere/elastic.v3"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
	"github.com/kevintavog/findaphoto/findaphotoserver/configuration"
	"github.com/kevintavog/findaphoto/findaphotoserver/controllers/api"
)

var _ lars.IAppContext = &applicationglobals.ApplicationGlobals{} // ensures ApplicationGlobals complies with lasr.IGlobals at compile time
var logDirectory = ""

func main() {
	configuration.ReadConfiguration()
	configureLogging()

	app := cli.App("findaphotoserver", "The FindAPhoto server")
	app.Spec = "[-d]"
	debugMode := app.BoolOpt("d", false, "Debug mode (hit <enter> to exit, listen on a different port, use a different index)")
	app.Action = func() { run(*debugMode) }

	app.Run(os.Args)
}

func run(debugMode bool) {
	listenPort := 2000
	easyExit := false

	if debugMode {
		fmt.Println("Using debug mode")
		common.MediaIndexName = "dev-" + common.MediaIndexName
		listenPort = 5000
		easyExit = true
	}

	fmt.Printf("Listening on port %d, using %s/%s", listenPort, configuration.Current.ElasticSearchUrl, common.MediaIndexName)
	fmt.Println()
	fmt.Printf(" Using %s for openmap reverse lookups", configuration.Current.OpenMapUrl)
	fmt.Println()

	common.ElasticSearchServer = configuration.Current.ElasticSearchUrl

	checkElasticServerAndIndex()
	checkOpenMapServer()

	l := configureApplicationGlobals()
	l.Get("/", Home)
	api.ConfigureRouting(l)

	startServerFunc := func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", listenPort), l.Serve())
		if err != nil {
			fmt.Println("Failed starting the service: %s", err.Error())
		}
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

func Home(c *lars.Context) {

	app := c.AppContext.(*applicationglobals.ApplicationGlobals)
	app.Error(http.StatusForbidden, "", "", nil)
}

func configureApplicationGlobals() *lars.LARS {
	globalsFn := func() lars.IAppContext {
		return &applicationglobals.ApplicationGlobals{}
	}

	l := lars.New()
	l.RegisterAppContext(globalsFn)
	return l
}

func checkElasticServerAndIndex() {
	client, err := elastic.NewClient(
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

func checkOpenMapServer() {

	url := fmt.Sprintf("%s/nominatim/v1/reverse?key=%s&format=json&lat=%f&lon=%f&addressdetails=1&zoom=18&accept-language=en-us",
		configuration.Current.OpenMapUrl, configuration.Current.OpenMapKey, 47.6216, -122.348133)

	_, err := http.Get(url)
	if err != nil {
		log.Fatal("The open street map values seem to be wrong, a location lookup failed: %s", err.Error())
	}
}

func configureLogging() {

	err := createDirectory(logDirectory)
	if err != nil {
		log.Fatal("Unable to create logging directory (%s): %s", logDirectory, err.Error())
	}

	logger := log.Logger("")

	lyt := layout.Pattern("%d %p: %m")
	layout.DefaultTimeLayout = "15:04:05.000000"

	rolling := appenders.RollingFile(path.Join(logDirectory, "findaphotoService.log"), true)
	rolling.MaxBackupIndex = 10
	rolling.MaxFileSize = 5 * 1024 * 1024
	rolling.SetLayout(lyt)

	console := appenders.Console()
	console.SetLayout(lyt)

	logger.SetAppender(appenders.Multiple(lyt, rolling, console))
}

func createDirectory(directory string) error {
	_, err := os.Stat(directory)
	if err != nil {
		return nil
	}

	return os.MkdirAll(directory, os.ModeDir)
}
