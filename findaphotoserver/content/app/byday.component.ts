import { Component, OnInit } from '@angular/core';
import { Router, ROUTER_DIRECTIVES, RouteParams } from '@angular/router-deprecated';
import { Location } from '@angular/common';

import { BaseSearchComponent } from './base.search.component';
import { SearchRequest } from './search-request';
import { SearchResults,SearchGroup,SearchItem,ByDayResult } from './search-results';
import { SearchService } from './search.service';
import { SearchRequestBuilder } from './search.request.builder';

import { DateStringToLocaleDatePipe } from './datestring-to-localedate.pipe';

@Component({
  selector: 'byday',
  templateUrl: 'app/search.component.html',
  styleUrls:  ['app/search.component.css'],
  directives: [ROUTER_DIRECTIVES],
  pipes: [DateStringToLocaleDatePipe]
})

export class ByDayComponent extends BaseSearchComponent implements OnInit {
    private static monthNames: string[] = [ "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December" ];

    activeDate: Date;


    constructor(
        router: Router,
        routeParams: RouteParams,
        location: Location,
        searchService: SearchService,
        searchRequestBuilder: SearchRequestBuilder)
    {
        super("/byday", router, routeParams, location, searchService, searchRequestBuilder)
    }

    ngOnInit() {
        this.showLinks = true
        this.showSearch = false
        this.initializeSearchRequest('d')
        this.activeDate = new Date(2016, this.searchRequest.month - 1, this.searchRequest.day, 0, 0, 0, 0)
        this.internalSearch(false)
    }

    processSearchResults() {
        // DOES NOT honor locale...
        this.pageMessage = "Pictures from " + ByDayComponent.monthNames[this.activeDate.getMonth()] + "  " + this.activeDate.getDate()

        this.typeLeftButtonText = this.byDayString(this.searchResults.previousAvailableByDay)
        this.typeRightButtonText = this.byDayString(this.searchResults.nextAvailableByDay)
    }

    byDayString(byday: ByDayResult) {
        if (byday == undefined)
            return null
        return ByDayComponent.monthNames[byday.month - 1] + " " + byday.day
    }

    typeLeftButton() {
        this._router.navigate( ['ByDay', {m:this.searchResults.previousAvailableByDay.month, d:this.searchResults.previousAvailableByDay.day}] );
    }

    typeRightButton() {
        this._router.navigate( ['ByDay', {m:this.searchResults.nextAvailableByDay.month, d:this.searchResults.nextAvailableByDay.day}] );
    }
}
