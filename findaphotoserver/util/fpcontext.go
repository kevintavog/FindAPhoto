package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/labstack/echo"
)

// The context used for REST calls - to provide the field logger and convenience methods.
type FpContext struct {
	echo.Context
	fields *FieldLogger
}

func NewFpContext(c echo.Context) *FpContext {
	return &FpContext{
		Context: c,
		fields:  NewFieldLogger(),
	}
}

func (fc *FpContext) Float64FromQuery(name string) float64 {
	s := fc.QueryParam(name)
	if s != "" {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			panic(&InvalidRequest{Message: fmt.Sprintf("'%s' is not a float: %s", name, s)})
		}
		return v
	}

	panic(&InvalidRequest{Message: fmt.Sprintf("'%s' is missing from the query parameter", name)})
}

func (fc *FpContext) OptionalFloat64FromQuery(name string, defaultValue float64) float64 {
	s := fc.QueryParam(name)
	if s != "" {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			panic(&InvalidRequest{Message: fmt.Sprintf("'%s' is not a float: %s", name, s)})
		}
		return v
	}

	return defaultValue
}

func (fc *FpContext) IntFromQuery(name string, defaultValue int) int {
	s := fc.QueryParam(name)
	if s != "" {
		return IntFromString(name, s)
	}
	return defaultValue
}

func (fc *FpContext) BoolFromQuery(name string, defaultValue bool) bool {
	s := fc.QueryParam(name)
	if s != "" {
		v, err := strconv.ParseBool(s)
		if err != nil {
			panic(&InvalidRequest{Message: fmt.Sprintf("'%s' is not a bool: %s", name, s)})
		}
		return v
	}
	return defaultValue
}

func (fc *FpContext) Time(name string, action func() error) error {
	return fc.fields.time(name, action)
}

func (fc *FpContext) LogError(message string, err error) {
	fc.fields.error(message, err)
}

func (fc *FpContext) Log(name string, v string) {
	fc.fields.add(name, v)
}

func (fc *FpContext) LogBool(name string, v bool) {
	fc.fields.add(name, strconv.FormatBool(v))
}

func (fc *FpContext) LogInt(name string, v int) {
	fc.fields.add(name, strconv.Itoa(v))
}

func (fc *FpContext) LogInt64(name string, v int64) {
	fc.fields.add(name, strconv.FormatInt(v, 10))
}

func (fc *FpContext) LogStringArray(name string, v []string) {
	fc.fields.add(name, strings.Join(v, ","))
}

func (fc *FpContext) RequestComplete() {
	fc.fields.close(fc.Context)
}
