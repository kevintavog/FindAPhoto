import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { Location } from '@angular/common';

import { BaseSearchComponent } from './base-search.component';
import { SearchRequestBuilder } from '../../models/search.request.builder';
import { ByDayResult } from '../../models/search-results';

import { DataDisplayer } from '../../providers/data-displayer';
import { NavigationProvider } from '../../providers/navigation.provider';
import { SearchResultsProvider } from '../../providers/search-results.provider';

@Component({
    selector: 'app-search-by-day',
    templateUrl: '../search/search.component.html',
    styleUrls: ['../search/search.component.css']
})

export class SearchByDayComponent extends BaseSearchComponent implements OnInit {
    private static monthNames: string[] = [
        'January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December' ];

    activeDate: Date;


    constructor(
      router: Router,
      route: ActivatedRoute,
      location: Location,
      searchRequestBuilder: SearchRequestBuilder,
      private searchResultsProvider: SearchResultsProvider,
      private navigationProvider: NavigationProvider,
      private displayer: DataDisplayer) {
          super('/byday', router, route, location, searchRequestBuilder, searchResultsProvider, navigationProvider);
    }

    ngOnInit() {
        this.uiState.showSearch = false;
        this.uiState.showResultCount = true;
        this.navigationProvider.initialize();
        this.searchResultsProvider.initializeRequest(SearchResultsProvider.QueryProperties, 'd');

        this.internalSearch(false);
    }

    processSearchResults() {
        this.activeDate = new Date(
            2016,
            this._searchResultsProvider.searchRequest.month - 1,
            this._searchResultsProvider.searchRequest.day, 0, 0, 0, 0);

        // DOES NOT honor locale...
        this.pageMessage = 'Pictures from '
            + SearchByDayComponent.monthNames[this.activeDate.getMonth()]
            + '  ' + this.activeDate.getDate();

        this.typeLeftButtonClass = 'fa fa-arrow-left';
        this.typeLeftButtonText = this.byDayString(this._searchResultsProvider.searchResults.previousAvailableByDay);
        this.typeRightButtonClass = 'fa fa-arrow-right';
        this.typeRightButtonText = this.byDayString(this._searchResultsProvider.searchResults.nextAvailableByDay);
    }

    byDayString(byday: ByDayResult) {
        if (byday === undefined) {
            return null;
        }
        return SearchByDayComponent.monthNames[byday.month - 1] + ' ' + byday.day;
    }

    typeLeftButton() {
        this.navigationProvider.searchByDay(
                this._searchResultsProvider.searchResults.previousAvailableByDay.month,
                this._searchResultsProvider.searchResults.previousAvailableByDay.day);
    }

    typeRightButton() {
        this.navigationProvider.searchByDay(
                this._searchResultsProvider.searchResults.nextAvailableByDay.month,
                this._searchResultsProvider.searchResults.nextAvailableByDay.day);
    }
}
