import { Component, OnInit } from 'angular2/core';
import { Router, ROUTER_DIRECTIVES, RouteParams, Location } from 'angular2/router';

import { BaseSearchComponent } from './base.search.component';
import { SearchRequest } from './search-request';
import { SearchResults,SearchGroup,SearchItem } from './search-results';
import { SearchService } from './search.service';
import { SearchRequestBuilder } from './search.request.builder';

import { DateStringToLocaleDatePipe } from './datestring-to-localedate.pipe';

@Component({
  selector: 'search',
  templateUrl: 'app/search.component.html',
  styleUrls:  ['app/search.component.css'],
  directives: [ROUTER_DIRECTIVES],
  pipes: [DateStringToLocaleDatePipe]
})

export class SearchComponent extends BaseSearchComponent implements OnInit {
    resultsSearchText: string;

    constructor(
        private _router: Router,
        private _searchService: SearchService,
        routeParams: RouteParams,
        private _location: Location,
        searchRequestBuilder: SearchRequestBuilder)
    {
        super(routeParams, searchRequestBuilder)
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

  updateUrl() {
      this._location.go("/search", this._searchRequestBuilder.toSearchQueryParameters(this.searchRequest) + "&p=" + this.currentPage)
  }

  internalSearch(updateUrl: boolean) {
      this.searchResults = undefined
      this.serverError = undefined
      this._searchService.search(this.searchRequest).subscribe(
          results => {
              this.searchResults = results
              this.resultsSearchText = this.searchRequest.searchText

              let resultIndex = 0
              for (var group of this.searchResults.groups) {
                  group.resultIndex = resultIndex
                  resultIndex += group.items.length
              }

              let pageCount = this.searchResults.totalMatches / BaseSearchComponent.ItemsPerPage
              this.totalPages = ((pageCount) | 0) + (pageCount > Math.floor(pageCount) ? 1 : 0)
              this.currentPage = 1 + (this.searchRequest.first / BaseSearchComponent.ItemsPerPage) | 0

              if (updateUrl) { this.updateUrl() }
          },
          error => this.serverError = error
     );
  }
}
