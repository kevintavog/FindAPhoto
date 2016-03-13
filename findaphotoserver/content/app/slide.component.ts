import { Component, OnInit } from 'angular2/core';
import { Router,RouteParams, ROUTER_DIRECTIVES, Location } from 'angular2/router';

import { SearchRequest } from './search-request';
import { SearchResults,SearchGroup,SearchItem } from './search-results';
import { SearchService } from './search.service';
import { SearchComponent } from './search.component';

import { DateStringToLocaleDatePipe } from './datestring-to-localedate.pipe';

@Component({
  selector: 'slide',
  templateUrl: 'app/slide.component.html',
  styleUrls:  ['app/slide.component.css'],
  directives: [ROUTER_DIRECTIVES],
  pipes: [DateStringToLocaleDatePipe],
  inputs: ['slideId']
})

export class SlideComponent implements OnInit {
  private static QueryProperties: string = "id,slideUrl,imageName,createdDate,keywords,city,thumbUrl,latitude,longitude,locationName,mimeType,mediaType,path,mediaUrl,warnings"

  searchRequest: SearchRequest;
  slideInfo: SearchItem;
  slideIndex: number;
  totalSearchMatches: number;
  searchPage: number;
  error: string;

  constructor(
    private _router: Router,
    private _routeParams: RouteParams,
    private _searchService: SearchService,
    private _location: Location) { }

  ngOnInit() {
    this.slideInfo = undefined
    this.error = undefined

    let slideId = this._routeParams.get('id');

    this.searchRequest = { searchText: "", first: 1, pageCount: 1, properties: SlideComponent.QueryProperties };
    this.searchRequest.searchText = this._routeParams.get('q');
    this.searchRequest.first = +this._routeParams.get('i');
    this.loadSlide()
  }

  previousSlide() {
      if (this.slideIndex > 1) {
        let index = this.searchRequest.first - 1
        this._router.navigate( ['Slide', {id: this.slideInfo.id, q:this.searchRequest.searchText, i:index}] );
    }
  }

  nextSlide() {
      if (this.slideIndex < this.totalSearchMatches) {
          let index = this.searchRequest.first + 1
          this._router.navigate( ['Slide', {id: this.slideInfo.id, q:this.searchRequest.searchText, i:index}] );
      }
  }

  loadSlide() {
    this.slideIndex = this.searchRequest.first
    this.searchPage = (1 + (this.searchRequest.first / SearchComponent.ItemsPerPage)) | 0
    this._searchService.search(this.searchRequest).subscribe(
      results => {
          if (results.groups.length > 0 && results.groups[0].items.length > 0) {
              this.slideInfo = results.groups[0].items[0]
              this.totalSearchMatches = results.totalMatches
          } else {
              this.error = "That slide cannot be found"
          }
      },
      error => this.error = "The server returned an error: " + error
    );
  }
}
