package files

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/go-playground/lars"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
)

const baseMediaUrl = "/files/media/"

func ToMediaUrl(aliasedPath string) string {
	return baseMediaUrl + url.QueryEscape(strings.Replace(aliasedPath, "\\", "/", -1))
}

func Media(c lars.Context) {
	fc := c.(*applicationglobals.FpContext)
	fc.AppContext.FieldLogger.Time("media", func() {
		mediaUrl := fc.Ctx.Request().URL.Path
		if !strings.HasPrefix(strings.ToLower(mediaUrl), baseMediaUrl) {
			fc.AppContext.FieldLogger.Add("missingMediaPrefix", "true")
			fc.Ctx.Response().WriteHeader(http.StatusNotFound)
			return
		}

		// The path must exist in the repository
		mediaPath := mediaUrl[len(baseMediaUrl):]
		mediaId, err := toRepositoryId(mediaPath)
		if err != nil {
			fc.Error(http.StatusNotFound, "invalidMediaId", "", err)
			return
		}

		client := common.CreateClient()
		searchResult, err := client.Search().
			Index(common.MediaIndexName).
			Type(common.MediaTypeName).
			Query(elastic.NewTermQuery("_id", mediaId)).
			Do(context.TODO())
		if err != nil {
			fc.AppContext.FieldLogger.Add("invalidMediaPrefix", "true")
			fc.Ctx.Response().WriteHeader(http.StatusNotFound)
			return
		}

		if searchResult.TotalHits() == 0 {
			fc.AppContext.FieldLogger.Add("notInRepository", "true")
			fc.Ctx.Response().WriteHeader(http.StatusNotFound)
			return
		}

		mediaFilename, err := aliasedToFullPath(mediaPath)
		if err != nil {
			fc.Error(http.StatusNotFound, "badAlias", "", err)
			return
		}

		if exists, _ := common.FileExists(mediaFilename); !exists {
			fc.AppContext.FieldLogger.Add("missingMedia", "true")
			fc.Ctx.Response().WriteHeader(http.StatusNotFound)
			return
		}

		http.ServeFile(fc.Ctx.Response().ResponseWriter, fc.Ctx.Request(), mediaFilename)
	})
}
