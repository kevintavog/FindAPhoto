package api

import (
	"github.com/go-playground/lars"
)

func Cors(context lars.Context) {

	if origin := context.Request().Header.Get("Origin"); origin != "" {
		context.Response().Header().Set("Access-Control-Allow-Origin", origin)
		context.Response().Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		context.Response().Header().Set("Access-Control-Allow-Headers", "Accept, Accept-Language, Content-Type")
	}

	if context.Request().Method == "OPTIONS" {
		return
	}

	context.Next()
}
