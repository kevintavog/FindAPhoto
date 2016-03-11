import { Component, OnInit } from 'angular2/core';
import { Router,RouteParams } from 'angular2/router';

import { SearchRequest } from './search-request';
import { SearchResults,SearchGroup,SearchItem } from './search-results';
import { SearchService } from './search.service';

import { DateStringToLocaleDatePipe } from './datestring-to-localedate.pipe';

@Component({
  selector: 'slide',
  templateUrl: 'app/slide.component.html',
  styleUrls:  ['app/slide.component.css'],
  pipes: [DateStringToLocaleDatePipe],
  inputs: ['slideId']
})

export class SlideComponent implements OnInit {
  public static QueryProperties: string = "id,slideUrl,imageName,createdDate,keywords,city,thumbUrl,latitude,longitude,locationName,mimeType,mediaType,path,mediaUrl"

  searchRequest: SearchRequest;
  slideInfo: SearchItem;
  error: string;

  constructor(
    private _router: Router,
    private _routeParams: RouteParams,
    private _searchService: SearchService) { }

  ngOnInit() {
    this.slideInfo = undefined
    this.error = undefined

    let slideId = this._routeParams.get('id');

    this.searchRequest = { searchText: "", first: 1, pageCount: 1, properties: SlideComponent.QueryProperties };
    this.searchRequest.searchText = this._routeParams.get('q');
    this.searchRequest.first = +this._routeParams.get('i');

    console.log("first is " + this.searchRequest.first)
    this._searchService.search(this.searchRequest).subscribe(
        results => {
            if (results.groups.length > 0 && results.groups[0].items.length > 0) {
                this.slideInfo = results.groups[0].items[0]
                console.log("found slide: " + this.slideInfo.imageName)
            } else {
                console.log("No search results returned")
            }
        },
        error => this.error = error
    );
  }
}
