import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, NavigationExtras, Router } from '@angular/router';
import { Location } from '@angular/common';

import { BaseSearchComponent } from '../base-search/base-search.component';
import { SearchRequestBuilder } from '../models/search.request.builder';
import { ByDayResult } from '../models/search-results';

import { NavigationProvider } from '../providers/navigation.provider';
import { SearchResultsProvider } from '../providers/search-results.provider';

@Component({
  selector: 'app-search-by-day',
  templateUrl: '../search/search.component.html',
  styleUrls: ['../search/search.component.css']
})

export class SearchByDayComponent extends BaseSearchComponent implements OnInit {
    private static monthNames: string[] = [ "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December" ];

    activeDate: Date;


    constructor(
      router: Router,
      route: ActivatedRoute,
      location: Location,
      searchRequestBuilder: SearchRequestBuilder,
      searchResultsProvider: SearchResultsProvider,
      navigationProvider: NavigationProvider) {
          super("/byday", router, route, location, searchRequestBuilder, searchResultsProvider, navigationProvider);
    }

    ngOnInit() {
        this.uiState.showSearch = false
        this.uiState.showResultCount = true
        this.initializeSearchRequest('d')
        this.activeDate = new Date(2016, this._searchResultsProvider.searchRequest.month - 1, this._searchResultsProvider.searchRequest.day, 0, 0, 0, 0)
        this.internalSearch(false)
    }

    processSearchResults() {
        // DOES NOT honor locale...
        this.pageMessage = "Pictures from " + SearchByDayComponent.monthNames[this.activeDate.getMonth()] + "  " + this.activeDate.getDate()

        this.typeLeftButtonClass = "fa fa-arrow-left"
        this.typeLeftButtonText = this.byDayString(this._searchResultsProvider.searchResults.previousAvailableByDay)
        this.typeRightButtonClass = "fa fa-arrow-right"
        this.typeRightButtonText = this.byDayString(this._searchResultsProvider.searchResults.nextAvailableByDay)
    }

    byDayString(byday: ByDayResult) {
        if (byday == undefined)
            return null
        return SearchByDayComponent.monthNames[byday.month - 1] + " " + byday.day
    }

    typeLeftButton() {
        let navigationExtras: NavigationExtras = {
            queryParams: { m:this._searchResultsProvider.searchResults.previousAvailableByDay.month,
                            d:this._searchResultsProvider.searchResults.previousAvailableByDay.day }
        };

        this._router.navigate( ['byday'], navigationExtras);
    }

    typeRightButton() {
        let navigationExtras: NavigationExtras = {
            queryParams: { m:this._searchResultsProvider.searchResults.nextAvailableByDay.month,
                d:this._searchResultsProvider.searchResults.nextAvailableByDay.day }
        };

        this._router.navigate( ['byday'], navigationExtras);
    }
}
