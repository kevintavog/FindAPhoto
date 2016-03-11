package api

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/go-playground/lars"

	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
)

type InternalError struct {
	message string
	err     error
}

type InvalidRequest struct {
	message string
	err     error
}

func (ir *InvalidRequest) Error() string {
	return ir.message
}

func (ie *InternalError) Error() string {
	return ie.message
}

func ConfigureRouting(l *lars.LARS) {
	l.Use(HandleErrors)

	api := l.Group("/api")
	api.Get("/search", Search)
}

func HandleErrors(c *lars.Context) {

	app := c.AppContext.(*applicationglobals.ApplicationGlobals)
	defer func() {
		if r := recover(); r != nil {
			if ie, ok := r.(*InternalError); ok {
				app.Error(http.StatusInternalServerError, "InternalError", ie.Error(), ie.err)
			} else if ir, ok := r.(*InvalidRequest); ok {
				app.Error(http.StatusBadRequest, "InvalidRequest", ir.Error(), ir.err)
			} else if e, ok := r.(runtime.Error); ok {
				app.Error(http.StatusInternalServerError, "UnhandledError", "", e)
			} else if s, ok := r.(string); ok {
				app.Error(http.StatusInternalServerError, "UnhandledError", s, nil)
			} else {
				app.Error(http.StatusInternalServerError, "UnhandledError", fmt.Sprintf("%v", r), nil)
			}
		}
	}()

	c.Next()
}
