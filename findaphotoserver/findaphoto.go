package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-playground/lars"
	"github.com/ian-kent/go-log/log"
	"github.com/jawher/mow.cli"
	"gopkg.in/olivere/elastic.v3"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
	"github.com/kevintavog/findaphoto/findaphotoserver/configuration"
	"github.com/kevintavog/findaphoto/findaphotoserver/controllers/api"
	"github.com/kevintavog/findaphoto/findaphotoserver/controllers/files"
)

var _ lars.IAppContext = &applicationglobals.ApplicationGlobals{} // ensures ApplicationGlobals complies with lasr.IGlobals at compile time
var logDirectory = ""

func main() {
	configuration.ReadConfiguration()
	common.ConfigureLogging(common.LogDirectory, "findaphotoserver")

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

	log.Info("Listening on port %d, using %s/%s", listenPort, configuration.Current.ElasticSearchUrl, common.MediaIndexName)
	log.Info(" Using %s for openmap reverse lookups", configuration.Current.OpenMapUrl)

	common.ElasticSearchServer = configuration.Current.ElasticSearchUrl

	checkElasticServerAndIndex()
	checkOpenMapServer()

	wd, _ := os.Getwd()
	log.Info("Serving html/css/js content from %s/%s", wd, "content")
	contentDir := http.Dir("./content/")
	_, e := contentDir.Open("index.html")
	if e != nil {
		log.Fatal("Unable to get files from the './content' folder: %s\n", e.Error())
	}
	fs := http.FileServer(contentDir)

	l := configureApplicationGlobals()
	l.Get("/", fs)
	l.Get("/*", fs)
	l.Get("/search", redirectToTop)  // '/search' is an Angular route
	l.Get("/slide/*", redirectToTop) // '/slide/' is an Angular route
	api.ConfigureRouting(l)
	files.ConfigureRouting(l)

	startServerFunc := func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", listenPort), l.Serve())
		if err != nil {
			log.Fatal("Failed starting the service: %s", err.Error())
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

func redirectToTop(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
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
