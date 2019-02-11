package api

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/util"
	"github.com/labstack/echo"
)

func mediaByIdAPI(c echo.Context) error {
	fc := c.(*util.FpContext)
	return fc.Time("mediabyid", func() error {
		id := c.Param("id")

		client := common.CreateClient()
		result, err := client.Search().
			Index(common.MediaIndexName).
			Type(common.MediaTypeName).
			Query(elastic.NewTermQuery("_id", id)).
			Do(context.TODO())

		if err != nil {
			panic(&util.InvalidRequest{Message: "SearchFailed", Err: err})
		}

		if result.TotalHits() == 0 {
			return util.ErrorJSON(c, http.StatusNotFound, "NoSuchId", "", nil)
		} else if result.TotalHits() > 1 {
			return util.ErrorJSON(c, http.StatusInternalServerError, "DuplicateIdFound", "", nil)
		} else {
			source := make(map[string]interface{})
			hit := result.Hits.Hits[0]
			err := json.Unmarshal(*hit.Source, &source)
			if err != nil {
				panic(&util.InvalidRequest{Message: "JsonUnmarshallFailed", Err: err})
			}
			return c.JSON(http.StatusOK, source)
		}
	})
}
