package resolveplacename

import (
	"database/sql"
	"fmt"
	"path"
	"strings"
	"sync/atomic"

	"github.com/ian-kent/go-log/log"
	"github.com/kevintavog/findaphoto/common"
	_ "github.com/mattn/go-sqlite3"
)

// Returns an empty string and a nil error if no match is found
func placenameFromLocalCache(latitude, longitude float64) (string, error) {

	db, err := sql.Open("sqlite3", databasePath())
	if err != nil {
		return "", err
	}
	defer db.Close()

	statement, err := db.Prepare("SELECT geoLocation,fullPlacename FROM LocationCache WHERE geoLocation LIKE ?")
	if err != nil {
		return "", err
	}
	defer statement.Close()

	latitudeString := fmt.Sprintf("%f", latitude)
	longitudeString := fmt.Sprintf("%f", longitude)
	geoLocationParameter := fmt.Sprintf("%s%%, %s%%", latitudeString, longitudeString)

	// At least temporarily, attempt to get usable info from the cache - do so by decreasing the accuracy
	// of what we're looking up. Unfortunately, using fmt.Sprintf("%f") rounds - with precision specified,
	// preventing lookups from working very well.
	var dbGeoLocation, dbPlacename string
	err = statement.QueryRow(geoLocationParameter).Scan(&dbGeoLocation, &dbPlacename)
	if err != nil {
		geoLocationParameter = generateQueryParameter(latitudeString, longitudeString, 6)
		err = statement.QueryRow(geoLocationParameter).Scan(&dbGeoLocation, &dbPlacename)
		if err != nil {
			geoLocationParameter = generateQueryParameter(latitudeString, longitudeString, 5)
			err = statement.QueryRow(geoLocationParameter).Scan(&dbGeoLocation, &dbPlacename)
			if err != nil {
				geoLocationParameter = generateQueryParameter(latitudeString, longitudeString, 4)
				err = statement.QueryRow(geoLocationParameter).Scan(&dbGeoLocation, &dbPlacename)
				if err != nil {
					geoLocationParameter = generateQueryParameter(latitudeString, longitudeString, 3)
					err = statement.QueryRow(geoLocationParameter).Scan(&dbGeoLocation, &dbPlacename)
					if err != nil {
						log.Warn("Failed cache lookup using %s", geoLocationParameter)
						return "", err
					}
				}
			}
		}
	}

	atomic.AddInt64(&PlacenameLookups, 1)
	return dbPlacename, nil
}

func generateQueryParameter(latitudeString, longitudeString string, digitsToUse int) string {
	dotLatitude := strings.Index(latitudeString, ".")
	dotLongitude := strings.Index(longitudeString, ".")
	latitudeString = latitudeString[:dotLatitude+digitsToUse]
	longitudeString = longitudeString[:dotLongitude+digitsToUse]
	return fmt.Sprintf("%s%%, %s%%", latitudeString, longitudeString)
}

func databasePath() string {
	return path.Join(common.LocationCacheDirectory, "location.cache.db")
}

func getRecords(offset, count int) (map[string]string, error) {

	result := map[string]string{}

	db, err := sql.Open("sqlite3", databasePath())
	if err != nil {
		return result, err
	}
	defer db.Close()

	statement, err := db.Prepare("SELECT geoLocation,fullPlacename FROM LocationCache LIMIT ? OFFSET ?")
	if err != nil {
		return result, err
	}
	defer statement.Close()

	rows, err := statement.Query(count, offset)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var latLon, placename string
		err = rows.Scan(&latLon, &placename)
		if err != nil {
			return result, err
		}

		result[latLon] = placename
	}
	err = rows.Err()
	if err != nil {
		return result, err
	}

	return result, nil
}
