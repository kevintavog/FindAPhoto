package api

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/go-playground/lars"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
	"gopkg.in/olivere/elastic.v5"
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
	return ir.message + " -- " + getDetailedErrorMessage(ir.err)
}

func (ie *InternalError) Error() string {
	return ie.message + " -- " + getDetailedErrorMessage(ie.err)
}

func handleErrors(c lars.Context) {
	fc := c.(*applicationglobals.FpContext)
	defer func() {
		if r := recover(); r != nil {
			logStack := true
			if ie, ok := r.(*InternalError); ok {
				fc.Error(http.StatusInternalServerError, "InternalError", ie.Error(), ie.err)
			} else if ir, ok := r.(*InvalidRequest); ok {
				logStack = false
				fc.Error(http.StatusBadRequest, "InvalidRequest", ir.Error(), ir.err)
			} else if ee, ok := r.(*elastic.Error); ok {
				logStack = false
				fc.Error(http.StatusBadRequest, "ElasticError", getDetailedErrorMessage(ee), ee)
			} else if e, ok := r.(runtime.Error); ok {
				fc.Error(http.StatusInternalServerError, "UnhandledError", "", e)
			} else if s, ok := r.(string); ok {
				fc.Error(http.StatusInternalServerError, "UnhandledError", s, nil)
			} else {
				fc.Error(http.StatusInternalServerError, "UnhandledError", fmt.Sprintf("%v", r), nil)
			}

			if logStack {
				buf := make([]byte, 1<<16)
				stackSize := runtime.Stack(buf, false)
				fc.AppContext.FieldLogger.Add("stack", string(buf[0:stackSize]))
			}
		}
	}()

	fc.Ctx.Next()
}

func propogateError(err error, message string) {
	if err != nil {
		if ee, ok := err.(*elastic.Error); ok {
			panic(ee)
		}
		panic(&InternalError{message: message, err: err})
	}
}

func getDetailedErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	if ee, ok := err.(*elastic.Error); ok {
		details := fmt.Sprintf("StatusCode: %d", ee.Status)
		if ee.Details != nil {
			if ee.Details.CausedBy != nil {
				haveReason := false

				if causedBy, ok := ee.Details.CausedBy["caused_by"].(map[string]interface{}); ok {
					if reason, ok := causedBy["reason"].(string); ok {
						haveReason = true
						details = fmt.Sprintf("%s; %s", details, strings.Replace(reason, "\n", "", -1))
					}
				}

				if !haveReason {
					if reason, ok := ee.Details.CausedBy["reason"]; ok {
						details = fmt.Sprintf("%s; %s", details, reason)
					}
				}
			}
		}

		if ee.Details != nil && ee.Details.CausedBy != nil {
			return details
		}
	}
	return err.Error()
}
