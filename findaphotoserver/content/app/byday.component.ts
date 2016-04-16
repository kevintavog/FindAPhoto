import { Component, OnInit } from 'angular2/core';
import { Router, ROUTER_DIRECTIVES, RouteParams, Location } from 'angular2/router';

import { BaseSearchComponent } from './base.search.component';
import { SearchRequest } from './search-request';
import { SearchResults,SearchGroup,SearchItem } from './search-results';
import { SearchService } from './search.service';
import { SearchRequestBuilder } from './search.request.builder';

import { DateStringToLocaleDatePipe } from './datestring-to-localedate.pipe';

@Component({
  selector: 'byday',
  templateUrl: 'app/search.component.html',
  styleUrls:  ['app/search.component.css'],
  directives: [ROUTER_DIRECTIVES],
  pipes: [DateStringToLocaleDatePipe]
})

export class ByDayComponent extends BaseSearchComponent implements OnInit {
    private static monthNames: string[] = [ "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December" ];

    pageMessage: string
    activeDate: Date;


    constructor(
        routeParams: RouteParams,
        private _searchService: SearchService,
        searchRequestBuilder: SearchRequestBuilder)
    {
        super(routeParams, searchRequestBuilder)
    }

    ngOnInit() {
        this.showSearch = false
        this.initializeSearchRequest('d')
        this.activeDate = new Date(2016, this.searchRequest.month - 1, this.searchRequest.day, 0, 0, 0, 0)
        this.internalSearch()
    }

    internalSearch() {
      this.searchResults = undefined
      this.serverError = undefined
      this.pageMessage = undefined

      this._searchService.search(this.searchRequest).subscribe(
          results => {
              this.searchResults = results

              // DOES NOT honor locale...
              this.pageMessage = "Your pictures from " + ByDayComponent.monthNames[this.activeDate.getMonth()] + "  " + this.activeDate.getDate()

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
