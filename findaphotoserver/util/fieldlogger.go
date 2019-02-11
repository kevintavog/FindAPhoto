package util

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ian-kent/go-log/levels"
	"github.com/ian-kent/go-log/log"
	"github.com/labstack/echo"
)

type FieldLogger struct {
	fields    map[string]string
	level     levels.LogLevel
	startTime time.Time
}

func NewFieldLogger() *FieldLogger {
	fl := &FieldLogger{
		fields:    make(map[string]string),
		level:     levels.INFO,
		startTime: time.Now(),
	}

	return fl
}

func (fl *FieldLogger) error(message string, err error) {
	fl.level = levels.ERROR
	fl.add("message", message)

	if err != nil {
		fl.add("error", err.Error())
	}
}

func (fl *FieldLogger) add(name, value string) {
	if strings.Contains(value, " ") {
		fl.fields[name] = "\"" + value + "\""
	} else {
		fl.fields[name] = value
	}
}

func (fl *FieldLogger) close(c echo.Context) {

	durationMsecs := time.Now().Sub(fl.startTime).Seconds() * 1000

	keyValues := make([]string, 0, len(fl.fields))
	keyValues = append(keyValues, fmt.Sprintf("statusCode=%d", c.Response().Status))
	keyValues = append(keyValues, fmt.Sprintf("method=%s", c.Request().Method))
	keyValues = append(keyValues, fmt.Sprintf("msec=%01.3f", durationMsecs))
	keyValues = append(keyValues, fmt.Sprintf("bytesOut=%d", c.Response().Size))
	keyValues = append(keyValues, fmt.Sprintf("path=%s", c.Request().URL.Path))
	keyValues = append(keyValues, fmt.Sprintf("query=%s", c.QueryString()))

	for key, value := range fl.fields {
		keyValues = append(keyValues, fmt.Sprintf("%s=%s", key, value))
	}

	if fl.level == levels.INFO && c.Response().Status >= http.StatusBadRequest {
		fl.level = levels.ERROR
	}
	log.Log(fl.level, "%s", strings.Join(keyValues, " "))
}

func (fl *FieldLogger) time(name string, action func() error) error {
	startTime := time.Now()
	err := action()
	durationMsecs := time.Now().Sub(startTime).Seconds() * 1000
	fl.add(name+"Msec", fmt.Sprintf("%01.3f", durationMsecs))
	return err
}
