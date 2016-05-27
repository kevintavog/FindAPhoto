import { RouteParams } from '@angular/router-deprecated';
import { Injectable } from '@angular/core';

import { SearchRequest } from './search-request';

@Injectable()
export class SearchRequestBuilder {

    toSearchQueryParameters(searchRequest: SearchRequest) {
        switch (searchRequest.searchType) {
            case 's':
                return "t=s&q=" + searchRequest.searchText
            case 'd':
                return "t=d&m=" + searchRequest.month + "&d=" + searchRequest.day
            case 'l':
                return "t=l&lat=" + searchRequest.latitude + "&lon=" + searchRequest.longitude
        }
        return ""
    }

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
        }
        return properties
    }

    createRequest(routeParams: RouteParams, itemsPerPage: number, queryProperties: string, defaultType: string) {

        let searchType = defaultType
        if ("t" in routeParams.params) {
            searchType = routeParams.get('t')
        }

        let searchText = routeParams.get("q")
        if (!searchText) {
            searchText = ""
        }

        let pageNumber = +routeParams.get("p")
        if (!pageNumber || pageNumber < 1) {
            pageNumber = 1
        }

        let firstItem = 1
        if ("i" in routeParams.params) {
            firstItem = +routeParams.get('i')
        } else {
            firstItem = 1 + ((pageNumber - 1) * itemsPerPage)
        }

        // Bydate search defaults to today
        let today = new Date()
        let month = today.getMonth() + 1
        let day = today.getDate()
        if ("m" in routeParams.params && "d" in routeParams.params) {
            month = +routeParams.get('m')
            day = +routeParams.get('d')
        }

        // Nearby search defaults to ... somewhere?
        let latitude = 0.00
        let longitude = 0.00
        if ("lat" in routeParams.params && "lon" in routeParams.params) {
            latitude = +routeParams.get('lat')
            longitude = +routeParams.get('lon')
        }

        return { searchType: searchType, searchText: searchText, first: firstItem, pageCount: itemsPerPage, properties: queryProperties, month: month, day: day, latitude: latitude, longitude: longitude }
    }
}
