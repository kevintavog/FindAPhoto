package files

import (
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/util"
	"github.com/labstack/echo"
)

const baseMediaUrl = "/files/media/"

func ToMediaUrl(aliasedPath string) string {
	return baseMediaUrl + url.QueryEscape(strings.Replace(aliasedPath, "\\", "/", -1))
}

func mediaFiles(c echo.Context) error {
	fc := c.(*util.FpContext)
	return fc.Time("media", func() error {
		mediaURL := c.Request().URL.Path
		if !strings.HasPrefix(strings.ToLower(mediaURL), baseMediaUrl) {
			fc.LogBool("missingMediaPrefix", true)
			return c.NoContent(http.StatusNotFound)
		}

		// The path must exist in the repository
		mediaPath := mediaURL[len(baseMediaUrl):]
		mediaID, err := toRepositoryId(mediaPath)
		if err != nil {
			return util.ErrorJSON(c, http.StatusNotFound, "invalidMediaId", "", err)
		}

		client := common.CreateClient()
		searchResult, err := client.Search().
			Index(common.MediaIndexName).
			Type(common.MediaTypeName).
			Query(elastic.NewTermQuery("_id", mediaID)).
			Do(context.TODO())
		if err != nil {
			fc.LogBool("invalidMediaPrefix", true)
			return c.NoContent(http.StatusNotFound)
		}

		if searchResult.TotalHits() == 0 {
			fc.LogBool("notInRepository", true)
			return c.NoContent(http.StatusNotFound)
		}

		mediaFilename, err := aliasedToFullPath(mediaPath)
		if err != nil {
			return util.ErrorJSON(c, http.StatusNotFound, "badAlias", "", err)
		}

		if exists, _ := common.FileExists(mediaFilename); !exists {
			fc.LogBool("missingMedia", true)
			return c.NoContent(http.StatusNotFound)
		}

		return c.File(mediaFilename)
	})
}
