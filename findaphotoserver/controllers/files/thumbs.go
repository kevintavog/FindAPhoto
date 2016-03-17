package files

import (
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/go-playground/lars"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
)

const baseThumbUrl = "/files/thumbs/"

func ToThumbUrl(aliasedPath string) string {
	thumbUrl := baseThumbUrl + url.QueryEscape(strings.Replace(aliasedPath, "\\", "/", -1))
	if strings.ToUpper(path.Ext(thumbUrl)) != ".JPG" {
		thumbUrl += ".JPG"
	}
	return thumbUrl
}

func Thumbs(c lars.Context) {
	fc := c.(*applicationglobals.FpContext)
	fc.AppContext.FieldLogger.Time("thumb", func() {
		thumbFilename := fc.Ctx.Request().URL.Path
		if !strings.HasPrefix(strings.ToLower(thumbFilename), baseThumbUrl) {
			fc.AppContext.FieldLogger.Add("missingThumbPrefix", "true")
			fc.Ctx.Response().WriteHeader(http.StatusNotFound)
			return
		}

		thumbFilename = thumbFilename[len(baseThumbUrl):]
		thumbFilename = path.Clean(path.Join(common.ThumbnailDirectory, thumbFilename))
		if !strings.HasPrefix(thumbFilename, common.ThumbnailDirectory) {
			fc.AppContext.FieldLogger.Add("invalidThumbPrefix", "true")
			fc.Ctx.Response().WriteHeader(http.StatusNotFound)
			return
		}

		thumbFilename, err := url.QueryUnescape(thumbFilename)
		if err != nil {
			fc.AppContext.FieldLogger.Add("badlyEscaped", "true")
			fc.AppContext.FieldLogger.Add("badlyEscapedError", err.Error())
			fc.Ctx.Response().WriteHeader(http.StatusNotFound)
			return
		}

		if exists, _ := common.FileExists(thumbFilename); !exists {
			fc.AppContext.FieldLogger.Add("missingThumbnail", "true")
			http.ServeFile(fc.Ctx.Response().ResponseWriter, fc.Ctx.Request(), "./content/images/MissingThumbnail.png")
			return
		}
		http.ServeFile(fc.Ctx.Response().ResponseWriter, fc.Ctx.Request(), thumbFilename)
	})
}
