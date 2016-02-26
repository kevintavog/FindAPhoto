package applicationglobals

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/lars"
	"github.com/kevintavog/findaphoto/findaphotoserver/fieldlogger"
)

type ApplicationGlobals struct {
	FieldLogger *fieldlogger.FieldLogger
	context     *lars.Context
}

func (g *ApplicationGlobals) WriteResponse(m map[string]interface{}) {

	g.context.Response.Header().Set(lars.ContentType, lars.ApplicationJSON)
	g.context.Response.WriteString(g.ToJson(m))
}

func (g *ApplicationGlobals) ToJson(m map[string]interface{}) string {
	return g.toJson(m, func(err error) {
		g.Error(http.StatusInternalServerError, "JsonConversionFailed", "", err)
	})
}

func (g *ApplicationGlobals) toJson(m map[string]interface{}, errorHandler func(error)) string {
	json, err := json.Marshal(m)
	if err != nil {
		errorHandler(err)
		return "{}"
	} else {
		return string(json)
	}
}

func (g *ApplicationGlobals) Error(httpCode int, errorCode, errorMessage string, err error) {
	g.FieldLogger.Error(errorMessage, err)
	g.FieldLogger.Add("errorCode", errorCode)

	g.context.Response.Header().Set(lars.ContentType, lars.ApplicationJSON)
	g.context.Response.WriteHeader(httpCode)

	data := map[string]interface{}{"errorCode": errorCode, "errorMessage": errorMessage}
	if err != nil {
		data["internalError"] = err.Error()
	}

	var ok = true
	json := g.toJson(data, func(err error) {
		g.FieldLogger.Add("marhsalError", err.Error())
		g.context.Response.WriteString(fmt.Sprintf("{\"errorCode\":\"%s\"", errorCode))
	})
	if ok {
		g.context.Response.WriteString(string(json))
	}
}

// Reset gets called just before a new HTTP request starts calling
// middleware + handlers
func (g *ApplicationGlobals) Reset(c *lars.Context) {
	g.context = c
	g.FieldLogger = fieldlogger.New()
}

// Done gets called after the HTTP request has completed right before
// Context gets put back into the pool
func (g *ApplicationGlobals) Done() {
	g.FieldLogger.Add("url", g.context.Request.RequestURI)
	g.FieldLogger.Add("method", g.context.Request.Method)
	g.FieldLogger.Close(g.context.Response.Status())
}
