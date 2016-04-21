package resolveplacename

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/Jeffail/gabs"
	"github.com/stretchr/testify/assert"

	"github.com/kevintavog/findaphoto/common"
)

const britishLibraryJson = `{"address":{
	"bus_stop":"British Library","road":"Euston Road","neighbourhood":"Holborn","suburb":"Bloomsbury",
	"city":"London","state_district":"Greater London","state":"England","postcode":"NW1 2QS",
	"country":"United Kingdom","country_code":"gb"}}`
const britishLibraryPlacename = "British Library, London, England, United Kingdom"

const bletchleyParkJson = `{"address":{
	"museum":"Bletchley Park","footway":"Enigma Place","suburb":"West Bletchley","town":"Bletchley",
	"county":"Milton Keynes","state_district":"South East","state":"England","postcode":"MK3 6GW",
	"country":"United Kingdom","country_code":"gb"}}`
const bletchleyParkPlacename = "Bletchley Park, Bletchley, England, United Kingdom"

const peaceArchParkJson = `{"address":{"address100":"Peace Arch Provincial Park","country":"Canada","country_code":"ca"}}`
const peaceArchParkPlacename = "Peace Arch Provincial Park, Canada"

func getAddressJson(t *testing.T, jsonString string) (*gabs.Container, error) {
	json, err := gabs.ParseJSON(bytes.NewBufferString(jsonString).Bytes())
	if err != nil {
		t.Fatalf("Unable to parse json: %s: (%s)", err.Error(), jsonString)
	}

	if !json.Exists("address") {
		if json.Exists("geoError") {
			return nil, errors.New(fmt.Sprintf("error from provider (%s)", jsonString))
		}
		t.Fatalf("'address' mising from json: %s", jsonString)
	}

	return json.Path("address"), nil
}

func logResult(t *testing.T, media *common.Media) {
	t.Logf("site: %s", media.LocationSiteName)
	t.Logf("city: %s", media.LocationCityName)
	t.Logf("state: %s", media.LocationStateName)
	t.Logf("country: %s", media.LocationCountryName)
	t.Logf("placename: %s", media.LocationPlaceName)
}

func TestGeneratePlacename(t *testing.T) {
	media := &common.Media{}
	address, _ := getAddressJson(t, britishLibraryJson)
	generatePlacename(media, address, nil)

	assert.Equal(t, britishLibraryPlacename, media.LocationHierarchicalName)
}

func TestGenerateAllPlacenames(t *testing.T) {
	jsonAndExpected := map[string]string{
		britishLibraryJson: britishLibraryPlacename,
		bletchleyParkJson:  bletchleyParkPlacename,
		peaceArchParkJson:  peaceArchParkPlacename}

	for k, v := range jsonAndExpected {
		media := &common.Media{}
		address, err := getAddressJson(t, k)
		assert.Nil(t, err, "Address error: %v", err)
		generatePlacename(media, address, nil)

		assert.Equal(t, v, media.LocationHierarchicalName)
	}
}

func checkFields(t *testing.T, address *gabs.Container, media *common.Media) {
	children, err := address.ChildrenMap()
	if err != nil {
		t.Fatalf("Unable to get children from address: %s", err.Error())
	}
	for name, value := range children {
		if !isKnownField(name) {
			t.Logf("Unexpected field: %s (%v) %s [%s]", name, value, media.LocationPlaceName, address.String())
		}
	}
}

func isKnownField(name string) bool {
	if listContains(prioritizedStateNameComponents, name) {
		return true
	}
	if listContains(prioritizedStateNameComponents, name) {
		return true
	}
	if listContains(prioritizedCityNameComponents, name) {
		return true
	}
	if listContains(prioritizedSiteComponents, name) {
		return true
	}
	if listContains(expectedAddressComponents, name) {
		return true
	}
	return false
}

func listContains(list []string, value string) bool {
	for _, s := range list {
		if s == value {
			return true
		}
	}

	return false
}

// For 'city' name
var expectedAddressComponents = []string{
	"address26",
	"address27",
	"address29",
	"atm",
	"bank",
	"beauty",
	"bicycle",
	"bicycle_parking",
	"books",
	"bridleway",
	"building",
	"bus_station",
	"car",
	"car_repair",
	"carpet",
	"chemist",
	"clothes",
	"clinic",
	"college",
	"commercial",
	"common",
	"construction",
	"convenience",
	"courthouse",
	"country_code",
	"country",
	"doctors",
	"dry_cleaning",
	"fast_food",
	"ferry_terminal",
	"fire_station",
	"fuel",
	"grass",
	"guest_house",
	"hairdresser",
	"hardware",
	"house",
	"house_number",
	"industrial",
	"jewelry",
	"junction",
	"mobile_phone",
	"newsagent",
	"nightclub",
	"outdoor",
	"pet",
	"pharmacy",
	"picnic_site",
	"post_office",
	"postcode",
	"post_box",
	"quarry",
	"raceway",
	"recreation_ground",
	"residential",
	"restaurant",
	"retail",
	"road",
	"scrub",
	"shop",
	"sports",
	"swimming_pool",
	"taxi",
	"telephone",
	"toilets",
	"toys",
	"track",
	"travel_agency",
	"tree",
	"university",
	"wood",
}
