import { Router, ROUTER_DIRECTIVES, RouteParams, Location } from 'angular2/router';

import { SearchRequestBuilder } from './search.request.builder';
import { SearchRequest } from './search-request';
import { SearchResults,SearchGroup,SearchItem } from './search-results';

export abstract class BaseSearchComponent {
    protected static QueryProperties: string = "id,city,keywords,imageName,createdDate,thumbUrl,slideUrl,warnings"
    public static ItemsPerPage: number = 30

    showSearch: boolean
    serverError: string
    searchRequest: SearchRequest;
    searchResults: SearchResults;
    currentPage: number;
    totalPages: number;

    constructor(
        protected _routeParams: RouteParams,
        protected _searchRequestBuilder: SearchRequestBuilder) { }


    initializeSearchRequest(searchType: string) {
        this.searchRequest = this._searchRequestBuilder.createRequest(this._routeParams, BaseSearchComponent.ItemsPerPage, BaseSearchComponent.QueryProperties, searchType)
    }


    slideSearchLinkParameters(item: SearchItem, imageIndex, groupIndex: number) {
        let properties = this._searchRequestBuilder.toLinkParametersObject(this.searchRequest)
        properties['id'] = item.id
        properties['i'] = imageIndex + groupIndex + this.searchRequest.first
        return properties
    }

}
