package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/lars"

	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
)

func MediaById(c lars.Context) {
	fc := c.(*applicationglobals.FpContext)
	err := fc.Ctx.ParseForm()
	if err != nil {
		panic(&InvalidRequest{message: "parseFormError", err: err})
	}

	fc.AppContext.FieldLogger.Time("mediabyid", func() {
		id := fc.Ctx.Request().Form.Get("id")

		client := common.CreateClient()
		result, err := client.Search().
			Index(common.MediaIndexName).
			Type(common.MediaTypeName).
			Query(elastic.NewTermQuery("_id", id)).
			Do(context.TODO())

		if err != nil {
			panic(&InternalError{message: "SearchFailed", err: err})
		}

		if result.TotalHits() == 0 {
			fc.Error(http.StatusNotFound, "NoSuchId", "", nil)
		} else if result.TotalHits() > 1 {
			fc.Error(http.StatusInternalServerError, "DuplicateIdFound", "", nil)
		} else {
			source := make(map[string]interface{})
			hit := result.Hits.Hits[0]
			err := json.Unmarshal(*hit.Source, &source)
			if err != nil {
				panic(&InternalError{message: "JsonUnmarshallFailed", err: err})
			}
			fc.WriteResponse(source)
		}
	})
}
