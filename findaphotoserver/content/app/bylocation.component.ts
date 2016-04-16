import { Component, OnInit } from 'angular2/core';
import { Router, ROUTER_DIRECTIVES, RouteParams, Location } from 'angular2/router';

import { BaseSearchComponent } from './base.search.component';
import { SearchRequest } from './search-request';
import { SearchResults,SearchGroup,SearchItem } from './search-results';
import { SearchService } from './search.service';
import { SearchRequestBuilder } from './search.request.builder';

import { DateStringToLocaleDatePipe } from './datestring-to-localedate.pipe';

@Component({
  selector: 'bylocation',
  templateUrl: 'app/search.component.html',
  styleUrls:  ['app/search.component.css'],
  directives: [ROUTER_DIRECTIVES],
  pipes: [DateStringToLocaleDatePipe]
})

export class ByLocationComponent extends BaseSearchComponent implements OnInit {
    pageMessage: string


    constructor(
        routeParams: RouteParams,
        private _searchService: SearchService,
        searchRequestBuilder: SearchRequestBuilder)
    {
        super(routeParams, searchRequestBuilder)
    }

    ngOnInit() {
        this.showSearch = false
        this.initializeSearchRequest('l')

        // If location not specified, use the browser location (if user allows)
        this.internalSearch()
    }

    internalSearch() {
      this.searchResults = undefined
      this.serverError = undefined
      this.pageMessage = undefined

      this._searchService.search(this.searchRequest).subscribe(
          results => {
              this.searchResults = results

console.log("by location search has a total result count of: " + this.searchResults.totalMatches + " - " + this.searchResults.resultCount)

              let resultIndex = 0
              for (var group of this.searchResults.groups) {
                  group.resultIndex = resultIndex
                  resultIndex += group.items.length
              }
          },
          error => this.serverError = "The server returned an error: " + error
     );
  }
}
