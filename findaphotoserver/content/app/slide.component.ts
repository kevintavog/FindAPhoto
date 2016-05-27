import { Component, OnInit } from '@angular/core';
import { Router,RouteParams, ROUTER_DIRECTIVES } from '@angular/router-deprecated';
import { Location } from '@angular/common';

import { BaseComponent } from './base.component';
import { SearchRequest } from './search-request';
import { SearchResults, SearchGroup, SearchItem } from './search-results';
import { SearchService } from './search.service';
import { BaseSearchComponent } from './base.search.component';
import { SearchRequestBuilder } from './search.request.builder';

import { DateStringToLocaleDatePipe } from './datestring-to-localedate.pipe';

@Component({
  selector: 'slide',
  templateUrl: 'app/slide.component.html',
  styleUrls:  ['app/slide.component.css'],
  directives: [ROUTER_DIRECTIVES],
  pipes: [DateStringToLocaleDatePipe],
  inputs: ['slideId']
})

export class SlideComponent extends BaseComponent implements OnInit {
  private static QueryProperties: string = "id,slideUrl,imageName,createdDate,keywords,city,thumbUrl,latitude,longitude,locationName,mimeType,mediaType,path,mediaUrl,warnings"
  private static NearbyProperties: string = "id,thumbUrl,latitude,longitude,distancekm"
  private static SameDateProperties: string = "id,thumbUrl,createdDate,city"

  searchRequest: SearchRequest;
  slideInfo: SearchItem;
  slideIndex: number;
  totalSearchMatches: number;
  searchPage: number;
  error: string;
  nearbyResults: SearchItem[];
  nearbyError: string;
  sameDateResults: SearchItem[];
  sameDateError: string;


  constructor(
    private _router: Router,
    private _routeParams: RouteParams,
    private _searchService: SearchService,
    private _location: Location,
    private _searchRequestBuilder: SearchRequestBuilder) { super() }

  ngOnInit() {
    this.slideInfo = undefined
    this.error = undefined

    let slideId = this._routeParams.get('id');
    this.searchRequest = this._searchRequestBuilder.createRequest(this._routeParams, 1, SlideComponent.QueryProperties, 's')
    this.loadSlide()
  }

  hasLocation() {
      return this.slideInfo.longitude != undefined && this.slideInfo.latitude != undefined
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
    this.searchPage = (1 + (this.searchRequest.first / BaseSearchComponent.ItemsPerPage)) | 0
    this._searchService.search(this.searchRequest).subscribe(
      results => {
          if (results.groups.length > 0 && results.groups[0].items.length > 0) {
              this.slideInfo = results.groups[0].items[0]
              this.totalSearchMatches = results.totalMatches

              this.loadNearby()
              this.loadSameDate()
          } else {
              this.error = "The slide cannot be found"
          }
      },
      error => this.error = "The server returned an error: " + error
    );
  }

  loadNearby() {
      if (!this.hasLocation()) { return }
      this._searchService.searchByLocation(this.slideInfo.latitude, this.slideInfo.longitude, SlideComponent.NearbyProperties, 1, 7).subscribe(
          results => {
              if (results.groups.length > 0 && results.groups[0].items.length > 0) {
                  let items = Array<SearchItem>()
                  let list = results.groups[0].items
                  for (let index = 0; index < list.length && items.length < 5; ++index) {
                      let si = list[index]
                      if (si.id != this.slideInfo.id) {
                          items.push(si)
                      }
                  }

                  this.nearbyResults = items
              } else {
                  this.nearbyError = "No nearby results"
              }
          },
          error => this.nearbyError = "The server returned an error: " + error
      )
  }

  loadSameDate() {
      let month = this.itemMonth(this.slideInfo)
      let day = this.itemDay(this.slideInfo)
      if (month < 0 || day < 0) { return }

      this._searchService.searchByDay(month, day, SlideComponent.SameDateProperties, 1, 7, true).subscribe(
          results => {
              if (results.groups.length > 0 && results.groups[0].items.length > 0) {
                  let items = Array<SearchItem>()
                  let list = results.groups[0].items
                  for (let index = 0; index < list.length && items.length < 5; ++index) {
                      let si = list[index]
                      if (si.id != this.slideInfo.id) {
                          items.push(si)
                      }
                  }

                  this.sameDateResults = items
              } else {
                  this.sameDateError = "No results with the same date"
              }
          },
          error => this.sameDateError = "The server returned an error: " + error
      )

  }

}
