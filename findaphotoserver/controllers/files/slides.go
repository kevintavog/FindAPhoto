package files

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/go-playground/lars"
	"gopkg.in/olivere/elastic.v3"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
)

const baseSlideUrl = "/files/slides/"
const slideMaxHeightDimension = 800

func ToSlideUrl(aliasedPath string) string {
	return baseSlideUrl + url.QueryEscape(strings.Replace(aliasedPath, "\\", "/", -1))
}

func Slides(c *lars.Context) {
	app := c.AppContext.(*applicationglobals.ApplicationGlobals)

	app.FieldLogger.Time("slide", func() {
		slideUrl := c.Request.URL.Path
		if !strings.HasPrefix(strings.ToLower(slideUrl), baseSlideUrl) {
			app.FieldLogger.Add("missingSlidePrefix", "true")
			c.Response.WriteHeader(http.StatusNotFound)
			return
		}

		// The path must exist in the repository
		slidePath := slideUrl[len(baseSlideUrl):]
		slideId, err := toRepositoryId(slidePath)
		if err != nil {
			app.Error(http.StatusNotFound, "invalidSlideId", "", err)
			return
		}

		client := common.CreateClient()
		searchResult, err := client.Search().
			Index(common.MediaIndexName).
			Type(common.MediaTypeName).
			Query(elastic.NewTermQuery("_id", slideId)).
			Do()
		if err != nil {
			app.FieldLogger.Add("invalidSlidePrefix", "true")
			c.Response.WriteHeader(http.StatusNotFound)
			return
		}

		if searchResult.TotalHits() == 0 {
			app.FieldLogger.Add("notInRepository", "true")
			c.Response.WriteHeader(http.StatusNotFound)
			return
		}

		// Convert alias to filename, generate temporary slide, return it
		slideFilename, err := aliasedToFullPath(slidePath)
		if err != nil {
			app.Error(http.StatusNotFound, "badAlias", "", err)
			return
		}

		if exists, _ := common.FileExists(slideFilename); !exists {
			app.FieldLogger.Add("missingMedia", "true")
			c.Response.WriteHeader(http.StatusNotFound)
			return
		}

		//		filename, err := resize(slideFilename)
		buffer, err := resize(slideFilename)
		if err != nil {
			app.Error(http.StatusInternalServerError, "failedSlideGeneration", "", err)
			return
		}

		http.ServeContent(c.Response.ResponseWriter, c.Request, slideFilename, time.Unix(0, 0), bytes.NewReader(buffer.Bytes()))

		//		defer os.Remove(filename)
		//		http.ServeFile(c.Response.ResponseWriter, c.Request, filename)
	})
}

//func resize(imageFilename string) (string, error) {
func resize(imageFilename string) (bytes.Buffer, error) {
	//	tmpFilename := path.Join(os.TempDir(), "findAPhoto", "slides", uuid.NewV4().String()+".JPG")

	var buffer bytes.Buffer
	//	err := common.CreateDirectory(path.Dir(tmpFilename))
	//	if err != nil {
	//		return nil, err
	//	}

	image, err := imaging.Open(imageFilename)
	if err != nil {
		return buffer, err
	}

	slideImage := imaging.Resize(image, 0, slideMaxHeightDimension, imaging.Box)
	//	var buffer bytes.Buffer
	//	writer := bufio.NewWriter(&buffer)

	err = imaging.Encode(&buffer, slideImage, imaging.JPEG)
	//	err = imaging.Save(slideImage, tmpFilename)
	if err != nil {
		//		os.Remove(tmpFilename)
		return buffer, err
	}

	return buffer, nil
	//	return tmpFilename, nil
}
