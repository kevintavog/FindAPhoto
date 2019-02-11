package files

import (
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/util"
	"github.com/labstack/echo"
)

const baseThumbUrl = "/files/thumbs/"

func ToThumbUrl(aliasedPath string) string {
	thumbUrl := baseThumbUrl + url.QueryEscape(strings.Replace(aliasedPath, "\\", "/", -1))
	if strings.ToUpper(path.Ext(thumbUrl)) != ".JPG" {
		thumbUrl += ".JPG"
	}
	return thumbUrl
}

func thumbFiles(c echo.Context) error {
	fc := c.(*util.FpContext)
	return fc.Time("thumb", func() error {
		thumbFilename := c.Request().URL.Path
		if !strings.HasPrefix(strings.ToLower(thumbFilename), baseThumbUrl) {
			// fc.AppContext.FieldLogger.Add("missingThumbPrefix", "true")
			return c.NoContent(http.StatusNotFound)
		}

		thumbFilename = thumbFilename[len(baseThumbUrl):]
		thumbFilename = path.Clean(path.Join(common.ThumbnailDirectory, thumbFilename))
		if !strings.HasPrefix(thumbFilename, common.ThumbnailDirectory) {
			fc.LogBool("invalidThumbPrefix", true)
			return c.NoContent(http.StatusNotFound)
		}

		thumbFilename, err := url.QueryUnescape(thumbFilename)
		if err != nil {
			fc.LogBool("badlyEscaped", true)
			fc.Log("badlyEscapedError", err.Error())
			return c.NoContent(http.StatusNotFound)
		}

		if exists, _ := common.FileExists(thumbFilename); !exists {
			fc.LogBool("missingThumbnail", true)
			return c.File("./content/MissingThumbnail.png")
		}
		return c.File(thumbFilename)
	})
}
