package resolveplacename

import (
	"github.com/kevintavog/findaphoto/common"

	"github.com/Jeffail/gabs"
)

func generatePlacename(media *common.Media, address *gabs.Container) {
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
	media.LocationPlaceName = placename(address)
}

func placename(address *gabs.Container) string {
	placename := ""
	children, _ := address.ChildrenMap()
	for _, value := range children {
		placename += value.Data().(string) + " "
	}
	return placename
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

// For 'cityname'
var prioritizedCityNameComponents = []string{
	"city",
	"city_district", // Perhaps shortest of this & city?
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
	"archaeological_site",
	"arts_centre",
	"attraction",
	"bakery",
	"bar",
	"basin",
	"building",
	"cafe",
	"car_wash",
	"chemist",
	"cinema",
	"cycleway",
	"department_store",
	"fast_food",
	"furniture",
	"garden",
	"garden_centre",
	"golf_course",
	"grave_yard",
	"hospital",
	"hotel",
	"house",
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
	"restaurant",
	"roman_road",
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

	"footway",
	"nature_reserve",
}
