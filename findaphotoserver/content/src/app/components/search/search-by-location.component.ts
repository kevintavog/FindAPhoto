import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { Location } from '@angular/common';
import { Title } from '@angular/platform-browser';

import { BaseSearchComponent } from './base-search.component';
import { SearchRequestBuilder } from '../../models/search.request.builder';

import { DataDisplayer } from '../../providers/data-displayer';
import { FieldsProvider } from '../../providers/fields.provider';
import { NavigationProvider } from '../../providers/navigation.provider';
import { SearchResultsProvider } from '../../providers/search-results.provider';

@Component({
  selector: 'app-search-by-location',
  templateUrl: './search.component.html',
  styleUrls: ['./search.component.css']
})

export class SearchByLocationComponent extends BaseSearchComponent implements OnInit {

    constructor(
        router: Router,
        route: ActivatedRoute,
        location: Location,
        searchRequestBuilder: SearchRequestBuilder,
        searchResultsProvider: SearchResultsProvider,
        navigationProvider: NavigationProvider,
        private displayer: DataDisplayer,
        private fieldsProvider: FieldsProvider,
        private titleService: Title) {
            super('/bylocation', router, route, location, searchRequestBuilder, searchResultsProvider, navigationProvider, fieldsProvider);
        }

    ngOnInit() {
        this.titleService.setTitle('Nearby - FindAPhoto');
        this.uiState.showSearch = false;
        this.uiState.showDistance = true;
        this.uiState.showGroup = false;

        let queryProps = SearchResultsProvider.QueryProperties += ',locationName,locationDisplayName,distancekm';
        this._navigationProvider.initialize();
        this.fieldsProvider.initialize();
        this._searchResultsProvider.initializeRequest(queryProps, 'l');

        this.internalSearch(false);
    }

    processSearchResults() {
        let firstResult = this._searchResultsProvider.firstResult();
        if (firstResult !== undefined && firstResult.locationName != null) {
            if (firstResult.latitude === this._searchResultsProvider.searchRequest.latitude &&
                firstResult.longitude === this._searchResultsProvider.searchRequest.longitude) {
                    this.setLocationName(firstResult.locationName, firstResult.locationDisplayName);
                    return;
                }
        }

        this.setLocationNameFallbacktMessage();
        // Ask the server for something nearby the given location
        // this._searchService.searchByLocation(this.searchRequest.latitude, this.searchRequest.longitude, 
        //      "distancekm,locationName,locationDisplayName", 1, 1, null).subscribe(
        //     results => {
        //         let messageSet = false
        //         if (results.totalMatches > 0) {
        //             let item = results.groups[0].items[0]
        //             if (item.distancekm <= 500) {
        //                 this.setLocationName(item.locationName, item.locationDetailedName)
        //                 messageSet = true
        //             }
        //         }
        //
        //         if (!messageSet) {
        //             this.setLocationNameFallbacktMessage()
        //         }
        //     },
        //     error => { this.setLocationNameFallbacktMessage() }
        // );

    }

    setLocationName(name: string, displayName: string) {
        this.pageMessage = 'Pictures near: ' + name;
        if (displayName !== undefined) {
            this.pageSubMessage = displayName;
        }
    }

    setLocationNameFallbacktMessage() {
        this.pageMessage = 'Pictures near: ' + this.displayer.latitudeDms(
            this._searchResultsProvider.searchRequest.latitude)
            + ', '
            + this.displayer.longitudeDms(this._searchResultsProvider.searchRequest.longitude);
        this.pageSubMessage = undefined;
    }
}
