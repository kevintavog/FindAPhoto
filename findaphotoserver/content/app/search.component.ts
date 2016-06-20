import { Component, OnInit } from '@angular/core';
import { Router, ROUTER_DIRECTIVES, RouteParams } from '@angular/router-deprecated';
import { Location } from '@angular/common';

import { BaseSearchComponent } from './base.search.component';
import { SearchRequest } from './search-request';
import { SearchResults,SearchGroup,SearchItem } from './search-results';
import { SearchService } from './search.service';
import { SearchRequestBuilder } from './search.request.builder';
import { CategoryTreeView } from './category-tree-view.component';

import { DateStringToLocaleDatePipe } from './datestring-to-localedate.pipe';

@Component({
  selector: 'search',
  templateUrl: 'app/search.component.html',
  styleUrls:  ['app/search.component.css'],
  directives: [ROUTER_DIRECTIVES, CategoryTreeView],
  pipes: [DateStringToLocaleDatePipe]
})

export class SearchComponent extends BaseSearchComponent implements OnInit {
    resultsSearchText: string;

    constructor(
        router: Router,
        searchService: SearchService,
        routeParams: RouteParams,
        location: Location,
        searchRequestBuilder: SearchRequestBuilder)
    {
        super("/search", router, routeParams, location, searchService, searchRequestBuilder)
    }

    ngOnInit() {
        this.showSearch = true
        this.initializeSearchRequest('s')

        if ("q" in this._routeParams.params) {
            this.internalSearch(false)
        }
    }

  userSearch() {
      // If the search is new or different, navigate so we can use browser back to get to previous search results
      if (this.resultsSearchText && this.resultsSearchText != this.searchRequest.searchText) {
          this._router.navigate( ['Search', { q: this.searchRequest.searchText, p: 1 }] );
          return
      }

      this.internalSearch(true)
  }

  processSearchResults() {
      this.resultsSearchText = this.searchRequest.searchText
  }
}
