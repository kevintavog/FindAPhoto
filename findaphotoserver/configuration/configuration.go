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
	ElasticSearchUrl     string
	RedisUrl             string
	OpenMapUrl           string
	OpenMapKey           string
	VipsExists           bool
	DefaultIndexPath     string
	ClarifaiClientId     string
	ClarifaiClientSecret string
}

var Current Configuration

func ReadConfiguration() {

	common.InitDirectories("FindAPhoto")
	configDirectory := common.ConfigDirectory

	configFile := path.Join(configDirectory, "rangic.findaphotoService")
	_, err := os.Stat(configFile)
	if err != nil {
		defaults := &Configuration{
			ElasticSearchUrl:     "elastic search url (http://somehost:9200)",
			RedisUrl:             "redis url (redis://somehost:6379)",
			OpenMapUrl:           "open map cache url (http://localhost:2000)",
			OpenMapKey:           "key goes here",
			ClarifaiClientId:     "clarifai.com client id goes here",
			ClarifaiClientSecret: "clarifai.com secret goes here",
		}
		json, jerr := json.Marshal(defaults)
		if jerr != nil {
			log.Fatal("Config file (%s) doesn't exist; attempt to write defaults failed: %s", configFile, jerr.Error())
		}

		werr := ioutil.WriteFile(configFile, json, os.ModePerm)
		if werr != nil {
			log.Fatal("Config file (%s) doesn't exist; attempt to write defaults failed: %s", configFile, werr.Error())
		} else {
			log.Fatal("Config file (%s) doesn't exist; one was written with defaults", configFile)
		}
	} else {
		viper.SetConfigFile(configFile)
		viper.SetConfigType("json")
		err := viper.ReadInConfig()
		if err != nil {
			log.Fatal("Error reading config file (%s): %s", configFile, err.Error())
		}
		err = viper.Unmarshal(&Current)
		if err != nil {
			log.Fatal("Failed converting configuration from (%s): %s", configFile, err.Error())
		}
	}

	Current.VipsExists = common.IsExecWorking(common.VipsThumbnailPath, "--vips-version")
}
