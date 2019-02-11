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

	"github.com/labstack/echo"
	"github.com/nfnt/resize"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/configuration"
	"github.com/kevintavog/findaphoto/findaphotoserver/util"
	"github.com/twinj/uuid"
)

const baseSlideUrl = "/files/slides/"
const slideMaxHeightDimension = 800

func ToSlideUrl(aliasedPath string) string {
	return baseSlideUrl + url.QueryEscape(strings.Replace(aliasedPath, "\\", "/", -1))
}

func slideFiles(c echo.Context) error {
	fc := c.(*util.FpContext)
	return fc.Time("slide", func() error {
		slideURL := c.Request().URL.Path
		if !strings.HasPrefix(strings.ToLower(slideURL), baseSlideUrl) {
			fc.LogBool("missingSlidePrefix", true)
			return c.NoContent(http.StatusNotFound)
		}

		// The path must exist in the repository
		slidePath := slideURL[len(baseSlideUrl):]
		slideID, err := toRepositoryId(slidePath)
		if err != nil {
			return util.ErrorJSON(c, http.StatusNotFound, "invalidSlideId", "", err)
		}

		client := common.CreateClient()
		searchResult, err := client.Search().
			Index(common.MediaIndexName).
			Type(common.MediaTypeName).
			Query(elastic.NewTermQuery("_id", slideID)).
			Do(context.TODO())
		if err != nil {
			fc.LogBool("invalidSlidePrefix", true)
			return c.NoContent(http.StatusNotFound)
		}

		if searchResult.TotalHits() == 0 {
			fc.LogBool("notInRepository", true)
			return c.NoContent(http.StatusNotFound)
		}

		// Convert alias to filename, generate temporary slide, return it
		slideFilename, err := aliasedToFullPath(slidePath)
		if err != nil {
			return util.ErrorJSON(c, http.StatusNotFound, "badAlias", "", err)
		}

		if exists, _ := common.FileExists(slideFilename); !exists {
			fc.LogBool("missingMedia", true)
			return c.NoContent(http.StatusNotFound)
		}

		var buffer bytes.Buffer
		if configuration.Current.VipsExists {
			buffer, err = generateVipsSlide(slideFilename)
		} else {
			buffer, err = generateNfntSlide(slideFilename)
		}

		if err != nil {
			return util.ErrorJSON(c, http.StatusInternalServerError, "failedSlideGeneration", "", err)
		}

		return c.Stream(http.StatusOK, "", bytes.NewReader(buffer.Bytes()))
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
