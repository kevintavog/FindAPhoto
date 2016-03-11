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

func Thumbs(c *lars.Context) {
	app := c.AppContext.(*applicationglobals.ApplicationGlobals)

	app.FieldLogger.Time("thumb", func() {
		thumbFilename := c.Request.URL.Path
		if !strings.HasPrefix(strings.ToLower(thumbFilename), baseThumbUrl) {
			app.FieldLogger.Add("missingThumbPrefix", "true")
			c.Response.WriteHeader(http.StatusNotFound)
			return
		}

		thumbFilename = thumbFilename[len(baseThumbUrl):]
		thumbFilename = path.Clean(path.Join(common.ThumbnailDirectory, thumbFilename))
		if !strings.HasPrefix(thumbFilename, common.ThumbnailDirectory) {
			app.FieldLogger.Add("invalidThumbPrefix", "true")
			c.Response.WriteHeader(http.StatusNotFound)
			return
		}

		if exists, _ := common.FileExists(thumbFilename); !exists {
			app.FieldLogger.Add("missingThumbnail", "true")
			http.ServeFile(c.Response.ResponseWriter, c.Request, "./content/images/MissingThumbnail.png")
			return
		}
		http.ServeFile(c.Response.ResponseWriter, c.Request, thumbFilename)
	})
}
