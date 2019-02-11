package util

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/labstack/echo"
	"gopkg.in/olivere/elastic.v3"
)

// Recover returns a middleware which recovers from panics anywhere in the chain
// and handles the control to the centralized HTTPErrorHandler.
func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {

					logStack := true
					if ie, ok := r.(*InternalError); ok {
						ErrorJSON(c, http.StatusInternalServerError, "InternalError", ie.Error(), ie.Err)
					} else if ir, ok := r.(*InvalidRequest); ok {
						logStack = false
						ErrorJSON(c, http.StatusBadRequest, "InvalidRequest", ir.Error(), ir.Err)
					} else if ee, ok := r.(*elastic.Error); ok {
						logStack = false
						ErrorJSON(c, http.StatusBadRequest, "ElasticError", GetDetailedErrorMessage(ee), ee)
					} else if e, ok := r.(runtime.Error); ok {
						ErrorJSON(c, http.StatusInternalServerError, "UnhandledError", "", e)
					} else if s, ok := r.(string); ok {
						ErrorJSON(c, http.StatusInternalServerError, "UnhandledError", s, nil)
					} else {
						ErrorJSON(c, http.StatusInternalServerError, "UnhandledError", fmt.Sprintf("%v", r), nil)
					}

					if logStack {
						buf := make([]byte, 1<<16)
						stackSize := runtime.Stack(buf, false)
						fc := c.(*FpContext)
						fc.Log("stack", string(buf[0:stackSize]))
					}

					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					c.Error(err)
				}
			}()
			return next(c)
		}
	}
}
