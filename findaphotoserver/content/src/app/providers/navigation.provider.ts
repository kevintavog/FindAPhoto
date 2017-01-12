import { Injectable } from '@angular/core';
import { NavigationExtras, Params, Router } from '@angular/router';

import { SearchRequestBuilder } from '../models/search.request.builder';
import { SearchResultsProvider } from '../providers/search-results.provider';


@Injectable()
export class NavigationProvider {
    updateSearchCallback: () => void;


    constructor(
        protected _searchResultsProvider: SearchResultsProvider,
        protected _searchRequestBuilder: SearchRequestBuilder,
        protected _router: Router
    ) {}

    initialize() {
    }

    hasSameQueryParams(newParams: Params) {
        let startingUrlTree = this._router.parseUrl(this._router.url);
        for (let key in newParams) {
            if (newParams.hasOwnProperty(key)) {
                let val = String(newParams[key]);
                if (val !== startingUrlTree.queryParams[key]) {
                    return false;
                }
            }
        }

        return true;
    }

    info() {
        this._router.navigate( ['info'] );
    }

    fieldValues() {
        let params = this._searchRequestBuilder.toLinkParametersObject(this._searchResultsProvider.searchRequest);
        let navigationExtras: NavigationExtras = { queryParams: params };
        this._router.navigate( ['fieldvalues'], navigationExtras );
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
        // If on another page, but with a location search, retain that search.
        if (this._searchResultsProvider.searchRequest
            && this._searchResultsProvider.searchRequest.searchType === 'l'
            && !this._router.isActive('bylocation', false)) {

            let queryParams = {
                lat: this._searchResultsProvider.searchRequest.latitude,
                lon: this._searchResultsProvider.searchRequest.longitude,
                t: 'l'};

            if (this._searchResultsProvider.searchRequest.drilldown && this._searchResultsProvider.searchRequest.drilldown.length > 0) {
                queryParams['drilldown'] = this._searchResultsProvider.searchRequest.drilldown;
            }

            let navigationExtras: NavigationExtras = { queryParams: queryParams };
            this._router.navigate( ['bylocation'], navigationExtras );
            this._searchResultsProvider.search(null);
        } else {
            this._router.navigate( ['bylocation']);
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
