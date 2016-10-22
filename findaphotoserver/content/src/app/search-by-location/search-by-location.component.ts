import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { Location } from '@angular/common';

import { BaseSearchComponent } from '../base-search/base-search.component';
import { SearchRequestBuilder } from '../models/search.request.builder';
import { ByDayResult } from '../models/search-results';

import { SearchService } from '../services/search.service';

@Component({
  selector: 'app-search-by-location',
  templateUrl: '../search/search.component.html',
  styleUrls: ['../search/search.component.css']
})

export class SearchByLocationComponent extends BaseSearchComponent implements OnInit {

    constructor(
        router: Router,
        route: ActivatedRoute,
        location: Location,
        searchRequestBuilder: SearchRequestBuilder,
        searchService: SearchService) {
            super("/bylocation", router, route, location, searchRequestBuilder, searchService);
        }

    ngOnInit() {
        this.showSearch = false
        this.showDistance = true
        this.showGroup = false

        this.extraProperties = "locationName,locationDisplayName,distancekm"
        this.initializeSearchRequest('l')

        this.internalSearch(false)
    }

    processSearchResults() {
        let firstResult = this.firstResult()
        if (firstResult != undefined && firstResult.locationName != null) {
            if (firstResult.latitude == this.searchRequest.latitude &&
                firstResult.longitude == this.searchRequest.longitude) {
                    this.setLocationName(firstResult.locationName, firstResult.locationDisplayName)
                    return
                }
        }

        // Ask the server for something nearby the given location
        this._searchService.searchByLocation(this.searchRequest.latitude, this.searchRequest.longitude, "distancekm,locationName,locationDisplayName", 1, 1, null).subscribe(
            results => {
                let messageSet = false
                if (results.totalMatches > 0) {
                    let item = results.groups[0].items[0]
                    if (item.distancekm <= 500) {
                        this.setLocationName(item.locationName, item.locationDetailedName)
                        messageSet = true
                    }
                }

                if (!messageSet) {
                    this.setLocationNameFallbacktMessage()
                }
            },
            error => { this.setLocationNameFallbacktMessage() }
        );

    }

    setLocationName(name: string, displayName: string) {
        this.pageMessage = "Pictures near: " + name
        if (displayName != undefined) {
            this.pageSubMessage = displayName
        }
    }

    setLocationNameFallbacktMessage() {
        this.pageMessage = "Pictures near: " + this.latitudeDms(this.searchRequest.latitude) + ", " + this.longitudeDms(this.searchRequest.longitude)
        this.pageSubMessage = undefined
    }
}
