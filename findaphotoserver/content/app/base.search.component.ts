import { Router, ROUTER_DIRECTIVES, RouteParams, Location } from 'angular2/router';

import { BaseComponent } from './base.component';
import { SearchService } from './search.service';
import { SearchRequestBuilder } from './search.request.builder';
import { SearchRequest } from './search-request';
import { SearchResults,SearchGroup,SearchItem } from './search-results';

export abstract class BaseSearchComponent extends BaseComponent {
    protected static QueryProperties: string = "id,city,keywords,imageName,createdDate,latitude,longitude,thumbUrl,slideUrl,warnings"
    public static ItemsPerPage: number = 30

    showSearch: boolean
    serverError: string
    searchRequest: SearchRequest;
    searchResults: SearchResults;
    currentPage: number;
    totalPages: number;
    pageMessage: string

    constructor(
        private _pageRoute: string,
        protected _routeParams: RouteParams,
        private _location: Location,
        private _searchService: SearchService,
        protected _searchRequestBuilder: SearchRequestBuilder) { super() }


    initializeSearchRequest(searchType: string) {
        this.searchRequest = this._searchRequestBuilder.createRequest(this._routeParams, BaseSearchComponent.ItemsPerPage, BaseSearchComponent.QueryProperties, searchType)
    }


    slideSearchLinkParameters(item: SearchItem, imageIndex, groupIndex: number) {
        let properties = this._searchRequestBuilder.toLinkParametersObject(this.searchRequest)
        properties['id'] = item.id
        properties['i'] = imageIndex + groupIndex + this.searchRequest.first
        return properties
    }

    updateUrl() {
        this._location.go(this._pageRoute, this._searchRequestBuilder.toSearchQueryParameters(this.searchRequest) + "&p=" + this.currentPage)
    }

    previousPage() {
        if (this.currentPage > 1) {
            let zeroBasedPage = this.currentPage - 1
            this.searchRequest.first = 1 + ((zeroBasedPage - 1) * BaseSearchComponent.ItemsPerPage)
            this.internalSearch(true)
        }
    }

    nextPage() {
        if (this.currentPage < this.totalPages) {
            let zeroBasedPage = this.currentPage - 1
            this.searchRequest.first = 1 + ((zeroBasedPage + 1) * BaseSearchComponent.ItemsPerPage)
            this.internalSearch(true)
        }
    }

    internalSearch(updateUrl: boolean) {
        this.searchResults = undefined
        this.serverError = undefined
        this.pageMessage = undefined

        this._searchService.search(this.searchRequest).subscribe(
            results => {
                this.searchResults = results

                let resultIndex = 0
                for (var group of this.searchResults.groups) {
                    group.resultIndex = resultIndex
                    resultIndex += group.items.length
                }

                let pageCount = this.searchResults.totalMatches / BaseSearchComponent.ItemsPerPage
                this.totalPages = ((pageCount) | 0) + (pageCount > Math.floor(pageCount) ? 1 : 0)
                this.currentPage = 1 + (this.searchRequest.first / BaseSearchComponent.ItemsPerPage) | 0

                if (updateUrl) { this.updateUrl() }

                this.processSearchResults()
            },
            error => this.serverError = error
       );
    }


    abstract processSearchResults()
}
