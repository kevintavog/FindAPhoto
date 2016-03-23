package files

import (
	"bytes"
	"image"
	"image/jpeg"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/go-playground/lars"
	"github.com/nfnt/resize"
	"gopkg.in/olivere/elastic.v3"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
)

const baseSlideUrl = "/files/slides/"
const slideMaxHeightDimension = 800

func ToSlideUrl(aliasedPath string) string {
	return baseSlideUrl + url.QueryEscape(strings.Replace(aliasedPath, "\\", "/", -1))
}

func Slides(c lars.Context) {
	fc := c.(*applicationglobals.FpContext)
	fc.AppContext.FieldLogger.Time("slide", func() {
		slideUrl := fc.Ctx.Request().URL.Path
		if !strings.HasPrefix(strings.ToLower(slideUrl), baseSlideUrl) {
			fc.AppContext.FieldLogger.Add("missingSlidePrefix", "true")
			fc.Ctx.Response().WriteHeader(http.StatusNotFound)
			return
		}

		// The path must exist in the repository
		slidePath := slideUrl[len(baseSlideUrl):]
		slideId, err := toRepositoryId(slidePath)
		if err != nil {
			fc.Error(http.StatusNotFound, "invalidSlideId", "", err)
			return
		}

		client := common.CreateClient()
		searchResult, err := client.Search().
			Index(common.MediaIndexName).
			Type(common.MediaTypeName).
			Query(elastic.NewTermQuery("_id", slideId)).
			Do()
		if err != nil {
			fc.AppContext.FieldLogger.Add("invalidSlidePrefix", "true")
			fc.Ctx.Response().WriteHeader(http.StatusNotFound)
			return
		}

		if searchResult.TotalHits() == 0 {
			fc.AppContext.FieldLogger.Add("notInRepository", "true")
			fc.Ctx.Response().WriteHeader(http.StatusNotFound)
			return
		}

		// Convert alias to filename, generate temporary slide, return it
		slideFilename, err := aliasedToFullPath(slidePath)
		if err != nil {
			fc.Error(http.StatusNotFound, "badAlias", "", err)
			return
		}

		if exists, _ := common.FileExists(slideFilename); !exists {
			fc.AppContext.FieldLogger.Add("missingMedia", "true")
			fc.Ctx.Response().WriteHeader(http.StatusNotFound)
			return
		}

		//		buffer, err := resizeImaging(slideFilename)
		buffer, err := resizeWithNfnt(slideFilename)
		if err != nil {
			fc.Error(http.StatusInternalServerError, "failedSlideGeneration", "", err)
			return
		}

		http.ServeContent(fc.Ctx.Response().ResponseWriter, fc.Ctx.Request(), slideFilename, time.Unix(0, 0), bytes.NewReader(buffer.Bytes()))
	})
}

func resizeImaging(imageFilename string) (bytes.Buffer, error) {
	var buffer bytes.Buffer

	image, err := imaging.Open(imageFilename)
	if err != nil {
		return buffer, err
	}

	slideImage := imaging.Resize(image, 0, slideMaxHeightDimension, imaging.Box)
	err = imaging.Encode(&buffer, slideImage, imaging.JPEG)
	if err != nil {
		return buffer, err
	}

	return buffer, nil
}

func resizeWithNfnt(imageFilename string) (bytes.Buffer, error) {
	var buffer bytes.Buffer

	file, err := os.Open(imageFilename)
	if err != nil {
		return buffer, err
	}
	defer file.Close()

	image, _, err := image.Decode(file)
	if err != nil {
		return buffer, err
	}

	slideImage := resize.Resize(0, slideMaxHeightDimension, image, resize.NearestNeighbor)
	jpeg.Encode(&buffer, slideImage, &jpeg.Options{Quality: 85})

	return buffer, nil
}
