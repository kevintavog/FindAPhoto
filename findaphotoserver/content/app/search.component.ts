import { Component, OnInit } from 'angular2/core';
import { Router, ROUTER_DIRECTIVES, RouteParams, Location } from 'angular2/router';

import { SearchRequest } from './search-request';
import { SearchResults,SearchGroup,SearchItem } from './search-results';
import { SearchService } from './search.service';

import { DateStringToLocaleDatePipe } from './datestring-to-localedate.pipe';

@Component({
  selector: 'search',
  templateUrl: 'app/search.component.html',
  styleUrls:  ['app/search.component.css'],
  directives: [ROUTER_DIRECTIVES],
  pipes: [DateStringToLocaleDatePipe]
})

export class SearchComponent implements OnInit {
    private static QueryProperties: string = "id,city,keywords,imageName,createdDate,thumbUrl,slideUrl,warnings"
    public static ItemsPerPage: number = 30

    showSearch: boolean
    serverError: string
    searchRequest: SearchRequest;
    searchResults: SearchResults;
    resultsSearchText: string;
    currentPage: number;
    totalPages: number;

  constructor(
    private _router: Router,
    private _searchService: SearchService,
    private _routeParams: RouteParams,
    private _location: Location) { }

  ngOnInit() {
    this.showSearch = true
    let searchText = this._routeParams.get("q")
    if (!searchText) {
        searchText = ""
    }

    let pageNumber = +this._routeParams.get("p")
    if (!pageNumber || pageNumber < 1) {
        pageNumber = 1
    }

    let firstItem = 1 + ((pageNumber - 1) * SearchComponent.ItemsPerPage)
    this.searchRequest = { searchText: searchText, first: firstItem, pageCount: SearchComponent.ItemsPerPage, properties: SearchComponent.QueryProperties };

    let autoSearch = ("q" in this._routeParams.params)
    if (autoSearch) {
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
          this.searchRequest.first = 1 + ((zeroBasedPage - 1) * SearchComponent.ItemsPerPage)
          this.internalSearch(true)
      }
  }

  nextPage() {
      if (this.currentPage < this.totalPages) {
          let zeroBasedPage = this.currentPage - 1
          this.searchRequest.first = 1 + ((zeroBasedPage + 1) * SearchComponent.ItemsPerPage)
          this.internalSearch(true)
      }
  }

  updateUrl() {
      this._location.go("/search", "q=" + this.searchRequest.searchText + "&p=" + this.currentPage)
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

              let pageCount = this.searchResults.totalMatches / SearchComponent.ItemsPerPage
              this.totalPages = ((pageCount) | 0) + (pageCount > Math.floor(pageCount) ? 1 : 0)
              this.currentPage = 1 + (this.searchRequest.first / SearchComponent.ItemsPerPage) | 0

              if (updateUrl) { this.updateUrl() }
          },
          error => this.serverError = "The server returned an error: " + error
     );
  }
}
