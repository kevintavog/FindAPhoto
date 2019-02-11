package util

import (
	"fmt"
	"strings"

	"github.com/labstack/echo"
	"gopkg.in/olivere/elastic.v3"
)

type InternalError struct {
	Message string
	Err     error
}

type InvalidRequest struct {
	Message string
	Err     error
}

func ErrorJSON(c echo.Context, httpCode int, errorCode, errorMessage string, err error) error {
	fc := c.(*FpContext)
	fc.LogError(errorMessage, err)
	fc.Log("errorCode", errorCode)

	data := map[string]interface{}{"errorCode": errorCode, "errorMessage": errorMessage}
	if err != nil {
		data["internalError"] = err.Error()
	}

	return c.JSON(httpCode, data)
}

func (ir *InvalidRequest) Error() string {
	detailed := GetDetailedErrorMessage(ir.Err)
	if len(detailed) > 0 {
		return fmt.Sprintf("%s -- %s", ir.Message, detailed)
	}
	return ir.Message
}

func (ie *InternalError) Error() string {
	detailed := GetDetailedErrorMessage(ie.Err)
	if len(detailed) > 0 {
		return fmt.Sprintf("%s -- %s", ie.Message, detailed)
	}
	return ie.Message
}

func GetDetailedErrorMessage(err error) string {
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

func PropogateError(err error, message string) {
	if err != nil {
		if ee, ok := err.(*elastic.Error); ok {
			panic(ee)
		}
		panic(&InternalError{Message: message, Err: err})
	}
}
