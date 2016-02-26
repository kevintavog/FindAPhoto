package api

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/go-playground/lars"

	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
)

type InvalidRequest struct {
	message string
}

func (ir *InvalidRequest) Error() string {
	return ir.message
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
			if ir, ok := r.(InvalidRequest); ok {
				app.Error(http.StatusInternalServerError, "InvalidRequest", ir.Error(), nil)
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

func Error(c *lars.Context) {

	time.Sleep(time.Second / 10)
	app := c.AppContext.(*applicationglobals.ApplicationGlobals)
	app.Error(http.StatusInternalServerError, "bogus", "This is the message", nil)
}
