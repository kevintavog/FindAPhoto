import { Component, OnInit } from 'angular2/core';
import { Router, ROUTER_DIRECTIVES, RouteParams, Location } from 'angular2/router';

import { SearchRequest } from './search-request';
import { SearchResults,SearchGroup,SearchItem } from './search-results';
import { SearchService } from './search.service';

import { DateStringToLocaleDatePipe } from './datestring-to-localedate.pipe';

@Component({
  selector: 'byday',
  templateUrl: 'app/search.component.html',
  styleUrls:  ['app/search.component.css'],
  directives: [ROUTER_DIRECTIVES],
  pipes: [DateStringToLocaleDatePipe]
})

export class ByDayComponent implements OnInit {
    private static QueryProperties: string = "id,city,keywords,imageName,createdDate,thumbUrl,slideUrl,warnings"
    public static ItemsPerPage: number = 30
    private static monthNames: string[] = [ "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December" ];

    pageMessage: string
    showSearch: boolean
    serverError: string
    searchRequest: SearchRequest;
    searchResults: SearchResults;
    currentPage: number;
    totalPages: number;
    activeDate: Date;


  constructor(
    private _searchService: SearchService) { }

  ngOnInit() {
      this.showSearch = false
      this.activeDate = new Date()
      this.searchRequest = { searchText: "", first: 1, pageCount: ByDayComponent.ItemsPerPage, properties: ByDayComponent.QueryProperties }
      this.internalSearch()
  }

  internalSearch() {
      this.searchResults = undefined
      this.serverError = undefined
      this.pageMessage = undefined

      this._searchService.today(this.activeDate.getMonth() + 1, this.activeDate.getDate(), ByDayComponent.QueryProperties).subscribe(
          results => {
              this.searchResults = results

              // DOES NOT honor locale. Nor is the month name shown
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
