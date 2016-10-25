import { Injectable } from '@angular/core';
import { Params } from '@angular/router';

import { SearchRequest } from './search-request';

@Injectable()
export class SearchRequestBuilder {

    toLinkParametersObject(searchRequest: SearchRequest) {
        let properties = {}
        properties['t'] = searchRequest.searchType
        switch (searchRequest.searchType) {
            case 's':
                properties['q'] = searchRequest.searchText
                break
            case 'd':
                properties['m'] = searchRequest.month
                properties['d'] = searchRequest.day
                break
            case 'l':
                properties['lat'] = searchRequest.latitude
                properties['lon'] = searchRequest.longitude
                break
            default:
                console.log("Unknown search type: %s", searchRequest.searchType)
                throw new Error("Unknown search type: " + searchRequest.searchType)
        }

        return properties
    }

    createRequest(params: Params, itemsPerPage: number, queryProperties: string, defaultType: string) {
        let searchType = defaultType
        if ("t" in params) {
            searchType = params['t']
        }

        let searchText = params["q"]
        if (!searchText) {
            searchText = ""
        }

        let pageNumber = +params["p"]
        if (!pageNumber || pageNumber < 1) {
            pageNumber = 1
        }

        let firstItem = 1
        if ("i" in params) {
            firstItem = +params['i']
        } else {
            firstItem = 1 + ((pageNumber - 1) * itemsPerPage)
        }

        // Bydate search defaults to today
        let today = new Date()
        let month = today.getMonth() + 1
        let day = today.getDate()
        if ("m" in params && "d" in params) {
            month = +params['m']
            day = +params['d']
        }

        // Nearby search defaults to ... somewhere?
        let latitude = 0.00
        let longitude = 0.00
        if ("lat" in params && "lon" in params) {
            latitude = +params['lat']
            longitude = +params['lon']
        }

        let drilldown = params["drilldown"]

        return { searchType: searchType, searchText: searchText, first: firstItem, pageCount: itemsPerPage,
            properties: queryProperties, month: month, day: day, byDayRandom: false, latitude: latitude, longitude: longitude, drilldown: drilldown }
    }
}
