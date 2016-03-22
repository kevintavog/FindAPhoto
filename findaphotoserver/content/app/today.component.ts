import { Component, OnInit } from 'angular2/core';
import { Router, ROUTER_DIRECTIVES, RouteParams, Location } from 'angular2/router';

import { SearchRequest } from './search-request';
import { SearchResults,SearchGroup,SearchItem } from './search-results';
import { SearchService } from './search.service';

import { DateStringToLocaleDatePipe } from './datestring-to-localedate.pipe';

@Component({
  selector: 'today',
  templateUrl: 'app/search.component.html',
  styleUrls:  ['app/search.component.css'],
  directives: [ROUTER_DIRECTIVES],
  pipes: [DateStringToLocaleDatePipe]
})

export class TodayComponent implements OnInit {
    private static QueryProperties: string = "id,city,keywords,imageName,createdDate,thumbUrl,slideUrl,warnings"
    public static ItemsPerPage: number = 30

    pageMessage: string
    showSearch: boolean
    serverError: string
    searchRequest: SearchRequest;
    searchResults: SearchResults;
    currentPage: number;
    totalPages: number;

  constructor(
    private _router: Router,
    private _searchService: SearchService,
    private _routeParams: RouteParams,
    private _location: Location) { }

  ngOnInit() {
      this.showSearch = false
      this.searchRequest = { searchText: "", first: 1, pageCount: TodayComponent.ItemsPerPage, properties: TodayComponent.QueryProperties }
      this.internalSearch()
  }

  internalSearch() {
      this.searchResults = undefined
      this.serverError = undefined
      this.pageMessage = undefined

      var date = new Date()
      this._searchService.today(date.getMonth() + 1, date.getDate(), TodayComponent.QueryProperties).subscribe(
          results => {
              this.searchResults = results
              // DOES NOT honor locale. Nor is the month name shown
              this.pageMessage = "Your pictures from " + (date.getMonth() + 1) + "-" + date.getDate()

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
