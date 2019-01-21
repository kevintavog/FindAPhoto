package applicationglobals

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/lars"
)

type FpContext struct {
	*lars.Ctx
	AppContext *ApplicationGlobals
}

func (fc *FpContext) Reset(w http.ResponseWriter, r *http.Request) {

	fc.Ctx.Reset(w, r)
	fc.AppContext.Reset()
}

func (fc *FpContext) RequestComplete() {
	fc.AppContext.FieldLogger.Add("url", fc.Ctx.Request().RequestURI)
	fc.AppContext.FieldLogger.Add("method", fc.Ctx.Request().Method)
	fc.AppContext.FieldLogger.Close(fc.Ctx.Response().Status())

	fc.AppContext.Done()
}

func NewContext(l *lars.LARS) lars.Context {
	return &FpContext{
		Ctx:        lars.NewContext(l),
		AppContext: newGlobals(),
	}
}

func (fc *FpContext) WriteResponse(m map[string]interface{}) {
	fc.Ctx.Response().Header().Set(lars.ContentType, lars.ApplicationJSON)
	fc.Ctx.Response().WriteString(fc.ToJson(m))
}

func (fc *FpContext) WriteStatus(httpCode int) {
	fc.Ctx.Response().Header().Set(lars.ContentType, lars.ApplicationJSON)
	fc.Ctx.Response().WriteHeader(httpCode)
}

func (fc *FpContext) ToJson(m map[string]interface{}) string {
	return toJson(m, func(err error) {
		fc.Error(http.StatusInternalServerError, "JsonConversionFailed", "", err)
	})
}

func toJson(m map[string]interface{}, errorHandler func(error)) string {
	json, err := json.Marshal(m)
	if err != nil {
		errorHandler(err)
		return "{}"
	} else {
		return string(json)
	}
}

func (fc *FpContext) Error(httpCode int, errorCode, errorMessage string, err error) {
	fc.AppContext.FieldLogger.Error(errorMessage, err)
	fc.AppContext.FieldLogger.Add("errorCode", errorCode)

	fc.Ctx.Response().Header().Set(lars.ContentType, lars.ApplicationJSON)
	fc.Ctx.Response().WriteHeader(httpCode)

	data := map[string]interface{}{"errorCode": errorCode, "errorMessage": errorMessage}
	if err != nil {
		data["internalError"] = err.Error()
	}

	var ok = true
	json := toJson(data, func(err error) {
		fc.AppContext.FieldLogger.Add("marhsalError", err.Error())
		fc.Ctx.Response().WriteString(fmt.Sprintf("{\"errorCode\":\"%s\"", errorCode))
	})
	if ok {
		fc.Ctx.Response().WriteString(string(json))
	}
}
