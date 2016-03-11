package files

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/go-playground/lars"
	"gopkg.in/olivere/elastic.v3"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
)

const baseMediaUrl = "/files/media/"

func ToMediaUrl(aliasedPath string) string {
	return baseMediaUrl + url.QueryEscape(strings.Replace(aliasedPath, "\\", "/", -1))
}

func Media(c *lars.Context) {
	app := c.AppContext.(*applicationglobals.ApplicationGlobals)

	app.FieldLogger.Time("media", func() {
		mediaUrl := c.Request.URL.Path
		if !strings.HasPrefix(strings.ToLower(mediaUrl), baseMediaUrl) {
			app.FieldLogger.Add("missingMediaPrefix", "true")
			c.Response.WriteHeader(http.StatusNotFound)
			return
		}

		// The path must exist in the repository
		mediaPath := mediaUrl[len(baseMediaUrl):]
		mediaId, err := toRepositoryId(mediaPath)
		if err != nil {
			app.Error(http.StatusNotFound, "invalidMediaId", "", err)
			return
		}

		client := common.CreateClient()
		searchResult, err := client.Search().
			Index(common.MediaIndexName).
			Type(common.MediaTypeName).
			Query(elastic.NewTermQuery("_id", mediaId)).
			Do()
		if err != nil {
			app.FieldLogger.Add("invalidMediaPrefix", "true")
			c.Response.WriteHeader(http.StatusNotFound)
			return
		}

		if searchResult.TotalHits() == 0 {
			app.FieldLogger.Add("notInRepository", "true")
			c.Response.WriteHeader(http.StatusNotFound)
			return
		}

		mediaFilename, err := aliasedToFullPath(mediaPath)
		if err != nil {
			app.Error(http.StatusNotFound, "badAlias", "", err)
			return
		}

		if exists, _ := common.FileExists(mediaFilename); !exists {
			app.FieldLogger.Add("missingMedia", "true")
			c.Response.WriteHeader(http.StatusNotFound)
			return
		}

		http.ServeFile(c.Response.ResponseWriter, c.Request, mediaFilename)
	})
}
