import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, NavigationExtras, Router } from '@angular/router';
import { Location } from '@angular/common';
import { Title } from '@angular/platform-browser';

import { Icon, LatLngTuple, Map, Marker } from 'leaflet';

import { SearchRequestBuilder } from '../../models/search.request.builder';
import { SearchItem } from '../../models/search-results';

import { DataDisplayer } from '../../providers/data-displayer';
import { SearchResultsProvider } from '../../providers/search-results.provider';
import { SearchService } from '../../services/search.service';

class SourceNameValue {
    readonly name: string;
    readonly value: string;

    constructor(name: string, value: string) {
        this.name = name;
        this.value = value;
    }
}

@Component({
    selector: 'app-single-item',
    templateUrl: './single-item.component.html',
    styleUrls: ['./single-item.component.css']
})

export class SingleItemComponent implements OnInit {
    private static QueryProperties: string = 'id,slideUrl,imageName,createdDate,keywords,city,thumbUrl' +
        ',latitude,longitude,locationDisplayName,locationName,mimeType,mediaType,path,mediaUrl,tags,warnings';
    private static NearbyProperties: string = 'id,thumbUrl,latitude,longitude,distancekm';
    private static SameDateProperties: string = 'id,thumbUrl,createdDate,city';
    private static CameraProperties: string = 'cameramake,cameramodel,lensinfo,lensmodel';
    private static ImageProperties: string = 'aperture,durationseconds,exposeureprogram,exposuretimestring,flash,fnumber,focallength' +
        ',height,iso,width';

    showMediaSource: boolean = false;
    mediaSource: SourceNameValue[];

    itemInfo: SearchItem;
    itemIndex: number;
    searchPage: number;
    nearbyResults: SearchItem[];
    nearbyError: string;
    sameDateResults: SearchItem[];
    sameDateError: string;
    nearbySearchResultsProvider: SearchResultsProvider;
    bydaySearchResultsProvider: SearchResultsProvider;

    map: Map;
    marker: Marker;
    markerIcon: Icon;


    get totalSearchMatches() {
        if (!this._searchResultsProvider.searchResults) {
            return 0;
        }
        return this._searchResultsProvider.searchResults.totalMatches;
    }

    constructor(
        protected _router: Router,
        protected _route: ActivatedRoute,
        protected _location: Location,
        protected _searchRequestBuilder: SearchRequestBuilder,
        protected _searchResultsProvider: SearchResultsProvider,
        protected searchService: SearchService,
        private displayer: DataDisplayer,
        private titleService: Title) {
            _searchResultsProvider.searchStartingCallback = (context) => {};
            _searchResultsProvider.searchCompletedCallback = (context) => this.loadItemCompleted();

            this.nearbySearchResultsProvider = new SearchResultsProvider(searchService, _route, _searchRequestBuilder);
            this.nearbySearchResultsProvider.searchStartingCallback = (context) => {};
            this.nearbySearchResultsProvider.searchCompletedCallback = (context) => this.nearbySearchCompleted();

            this.bydaySearchResultsProvider = new SearchResultsProvider(searchService, _route, _searchRequestBuilder);
            this.bydaySearchResultsProvider.searchStartingCallback = (context) => {};
            this.bydaySearchResultsProvider.searchCompletedCallback = (context) => this.bydaySearchCompleted();
    }

    ngOnInit() {
        this.itemInfo = undefined;
        this.nearbySearchResultsProvider.initializeRequest('', 'l');
        this.bydaySearchResultsProvider.initializeRequest('', 'd');
        this._searchResultsProvider.initializeRequest(
            SingleItemComponent.QueryProperties + ',' + SingleItemComponent.CameraProperties + ',' + SingleItemComponent.ImageProperties,
            's');

        this._route.queryParams.subscribe(params => {
            this.loadItem();
         });
    }

    toggleMediaSource() {
        if (this.showMediaSource !== true) {
            this.showMediaSource = true;
        } else {
            this.showMediaSource = false;
        }
    }

    hasLocation() {
        return this.itemInfo.longitude && this.itemInfo.latitude;
    }

    firstItem() {
        if (this.itemIndex > 1) {
            let index = 1;
            this._router.navigate( ['singleitem'], this.getNavigationExtras({ id: this.itemInfo.id, i: index }));
        }
    }

    previousItem() {
        if (this.itemIndex > 1) {
          let index = this._searchResultsProvider.searchRequest.first - 1;
          this._router.navigate( ['singleitem'], this.getNavigationExtras({ id: this.itemInfo.id, i: index }));
      }
    }

    lastItem() {
        if (this.itemIndex < this._searchResultsProvider.searchResults.totalMatches) {
            let index = this._searchResultsProvider.searchResults.totalMatches;
            this._router.navigate( ['singleitem'], this.getNavigationExtras({ id: this.itemInfo.id, i: index }));
        }
    }

    nextItem() {
        if (this.itemIndex < this._searchResultsProvider.searchResults.totalMatches) {
            let index = this._searchResultsProvider.searchRequest.first + 1;
            this._router.navigate( ['singleitem'], this.getNavigationExtras({ id: this.itemInfo.id, i: index }));
        }
    }

    loadItem() {
        this.nearbyResults = undefined;
        this.sameDateResults = undefined;
        this.mediaSource = undefined;
        this.itemIndex = this._searchResultsProvider.searchRequest.first;
        this.searchPage = Math.round(1 + (this._searchResultsProvider.searchRequest.first / SearchResultsProvider.ItemsPerPage));
        this._searchResultsProvider.search(null);
    }

    loadItemCompleted() {
        if (this._searchResultsProvider.searchResults.groups.length > 0
            && this._searchResultsProvider.searchResults.groups[0].items.length > 0) {
            this.itemInfo = this._searchResultsProvider.searchResults.groups[0].items[0];

            this.titleService.setTitle(this.itemInfo.imageName + ' - FindAPhoto');

            // The map can't be initialized until the *ngIf finishes processesing the 'this.itemInfo' update
            // (because the element used for the map doesn't yet exist)
            let timer = setInterval( () => {
                clearTimeout(timer);
                if (this.hasLocation()) {
                    this.initializeMap();

                    let location:LatLngTuple = [this.itemInfo.latitude, this.itemInfo.longitude];
                    this.map.setView(location, this.map.getZoom());

                    if (this.marker) {
                        this.marker.setLatLng(location);
                    } else {
                        this.marker = L.marker(location, { icon: this.markerIcon });
                        this.marker.addTo(this.map);
                    }
                } else {
                    if (this.marker) {
                        this.marker.remove();
                        this.marker = null;
                    }
                }
            },
            1);

            this.loadNearby();
            this.loadSameDate();
            this.loadMediaSource();
        } else {
            this.titleService.setTitle('FindAPhoto');
            this._searchResultsProvider.serverError = 'The item cannot be found';
        }
    }

    loadMediaSource() {
        this.searchService.mediaSource(encodeURI(this.itemInfo.path)).subscribe(
            result => {
                this.mediaSource = this.objectToSourceNameValues(result, '');
            },
            error => {
                console.log('loadMediaSource failed: %o', error);
            }
        );
    }

    objectToSourceNameValues(obj: Object, prefix: string): SourceNameValue[] {
        let values = Array<SourceNameValue>();

        for (let key in obj) {
            if (obj.hasOwnProperty(key)) {
                let val = obj[key];
                let name = prefix + key;
                let snv: SourceNameValue;

                if (typeof val === 'string' || typeof val === 'number') {
                    snv = new SourceNameValue(name, String(val));
                } else {
                    if (val instanceof Array) {
                        snv = new SourceNameValue(name, val.join(', '));
                    } else {
                        values.push(...this.objectToSourceNameValues(val, name + '.'));
                    }
                }

                if (snv) {
                    values.push(snv);
                }
            }
        }

        return values;
    }

    loadNearby() {
        if (!this.hasLocation()) { return; }

        this.nearbySearchResultsProvider.searchRequest.searchType = 'l';
        this.nearbySearchResultsProvider.searchRequest.latitude = this.itemInfo.latitude;
        this.nearbySearchResultsProvider.searchRequest.longitude = this.itemInfo.longitude;
        this.nearbySearchResultsProvider.searchRequest.properties = SingleItemComponent.NearbyProperties;
        this.nearbySearchResultsProvider.searchRequest.first = 1;
        this.nearbySearchResultsProvider.searchRequest.pageCount = 7;

        this.nearbySearchResultsProvider.search(null);
    }

    nearbySearchCompleted() {
        let results = this.nearbySearchResultsProvider.searchResults;
        if (results && results.groups.length > 0 && results.groups[0].items.length > 0) {
            let items = Array<SearchItem>();
            let list = results.groups[0].items;
            for (let index = 0; index < list.length && items.length < 5; ++index) {
                let si = list[index];
                if (si.id !== this.itemInfo.id) {
                    items.push(si);
                }
            }

            this.nearbyResults = items;
        } else {
            this.nearbySearchResultsProvider.serverError = 'No nearby results';
        }
    }

    loadSameDate() {
        let month = this.displayer.itemMonth(this.itemInfo);
        let day = this.displayer.itemDay(this.itemInfo);
        if (month < 0 || day < 0) { return; }


        this.bydaySearchResultsProvider.searchRequest.searchType = 'd';
        this.bydaySearchResultsProvider.searchRequest.byDayRandom = true;
        this.bydaySearchResultsProvider.searchRequest.month = month;
        this.bydaySearchResultsProvider.searchRequest.day = day;
        this.bydaySearchResultsProvider.searchRequest.properties = SingleItemComponent.SameDateProperties;
        this.bydaySearchResultsProvider.searchRequest.first = 1;
        this.bydaySearchResultsProvider.searchRequest.pageCount = 7;

        this.bydaySearchResultsProvider.search(null);
    }

    bydaySearchCompleted() {
        let results = this.bydaySearchResultsProvider.searchResults;
        if (results && results.groups.length > 0 && results.groups[0].items.length > 0) {
            let items = Array<SearchItem>();
            let list = results.groups[0].items;
            for (let index = 0; index < list.length && items.length < 5; ++index) {
                let si = list[index];
                if (si.id !== this.itemInfo.id) {
                    items.push(si);
                }
            }

            this.sameDateResults = items;
        } else {
            this.nearbySearchResultsProvider.serverError = 'No results with the same date';
        }
    }

    home() {
        this._router.navigate( ['search'], this.getNavigationExtras() );
    }

    searchToday() {
        this._router.navigate( ['byday'] );
    }

    searchNearby() {
        this._router.navigate( ['bylocation'] );
    }

    getNavigationExtras(extraParams?: Object) {
        let params = this._searchRequestBuilder.toLinkParametersObject(this._searchResultsProvider.searchRequest);
        if (extraParams !== undefined) {
            Object.assign(params, extraParams);
        }
        let navigationExtras: NavigationExtras = { queryParams: params };
        return navigationExtras;
    }

    initializeMap() {
        if (this.map) { return; }

        this.markerIcon = L.icon({
            iconUrl: 'assets/leaflet/marker-icon.png',
            iconRetinaUrl: 'assets/leaflet/marker-icon-2x.png',
            shadowUrl: 'assets/leaflet/marker-shadow.png',
            iconSize:    [25, 41],
            iconAnchor:  [12, 41],
            popupAnchor: [1, -34],
            shadowSize:  [41, 41]
        });

        this.map = L.map('singleMap', {
            center: [20, 0],
            zoom: 14,
            minZoom: 3,
            zoomControl: false
        });

        L.control.zoom({ position: 'topright' }).addTo(this.map);

        L.tileLayer('http://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            maxZoom: 19,
            attribution: '&copy; <a href="http://openstreetmap.org">OpenStreetMap</a>'
        }).addTo(this.map);
    }
}
