import { Component, OnInit } from 'angular2/core';
import { Router, ROUTER_DIRECTIVES, RouteParams, Location } from 'angular2/router';

import { BaseSearchComponent } from './base.search.component';
import { SearchRequest } from './search-request';
import { SearchResults,SearchGroup,SearchItem } from './search-results';
import { SearchService } from './search.service';
import { SearchRequestBuilder } from './search.request.builder';

import { DateStringToLocaleDatePipe } from './datestring-to-localedate.pipe';

@Component({
  selector: 'bylocation',
  templateUrl: 'app/search.component.html',
  styleUrls:  ['app/search.component.css'],
  directives: [ROUTER_DIRECTIVES],
  pipes: [DateStringToLocaleDatePipe]
})

export class ByLocationComponent extends BaseSearchComponent implements OnInit {

    constructor(
        router: Router,
        routeParams: RouteParams,
        location: Location,
        searchService: SearchService,
        searchRequestBuilder: SearchRequestBuilder)
    {
        super("/byloc", router, routeParams, location, searchService, searchRequestBuilder)
    }

    ngOnInit() {
        this.showLinks = true
        this.showSearch = false
        this.showDistance = true
        this.showGroup = false

        this.extraProperties = "locationName,distancekm"
        this.initializeSearchRequest('l')

        // TODO: If location not specified, use the browser location (if user allows)
        this.internalSearch(false)
    }

    processSearchResults() {
        let firstResult = this.firstResult()
        if (firstResult != undefined && firstResult.locationName != null) {
            if (firstResult.latitude == this.searchRequest.latitude &&
                firstResult.longitude == this.searchRequest.longitude) {
                    this.pageMessage = "Your pictures near " + firstResult.locationName
                    return
                }
        }

        this.pageMessage = "Pictures near " + this.latitudeDms(this.searchRequest.latitude) + ", " + this.longitudeDms(this.searchRequest.longitude)
    }

    searchNearby() {

    }
}
