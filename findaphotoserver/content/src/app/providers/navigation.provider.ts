import { Injectable } from '@angular/core';
import { NavigationExtras, Router } from '@angular/router';

import { SearchRequestBuilder } from '../models/search.request.builder';
import { SearchResultsProvider } from '../providers/search-results.provider';


@Injectable()
export class NavigationProvider {
    locationError: string;
    updateSearchCallback: () => void;


    constructor(
        protected _searchResultsProvider: SearchResultsProvider,
        protected _searchRequestBuilder: SearchRequestBuilder,
        protected _router: Router
    ) {}

    initialize() {
        this.locationError = undefined;
    }

    info() {
        this._router.navigate( ['info'] );
    }

    home() {
        // If not on the home page and there's a text search, retain the text search
        if (this._searchResultsProvider.searchRequest 
            && this._searchResultsProvider.searchRequest.searchType === 's'
            && !this._router.isActive('search', false)) {

            let queryParams = { q: this._searchResultsProvider.searchRequest.searchText };
            if (this._searchResultsProvider.searchRequest.drilldown && this._searchResultsProvider.searchRequest.drilldown.length > 0) {
                queryParams['drilldown'] = this._searchResultsProvider.searchRequest.drilldown;
            }

            let navigationExtras: NavigationExtras = { queryParams: queryParams };
            this._router.navigate( ['search'], navigationExtras );
            this._searchResultsProvider.search(null);
        } else {
            this._searchResultsProvider.setEmptyRequest();
            let navigationExtras: NavigationExtras = { queryParams: { } };
            this._router.navigate( ['search'], navigationExtras );
        }
    }

    searchToday() {
        let navigationExtras: NavigationExtras = { queryParams: { } };
        this._router.navigate( ['byday'], navigationExtras );

        // If on the byday page, refresh the search (to clear a different date or page)
        if (this._router.isActive('byday', false)) {
            let today = new Date();
            this._searchResultsProvider.searchRequest.month = today.getMonth() + 1;
            this._searchResultsProvider.searchRequest.day = today.getDate();
            this._searchResultsProvider.searchRequest.first = 1;
            this._searchResultsProvider.search(null);
        }
    }

    searchByDay(month: number, day: number) {
        this._searchResultsProvider.searchRequest.day = day;
        this._searchResultsProvider.searchRequest.month = month;
        this._searchResultsProvider.searchRequest.first = 1;
        this._searchResultsProvider.searchRequest.searchType = 'd';

        let navigationExtras: NavigationExtras = { queryParams: { m: month, d: day } };
        this._router.navigate(['byday'], navigationExtras);
        this.updateSearchCallback();
    }

    searchNearby() {
        this.locationError = undefined;

        if (window.navigator.geolocation) {
            let timer = setTimeout( () => this.locationError = 'Unable to get location: timeout', 5000);

            window.navigator.geolocation.getCurrentPosition(
                (position: Position) => {

                    let navigationExtras: NavigationExtras = {
                        queryParams: { lat: position.coords.latitude, lon: position.coords.longitude }
                    };

                    this._router.navigate( ['bylocation'], navigationExtras);
                },
                (error: PositionError) => {
                    this.locationError = 'Unable to get location: ' + error.message + ' (' + error.code + ')';
                },
                { timeout: 5000 });

            clearTimeout(timer);
        } else {
            this.locationError = 'Unable to get window.navigator.geolocation';
        }
    }

    searchMap() {
        let params = this._searchRequestBuilder.toLinkParametersObject(this._searchResultsProvider.searchRequest);
        let navigationExtras: NavigationExtras = { queryParams: params };
        this._router.navigate( ['map'], navigationExtras );
    }

    gotoPage(pageOneBased: number) {
        if (this._searchResultsProvider.currentPage !== pageOneBased) {
            this._searchResultsProvider.searchRequest.first = 1 + (pageOneBased - 1) * SearchResultsProvider.ItemsPerPage;
            this.updateSearchCallback();
        }
    }

    firstPage() {
        if (this._searchResultsProvider.currentPage > 1) {
            this._searchResultsProvider.searchRequest.first = 1;
            this.updateSearchCallback();
        }
    }

    lastPage() {
        if (this._searchResultsProvider.currentPage < this._searchResultsProvider.totalPages) {
            this._searchResultsProvider.searchRequest.first =
                (this._searchResultsProvider.totalPages - 1) * SearchResultsProvider.ItemsPerPage;
            this.updateSearchCallback();
        }
    }

    previousPage() {
        if (this._searchResultsProvider.currentPage > 1) {
            let zeroBasedPage = this._searchResultsProvider.currentPage - 1;
            this._searchResultsProvider.searchRequest.first = 1 + ((zeroBasedPage - 1) * SearchResultsProvider.ItemsPerPage);
            this.updateSearchCallback();
        }
    }

    nextPage() {
        if (this._searchResultsProvider.currentPage < this._searchResultsProvider.totalPages) {
            let zeroBasedPage = this._searchResultsProvider.currentPage - 1;
            this._searchResultsProvider.searchRequest.first = 1 + ((zeroBasedPage + 1) * SearchResultsProvider.ItemsPerPage);
            this.updateSearchCallback();
        }
    }
}
