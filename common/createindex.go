package common

import (
	"errors"

	"github.com/ian-kent/go-log/log"
	"gopkg.in/olivere/elastic.v3"
)

func CreateFindAPhotoIndex(client *elastic.Client) error {
	log.Warn("Creating index '%s'", MediaIndexName)

	mapping := `{
		"settings": {
			"number_of_shards": 1,
	        "number_of_replicas": 0
		},
		"mappings": {
			"media": {
				"_all": {
					"enabled": false
			    },
				"properties" : {
					"date" : {
					  "type" : "string"
					},
					"datetime" : {
					  "type" : "date",
					  "format": "yyyy-MM-dd'T'HH:mm:ssZ"
					},
					"filename" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"location" : {
					  "type" : "geo_point"
					},
					"lengthinbytes" : {
					  "type" : "long"
					},
					"path" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"signature" : {
					  "type" : "string",
					  "index" : "not_analyzed"
					},
					
					"aperture" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"exposureprogram" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"exposuretime" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"flash" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"fnumber" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"focallength" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"iso" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"whitebalance" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"lensinfo" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"lensmodel" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"cameramake" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"cameramodel" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					
					"cityname" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"sitename" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"countryname" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					},
					"countrycode" : {
					  "type" : "string",
			          "fields": {
			            "value": { 
			              "type":  "string",
			              "index": "not_analyzed"
			            }
					  }
					}
				}
			}
		}
	}`

	response, err := client.CreateIndex(MediaIndexName).BodyString(mapping).Do()
	if err != nil {
		return err
	}

	if response.Acknowledged != true {
		return errors.New("Index creation not acknowledged")
	}
	return nil
}
