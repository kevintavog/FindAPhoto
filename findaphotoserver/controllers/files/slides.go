package files

import (
	"bytes"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/lars"
	"github.com/nfnt/resize"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
	"github.com/kevintavog/findaphoto/findaphotoserver/configuration"
	"github.com/twinj/uuid"
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
			Do(context.TODO())
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

		var buffer bytes.Buffer
		if configuration.Current.VipsExists {
			buffer, err = generateVipsSlide(slideFilename)
		} else {
			buffer, err = generateNfntSlide(slideFilename)
		}

		if err != nil {
			fc.Error(http.StatusInternalServerError, "failedSlideGeneration", "", err)
			return
		}

		http.ServeContent(fc.Ctx.Response().ResponseWriter, fc.Ctx.Request(), slideFilename, time.Unix(0, 0), bytes.NewReader(buffer.Bytes()))
	})
}

func generateNfntSlide(imageFilename string) (bytes.Buffer, error) {
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

func generateVipsSlide(imageFilename string) (bytes.Buffer, error) {
	var buffer bytes.Buffer

	tmpFilename := path.Join(os.TempDir(), "findAPhoto", "slides", uuid.NewV4().String()+".JPG")
	err := common.CreateDirectory(path.Dir(tmpFilename))
	if err != nil {
		return buffer, err
	}

	_, err = exec.Command(common.VipsThumbnailPath, "-d", "-s", "20000x"+strconv.Itoa(slideMaxHeightDimension), "-f", tmpFilename+"[optimize_coding,strip]", imageFilename).Output()
	if err != nil {
		return buffer, err
	}
	defer os.Remove(tmpFilename)

	ar, err := ioutil.ReadFile(tmpFilename)
	if err != nil {
		return buffer, err
	}

	return *bytes.NewBuffer(ar), nil
}
