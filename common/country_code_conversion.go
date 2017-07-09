package common

import (
	"strings"
)

var countryCodeToName = map[string]string{
	"be": "Belgium",
	"ca": "Canada",
	"de": "Germany",
	"es": "Spain",
	"fi": "Finland",
	"fr": "France",
	"gb": "Great Britain",
	"is": "Iceland",
	"mx": "Mexico",
	"us": "USA",
}

func ConvertToCountryName(code string, defValue string) string {
	name, ok := countryCodeToName[strings.ToLower(code)]
	if ok {
		return name
	}
	return defValue
}
