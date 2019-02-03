package configuration

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/kevintavog/findaphoto/common"

	"github.com/ian-kent/go-log/log"
	"github.com/spf13/viper"
)

type Configuration struct {
	ElasticSearchURL  string `json:"ElasticSearchUrl"`
	RedisURL          string `json:"RedisUrl"`
	LocationLookupURL string `json:"LocationLookupUrl"`
	VipsExists        bool   `json:"VipsExists"`
	DefaultIndexPath  string `json:"DefaultIndexPath"`
	ClarifaiAPIKey    string `json:"ClarifaiApiKey"`
}

var Current Configuration

func ReadConfiguration() {
	common.InitDirectories("FindAPhoto")
	configDirectory := common.ConfigDirectory

	configFile := path.Join(configDirectory, "rangic.findaphotoService")
	_, err := os.Stat(configFile)
	if err != nil {
		defaults := &Configuration{
			ElasticSearchURL:  "elastic search url (http://somehost:9200)",
			RedisURL:          "redis url (redis://somehost:6379)",
			LocationLookupURL: "location lookup url",
			ClarifaiAPIKey:    "clarifai.com api key goes here",
		}
		json, jerr := json.Marshal(defaults)
		if jerr != nil {
			log.Fatalf("Config file (%s) doesn't exist; attempt to write defaults failed: %s", configFile, jerr.Error())
		}

		werr := ioutil.WriteFile(configFile, json, os.ModePerm)
		if werr != nil {
			log.Fatalf("Config file (%s) doesn't exist; attempt to write defaults failed: %s", configFile, werr.Error())
		} else {
			log.Fatalf("Config file (%s) doesn't exist; one was written with defaults", configFile)
		}
	} else {
		viper.SetConfigFile(configFile)
		viper.SetConfigType("json")
		err := viper.ReadInConfig()
		if err != nil {
			log.Fatalf("Error reading config file (%s): %s", configFile, err.Error())
		}
		err = viper.Unmarshal(&Current)
		if err != nil {
			log.Fatalf("Failed converting configuration from (%s): %s", configFile, err.Error())
		}
	}

	Current.VipsExists = common.IsExecWorking(common.VipsThumbnailPath, "--vips-version")
}
