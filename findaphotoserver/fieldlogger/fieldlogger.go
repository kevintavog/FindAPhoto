package fieldlogger

import (
	"fmt"
	"strings"
	"time"

	"github.com/ian-kent/go-log/levels"
	"github.com/ian-kent/go-log/log"
)

type FieldLogger struct {
	fields    map[string]string
	level     levels.LogLevel
	startTime time.Time
}

func New() *FieldLogger {

	fl := &FieldLogger{
		fields:    make(map[string]string),
		level:     levels.INFO,
		startTime: time.Now(),
	}

	return fl
}

func (fl *FieldLogger) Error(message string, err error) {
	fl.level = levels.ERROR
	fl.Add("message", message)

	if err != nil {
		fl.Add("error", err.Error())
	}
}

func (fl *FieldLogger) Add(name, value string) {
	fl.fields[name] = value
}

func (fl *FieldLogger) Close(httpStatusCode int) {

	keyValues := make([]string, 0, len(fl.fields))

	durationMsecs := time.Now().Sub(fl.startTime).Seconds() * 1000
	keyValues = append(keyValues, fmt.Sprintf("httpStatusCode=%d", httpStatusCode))
	keyValues = append(keyValues, fmt.Sprintf("msec=%01.3f", durationMsecs))

	for key, value := range fl.fields {
		keyValues = append(keyValues, fmt.Sprintf("%s=%s", key, value))
	}

	log.Log(fl.level, "%s", strings.Join(keyValues, " "))
}

func (fl *FieldLogger) Time(name string, action func()) {
	action()
}
