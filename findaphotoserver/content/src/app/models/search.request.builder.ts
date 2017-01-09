import { Injectable } from '@angular/core';
import { Params } from '@angular/router';

import { SearchRequest } from './search-request';
// import { DataDisplayer } from '../providers/data-displayer';
import { DataDisplayer } from '../providers/data-displayer';

@Injectable()
export class SearchRequestBuilder {
    public static monthNames: string[] = [
        'January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December' ];

    constructor(private displayer: DataDisplayer) {}

    toReadableString(searchRequest: SearchRequest) {
        switch (searchRequest.searchType) {
            case 's':
                return 'for \'' + searchRequest.searchText + '\'';
            case 'd':
                let activeDate = new Date(
                    2016,
                    searchRequest.month - 1,
                    searchRequest.day, 0, 0, 0, 0);
                return 'on ' + SearchRequestBuilder.monthNames[activeDate.getMonth()] + ' ' + searchRequest.day;
            case 'l':
                return 'Nearby ' + this.displayer.latitudeDms(searchRequest.latitude) 
                    + ', ' + this.displayer.longitudeDms(searchRequest.longitude);
            default:
                return 'Unknown search type: ' + searchRequest.searchType;
        }
    }

    toLinkParametersObject(searchRequest: SearchRequest) {
        let properties = {};

        if (searchRequest == null) {
            return properties;
        }

        properties['t'] = searchRequest.searchType;
        switch (searchRequest.searchType) {
            case 's':
                properties['q'] = searchRequest.searchText;
                break;
            case 'd':
                properties['m'] = searchRequest.month;
                properties['d'] = searchRequest.day;
                break;
            case 'l':
                properties['lat'] = searchRequest.latitude;
                properties['lon'] = searchRequest.longitude;
                break;
            default:
                console.log('Unknown search type: %s', searchRequest.searchType);
                throw new Error('Unknown search type: ' + searchRequest.searchType);
        }

        if (searchRequest.drilldown && searchRequest.drilldown.length > 0) {
            properties['drilldown'] = searchRequest.drilldown;
        }

        return properties;
    }

    createRequest(params: Params, itemsPerPage: number, queryProperties: string, defaultType: string) {
        let searchType = defaultType;
        if ('t' in params) {
            searchType = params['t'];
        }

        let searchText = params['q'];
        if (!searchText) {
            searchText = '';
        }

        let pageNumber = +params['p'];
        if (!pageNumber || pageNumber < 1) {
            pageNumber = 1;
        }

        let firstItem = 1;
        if ('i' in params) {
            firstItem = +params['i'];
        } else {
            firstItem = 1 + ((pageNumber - 1) * itemsPerPage);
        }

        // Bydate search defaults to today
        let today = new Date();
        let month = today.getMonth() + 1;
        let day = today.getDate();
        if ('m' in params && 'd' in params) {
            month = +params['m'];
            day = +params['d'];
        }

        // Nearby search defaults to ... somewhere?
        let latitude = 0.00;
        let longitude = 0.00;
        if ('lat' in params && 'lon' in params) {
            latitude = +params['lat'];
            longitude = +params['lon'];
        }

        let drilldown = params['drilldown'];

        return {
            searchType: searchType,
            searchText: searchText,
            first: firstItem,
            pageCount: itemsPerPage,
            properties: queryProperties,
            month: month,
            day: day,
            byDayRandom: false,
            latitude: latitude,
            longitude: longitude,
            maxKilometers: 0,
            drilldown: drilldown };
    }
}
