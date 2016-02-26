package configuration

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"runtime"

	"github.com/ian-kent/go-log/log"
	"github.com/spf13/viper"
)

type Configuration struct {
	ElasticSearchUrl string
	OpenMapUrl       string
	OpenMapKey       string
	PathAndAliases   []PathAndAliasConfiguration
}

type PathAndAliasConfiguration struct {
	Path  string
	Alias string
}

var Current Configuration
var LogDirectory string

func ReadConfiguration() {

	// require: open map url, open map key, elastic search url

	var configDirectory string

	if runtime.GOOS == "darwin" {
		homeDirectory := os.Getenv("HOME")
		LogDirectory = path.Join(homeDirectory, "Library", "Logs", "FindAPhoto")
		configDirectory = path.Join(homeDirectory, "Library", "Preferences")

	} else if runtime.GOOS == "linux" {
		log.Fatal("Come up with directories for Linux!")
	} else {
		log.Fatal("Come up with directories for: %v", runtime.GOOS)
	}

	configFile := path.Join(configDirectory, "rangic.findaphotoService")
	_, err := os.Stat(configFile)
	if err != nil {
		defaultPaths := make([]PathAndAliasConfiguration, 2)
		defaultPaths[0].Path = "first path"
		defaultPaths[0].Alias = "1"
		defaultPaths[1].Path = "second path"
		defaultPaths[1].Alias = "2"

		defaults := &Configuration{
			ElasticSearchUrl: "provideUrl",
			OpenMapUrl:       "provideUrl",
			OpenMapKey:       "key goes here",
			PathAndAliases:   defaultPaths,
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
}
