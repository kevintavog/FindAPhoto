import { Component, OnInit } from 'angular2/core';
import { Router, ROUTER_DIRECTIVES } from 'angular2/router';

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
  searchRequest: SearchRequest;
  searchResults: SearchResults;

  constructor(
    private _router: Router,
    private _searchService: SearchService) { }

  ngOnInit() {
    this.searchRequest = { searchText: "", first: 1, pageCount: 20, properties: "id,city,keywords,imageName,createdDate,thumbUrl,slideUrl" };
  }

  search() {
      this.searchResults = undefined;
      this._searchService.search(this.searchRequest).subscribe(
          results => this.searchResults = results,
          error => console.log("Handle error: " + error) );
  }
}
