import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, NavigationExtras, Router } from '@angular/router';
import { Location } from '@angular/common';

import { BaseComponent } from '../base/base.component';
import { BaseSearchComponent } from '../base-search/base-search.component';
import { SearchRequest } from '../models/search-request';
import { SearchRequestBuilder } from '../models/search.request.builder';
import { SearchItem } from '../models/search-results';

import { SearchService } from '../services/search.service';


@Component({
    selector: 'app-single-item',
    templateUrl: './single-item.component.html',
    styleUrls: ['./single-item.component.css']
})

export class SingleItemComponent extends BaseComponent implements OnInit {
    private static QueryProperties: string = "id,slideUrl,imageName,createdDate,keywords,city,thumbUrl,latitude,longitude,locationName,mimeType,mediaType,path,mediaUrl,warnings"
    private static NearbyProperties: string = "id,thumbUrl,latitude,longitude,distancekm"
    private static SameDateProperties: string = "id,thumbUrl,createdDate,city"

    searchRequest: SearchRequest;
    itemInfo: SearchItem;
    itemIndex: number;
    totalSearchMatches: number;
    searchPage: number;
    error: string;
    nearbyResults: SearchItem[];
    nearbyError: string;
    sameDateResults: SearchItem[];
    sameDateError: string;


    constructor(
        protected _router: Router,
        protected _route: ActivatedRoute,
        protected _location: Location,
        protected _searchRequestBuilder: SearchRequestBuilder,
        protected _searchService: SearchService) { super() }

    ngOnInit() {
        this.itemInfo = undefined
        this.error = undefined

        this._route.queryParams.subscribe(params => {
            let itemId = params['id'];
            this.searchRequest = this._searchRequestBuilder.createRequest(params, 1, SingleItemComponent.QueryProperties, 's')
            this.loadItem()
         })
    }

    hasLocation() {
        return this.itemInfo.longitude != undefined && this.itemInfo.latitude != undefined
    }

    firstItem() {
        if (this.itemIndex > 1) {
          let index = 1
          this._router.navigate( ['singleitem'], this.getNavigationExtras({ id: this.itemInfo.id, i:index }))
      }
    }

    previousItem() {
        if (this.itemIndex > 1) {
          let index = this.searchRequest.first - 1
          this._router.navigate( ['singleitem'], this.getNavigationExtras({ id: this.itemInfo.id, i:index }))
      }
    }

    lastItem() {
        if (this.itemIndex < this.totalSearchMatches) {
            let index = this.totalSearchMatches
            this._router.navigate( ['singleitem'], this.getNavigationExtras({ id: this.itemInfo.id, i:index }))
        }
    }

    nextItem() {
        if (this.itemIndex < this.totalSearchMatches) {
            let index = this.searchRequest.first + 1
            this._router.navigate( ['singleitem'], this.getNavigationExtras({ id: this.itemInfo.id, i:index }))
        }
    }

    loadItem() {
      this.itemIndex = this.searchRequest.first
      this.searchPage = (1 + (this.searchRequest.first / BaseSearchComponent.ItemsPerPage)) | 0
      this._searchService.search(this.searchRequest).subscribe(
        results => {
            if (results.groups.length > 0 && results.groups[0].items.length > 0) {
                this.itemInfo = results.groups[0].items[0]
                this.totalSearchMatches = results.totalMatches

                this.loadNearby()
                this.loadSameDate()
            } else {
                this.error = "The item cannot be found"
            }
        },
        error => this.error = "The server returned an error: " + error
      );
    }

    loadNearby() {
        if (!this.hasLocation()) { return }
        this._searchService.searchByLocation(this.itemInfo.latitude, this.itemInfo.longitude, SingleItemComponent.NearbyProperties, 1, 7, null).subscribe(
            results => {
                if (results.groups.length > 0 && results.groups[0].items.length > 0) {
                    let items = Array<SearchItem>()
                    let list = results.groups[0].items
                    for (let index = 0; index < list.length && items.length < 5; ++index) {
                        let si = list[index]
                        if (si.id != this.itemInfo.id) {
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
        let month = this.itemMonth(this.itemInfo)
        let day = this.itemDay(this.itemInfo)
        if (month < 0 || day < 0) { return }

        this._searchService.searchByDay(month, day, SingleItemComponent.SameDateProperties, 1, 7, true, null).subscribe(
            results => {
                if (results.groups.length > 0 && results.groups[0].items.length > 0) {
                    let items = Array<SearchItem>()
                    let list = results.groups[0].items
                    for (let index = 0; index < list.length && items.length < 5; ++index) {
                        let si = list[index]
                        if (si.id != this.itemInfo.id) {
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

    home() {
        this._router.navigate( ['search'], this.getNavigationExtras() )
    }

    searchToday() {
        this._router.navigate( ['byday'] )
    }

    searchNearby() {
        this._router.navigate( ['bylocation'] )
    }

    getNavigationExtras(extraParams?: Object) {
        let params = this._searchRequestBuilder.toLinkParametersObject(this.searchRequest);
        if (extraParams != undefined) {
            Object.assign(params, extraParams)
        }
        let navigationExtras: NavigationExtras = { queryParams: params };
        return navigationExtras;
    }
}
