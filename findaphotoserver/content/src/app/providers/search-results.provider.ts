import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { ActivatedRoute } from '@angular/router';

import { SearchService } from '../services/search.service';
import { SearchRequest } from '../models/search-request';
import { SearchCategory, SearchResults } from '../models/search-results';
import { SearchRequestBuilder } from '../models/search.request.builder';


@Injectable()
export class SearchResultsProvider {
    public static QueryProperties: string = 'id,city,keywords,imageName,createdDate,latitude,longitude,thumbUrl,slideUrl,warnings';
    public static ItemsPerPage: number = 50;

    searchRequest: SearchRequest;
    searchResults: SearchResults;
    serverError: string;

    totalPages: number;
    currentPage: number;

    searchStartingCallback: (context: Map<string, any>) => void;
    searchCompletedCallback: (context: Map<string, any>) => void;


    constructor(
        private _searchService: SearchService,
        private _route: ActivatedRoute,
        private _searchRequestBuilder: SearchRequestBuilder) {
    }

    search(context: Map<string, any>) {
        this.searchResults = undefined;
        this.serverError = undefined;
        if (this.searchStartingCallback) {
            this.searchStartingCallback(context);
        }

        this.searchWithRequest().subscribe(
            results => {
                this.searchResults = results;

                let resultIndex = 0;
                for (let group of this.searchResults.groups) {
                    group.resultIndex = resultIndex;
                    resultIndex += group.items.length;
                }

                let pageCount = this.searchResults.totalMatches / SearchResultsProvider.ItemsPerPage;
                this.totalPages = Math.ceil(pageCount);
                this.currentPage = Math.round(1 + (this.searchRequest.first / SearchResultsProvider.ItemsPerPage));

                let dates = this.categoryDate();
                if (dates != null) {
                    for (let detail of dates.details) {
                        if (detail.value.length === 8 && Number.isInteger(Number(detail.value))) {
                            let year = Number(detail.value.substring(0, 4));
                            let month = Number(detail.value.substring(4, 6));
                            let day = Number(detail.value.substring(6, 8));

                            detail.displayValue = new Date(year, month - 1, day).toLocaleDateString();
                        }
                    }
                }

                if (this.searchCompletedCallback) {
                    this.searchCompletedCallback(context);
                }
            },
            error => {
                this.serverError = error;
                if (this.searchCompletedCallback) {
                    this.searchCompletedCallback(context);
                }
            }
       );
    }

    private searchWithRequest() {
        switch (this.searchRequest.searchType) {
            case 's':
                return this._searchService.searchByText(
                    this.searchRequest.searchText,
                    this.searchRequest.properties,
                    this.searchRequest.first,
                    this.searchRequest.pageCount,
                    this.searchRequest.drilldown);
            case 'd':
                return this._searchService.searchByDay(
                    this.searchRequest.month,
                    this.searchRequest.day,
                    this.searchRequest.properties,
                    this.searchRequest.first,
                    this.searchRequest.pageCount,
                    this.searchRequest.byDayRandom,
                    this.searchRequest.drilldown);
            case 'l':
                return this._searchService.searchByLocation(
                    this.searchRequest.latitude,
                    this.searchRequest.longitude,
                    this.searchRequest.maxKilometers,
                    this.searchRequest.properties,
                    this.searchRequest.first,
                    this.searchRequest.pageCount,
                    this.searchRequest.drilldown);
        }

        return Observable.throw('Unknown search type: ' + this.searchRequest.searchType);
    }

    setEmptyRequest() {
        this.serverError = undefined;
        this.searchResults = undefined;

        if (this.searchRequest) {
            this.searchRequest.searchType = 's';
            this.searchRequest.searchText = '';
            this.searchRequest.first = 1;
            this.searchRequest.drilldown = null;
        }
    }

    initializeRequest(queryProps: string, searchType: string) {
        this.serverError = undefined;
        this.searchResults = undefined;

        this._route.queryParams.subscribe(params => {
            this.searchRequest = this._searchRequestBuilder.createRequest(
                params, SearchResultsProvider.ItemsPerPage, queryProps, searchType);
         });
    }

    firstResult() {
        if (this.searchResults !== undefined && this.searchResults.totalMatches > 0) {
            return this.searchResults.groups[0].items[0];
        }
        return undefined;
    }

    categoryDate(): SearchCategory {
        return this.categoryByField('dateYear');
    }

    categoryKeywords(): SearchCategory {
        return this.categoryByField('keywords');
    }

    categoryPlacenames(): SearchCategory {
        return this.categoryByField('countryName');
    }

    categoryByField(field: string): SearchCategory {
        if (this.searchResults.categories != null) {
            for (let category of this.searchResults.categories) {
                if (category.field === field) { return category; }
            }
        }
        return null;
    }
}
