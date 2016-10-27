import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, NavigationExtras, Router } from '@angular/router';
import { Location } from '@angular/common';

import { SearchRequestBuilder } from '../../models/search.request.builder';
import { SearchItem } from '../../models/search-results';

import { DataDisplayer } from '../../providers/data-displayer';
import { SearchResultsProvider } from '../../providers/search-results.provider';
import { SearchService } from '../../services/search.service';


@Component({
    selector: 'app-single-item',
    templateUrl: './single-item.component.html',
    styleUrls: ['./single-item.component.css']
})

export class SingleItemComponent implements OnInit {
    private static QueryProperties: string = 'id,slideUrl,imageName,createdDate,keywords,city,thumbUrl' +
        ',latitude,longitude,locationName,mimeType,mediaType,path,mediaUrl,warnings';
    private static NearbyProperties: string = 'id,thumbUrl,latitude,longitude,distancekm';
    private static SameDateProperties: string = 'id,thumbUrl,createdDate,city';

    itemInfo: SearchItem;
    itemIndex: number;
    searchPage: number;
    nearbyResults: SearchItem[];
    nearbyError: string;
    sameDateResults: SearchItem[];
    sameDateError: string;
    nearbySearchResultsProvider: SearchResultsProvider;
    bydaySearchResultsProvider: SearchResultsProvider;

    get totalSearchMatches() {
        if (!this._searchResultsProvider.searchResults) {
            return 0;
        }
        return this._searchResultsProvider.searchResults.totalMatches;
    }

    constructor(
        protected _router: Router,
        protected _route: ActivatedRoute,
        protected _location: Location,
        protected _searchRequestBuilder: SearchRequestBuilder,
        protected _searchResultsProvider: SearchResultsProvider,
        protected searchService: SearchService,
        private displayer: DataDisplayer) {
            _searchResultsProvider.searchStartingCallback = (context) => {};
            _searchResultsProvider.searchCompletedCallback = (context) => this.loadItemCompleted();

            this.nearbySearchResultsProvider = new SearchResultsProvider(searchService, _route, _searchRequestBuilder);
            this.nearbySearchResultsProvider.searchStartingCallback = (context) => {};
            this.nearbySearchResultsProvider.searchCompletedCallback = (context) => this.nearbySearchCompleted();

            this.bydaySearchResultsProvider = new SearchResultsProvider(searchService, _route, _searchRequestBuilder);
            this.bydaySearchResultsProvider.searchStartingCallback = (context) => {};
            this.bydaySearchResultsProvider.searchCompletedCallback = (context) => this.bydaySearchCompleted();
    }

    ngOnInit() {
        this.itemInfo = undefined;
        this.nearbySearchResultsProvider.initializeRequest('', 'l');
        this.bydaySearchResultsProvider.initializeRequest('', 'd');

        this._searchResultsProvider.initializeRequest(SingleItemComponent.QueryProperties, 's');

        this._route.queryParams.subscribe(params => {
            this.loadItem();
         });
    }

    hasLocation() {
        return this.itemInfo.longitude !== undefined && this.itemInfo.latitude !== undefined;
    }

    firstItem() {
        if (this.itemIndex > 1) {
            let index = 1;
            this._router.navigate( ['singleitem'], this.getNavigationExtras({ id: this.itemInfo.id, i: index }));
        }
    }

    previousItem() {
        if (this.itemIndex > 1) {
          let index = this._searchResultsProvider.searchRequest.first - 1;
          this._router.navigate( ['singleitem'], this.getNavigationExtras({ id: this.itemInfo.id, i: index }));
      }
    }

    lastItem() {
        if (this.itemIndex < this._searchResultsProvider.searchResults.totalMatches) {
            let index = this._searchResultsProvider.searchResults.totalMatches;
            this._router.navigate( ['singleitem'], this.getNavigationExtras({ id: this.itemInfo.id, i: index }));
        }
    }

    nextItem() {
        if (this.itemIndex < this._searchResultsProvider.searchResults.totalMatches) {
            let index = this._searchResultsProvider.searchRequest.first + 1;
            this._router.navigate( ['singleitem'], this.getNavigationExtras({ id: this.itemInfo.id, i: index }));
        }
    }

    loadItem() {
        this.nearbyResults = undefined;
        this.sameDateResults = undefined;
        this.itemIndex = this._searchResultsProvider.searchRequest.first;
        this.searchPage = Math.round(1 + (this._searchResultsProvider.searchRequest.first / SearchResultsProvider.ItemsPerPage));
        this._searchResultsProvider.search(null);
    }

    loadItemCompleted() {
        if (this._searchResultsProvider.searchResults.groups.length > 0
            && this._searchResultsProvider.searchResults.groups[0].items.length > 0) {
            this.itemInfo = this._searchResultsProvider.searchResults.groups[0].items[0];

            this.loadNearby();
            this.loadSameDate();
        } else {
            this._searchResultsProvider.serverError = 'The item cannot be found';
        }
    }

    loadNearby() {
        if (!this.hasLocation()) { return; }

        this.nearbySearchResultsProvider.searchRequest.searchType = 'l';
        this.nearbySearchResultsProvider.searchRequest.latitude = this.itemInfo.latitude;
        this.nearbySearchResultsProvider.searchRequest.longitude = this.itemInfo.longitude;
        this.nearbySearchResultsProvider.searchRequest.properties = SingleItemComponent.NearbyProperties;
        this.nearbySearchResultsProvider.searchRequest.first = 1;
        this.nearbySearchResultsProvider.searchRequest.pageCount = 7;

        this.nearbySearchResultsProvider.search(null);
    }

    nearbySearchCompleted() {
        let results = this.nearbySearchResultsProvider.searchResults;
        if (results && results.groups.length > 0 && results.groups[0].items.length > 0) {
            let items = Array<SearchItem>();
            let list = results.groups[0].items;
            for (let index = 0; index < list.length && items.length < 5; ++index) {
                let si = list[index];
                if (si.id !== this.itemInfo.id) {
                    items.push(si);
                }
            }

            this.nearbyResults = items;
        } else {
            this.nearbySearchResultsProvider.serverError = 'No nearby results';
        }
    }

    loadSameDate() {
        let month = this.displayer.itemMonth(this.itemInfo);
        let day = this.displayer.itemDay(this.itemInfo);
        if (month < 0 || day < 0) { return; }


        this.bydaySearchResultsProvider.searchRequest.searchType = 'd';
        this.bydaySearchResultsProvider.searchRequest.byDayRandom = true;
        this.bydaySearchResultsProvider.searchRequest.month = month;
        this.bydaySearchResultsProvider.searchRequest.day = day;
        this.bydaySearchResultsProvider.searchRequest.properties = SingleItemComponent.SameDateProperties;
        this.bydaySearchResultsProvider.searchRequest.first = 1;
        this.bydaySearchResultsProvider.searchRequest.pageCount = 7;

        this.bydaySearchResultsProvider.search(null);
    }

    bydaySearchCompleted() {
        let results = this.bydaySearchResultsProvider.searchResults;
        if (results && results.groups.length > 0 && results.groups[0].items.length > 0) {
            let items = Array<SearchItem>();
            let list = results.groups[0].items;
            for (let index = 0; index < list.length && items.length < 5; ++index) {
                let si = list[index];
                if (si.id !== this.itemInfo.id) {
                    items.push(si);
                }
            }

            this.sameDateResults = items;
        } else {
            this.nearbySearchResultsProvider.serverError = 'No results with the same date';
        }
    }

    home() {
        this._router.navigate( ['search'], this.getNavigationExtras() );
    }

    searchToday() {
        this._router.navigate( ['byday'] );
    }

    searchNearby() {
        this._router.navigate( ['bylocation'] );
    }

    getNavigationExtras(extraParams?: Object) {
        let params = this._searchRequestBuilder.toLinkParametersObject(this._searchResultsProvider.searchRequest);
        if (extraParams !== undefined) {
            Object.assign(params, extraParams);
        }
        let navigationExtras: NavigationExtras = { queryParams: params };
        return navigationExtras;
    }
}
