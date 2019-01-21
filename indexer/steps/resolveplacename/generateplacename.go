package resolveplacename

import (
	"github.com/kevintavog/findaphoto/common"

	"github.com/Jeffail/gabs"
)

func generatePlacename(media *common.Media, address *gabs.Container, displayName *string) {
	countryCode, ok := address.Path("country_code").Data().(string)
	if ok {
		media.LocationCountryCode = countryCode
	}
	countryName, ok := address.Path("country").Data().(string)
	if ok {
		media.LocationCountryName = countryName
	}
	media.LocationCityName = firstMatch(prioritizedCityNameComponents, address)
	media.LocationSiteName = firstMatch(prioritizedSiteComponents, address)
	media.LocationStateName = firstMatch(prioritizedStateNameComponents, address)

	media.LocationHierarchicalName = joinSkipEmpty(",", media.LocationSiteName, media.LocationCityName, media.LocationStateName, media.LocationCountryName)
	media.LocationPlaceName = placename(address)

	if displayName != nil {
		media.LocationDisplayName = *displayName
	}
}

func firstMatch(list []string, address *gabs.Container) string {
	for _, s := range list {
		value, ok := address.Path(s).Data().(string)
		if ok {
			return value
		}
	}

	return ""
}

func placename(address *gabs.Container) string {
	placename := ""
	children, _ := address.ChildrenMap()
	for _, value := range children {
		placename += value.Data().(string) + " "
	}
	return placename
}

// For 'state' name
var prioritizedStateNameComponents = []string{
	"state",
	"state_district",
}

// For 'city' name
var prioritizedCityNameComponents = []string{
	"city",
	"city_district",
	"town",
	"hamlet",
	"locality",
	"neighbourhood",
	"suburb",
	"village",
	"county",
}

// For the point of interest / site / building
var prioritizedSiteComponents = []string{
	"playground",

	"aerodrome",
	"address100",
	"archaeological_site",
	"arts_centre",
	"artwork",
	"attraction",
	"bakery",
	"bar",
	"basin",
	"bay",
	"beach",
	"cafe",
	"car_wash",
	"castle",
	"cemetery",
	"cinema",
	"community_centre",
	"cycleway",
	"department_store",
	"farmyard",
	"forest",
	"furniture",
	"garden",
	"garden_centre",
	"golf_course",
	"grave_yard",
	"hospital",
	"hotel",
	"information",
	"library",
	"mall",
	"marina",
	"memorial",
	"military",
	"monument",
	"motel",
	"museum",
	"park",
	"parking",
	"path",
	"pedestrian",
	"pitch",
	"place_of_worship",
	"pub",
	"public_building",
	"roman_road",
	"ruins",
	"school",
	"slipway",
	"sports_centre",
	"stadium",
	"supermarket",
	"theatre",
	"townhall",
	"viewpoint",
	"water",
	"zoo",

	"bus_stop",
	"footway",
	"nature_reserve",
	"wetland",
}
