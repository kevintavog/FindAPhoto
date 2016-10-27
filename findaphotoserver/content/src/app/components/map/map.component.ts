
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

import { FeatureGroup, Icon, LatLngBounds, LatLngBoundsLiteral, LatLngTuple, Layer, LayerGroup, Map, Marker, Popup } from "leaflet";
import { MarkerClusterGroup } from 'leaflet.markercluster';

import { SearchItem } from '../../models/search-results'
import { SearchRequestBuilder } from '../../models/search.request.builder';

import { NavigationProvider } from '../../providers/navigation.provider';
import { SearchResultsProvider } from '../../providers/search-results.provider';

import { SearchService } from '../../services/search.service';


@Component({
    selector: 'app-map',
    templateUrl: './map.component.html',
    styleUrls: ['./map.component.css']
})

export class MapComponent implements OnInit {
    public static QueryProperties: string = "id,imageName,latitude,longitude,thumbUrl"

    markerIcon: Icon;
    highlightedMarkerIcon: Icon;

    map: Map;
    cluster: MarkerClusterGroup;
    selectedMarker: Marker;
    popup: Popup;

    southWestCornerLatLng: LatLngTuple;
    northEastCornerLatLng: LatLngTuple;

    isLoading: boolean;
    totalMatches: number;
    matchesRetrieved: number;


    get percentageLoadedWidth() {
        return this.percentageLoaded.toString() + "%";
    }

    get percentageLoaded() {
        if (!this.isLoading) { return 100; }
        return Math.round(this.matchesRetrieved * 100 / this.totalMatches);
    }


    constructor(
        private route: ActivatedRoute,
        private navigationProvider: NavigationProvider,
        private searchResultsProvider: SearchResultsProvider) {

            searchResultsProvider.searchStartingCallback = (context) => {};
            searchResultsProvider.searchCompletedCallback = (context) => this.mapSearchCompleted();
    }

    ngOnInit() {
        this.navigationProvider.initialize()
        this.searchResultsProvider.initializeRequest(MapComponent.QueryProperties, 's')
        this.searchResultsProvider.searchRequest.pageCount = 100;
        this.initializeMap()


        this.route.queryParams.subscribe(params => {
            if ('q' in params || 't' in params) {
                this.startSearch();
            }
        })

        this.markerIcon = L.icon({
            iconUrl: 'assets/leaflet/marker-icon.png',
            iconRetinaUrl: 'assets/leaflet/marker-icon-2x.png',
            shadowUrl: 'assets/leaflet/marker-shadow.png',
            iconSize:    [25, 41],
    		iconAnchor:  [12, 41],
    		popupAnchor: [1, -34],
    		shadowSize:  [41, 41]
        });

        this.highlightedMarkerIcon = L.icon({
            iconUrl: 'assets/leaflet/marker-highlight.png',
            iconRetinaUrl: 'assets/leaflet/marker-highlight-2x.png',
            shadowUrl: 'assets/leaflet/marker-shadow.png',
            iconSize:    [25, 41],
    		iconAnchor:  [12, 41],
    		popupAnchor: [1, -34],
    		shadowSize:  [41, 41]
        });
    }

    startSearch() {
        this.southWestCornerLatLng = [90, 180];
        this.northEastCornerLatLng = [-90, -180];

        this.totalMatches = this.matchesRetrieved = 0;
        this.isLoading = true;
        this.searchResultsProvider.search(null);
    }

    mapSearchCompleted() {
        if (this.searchResultsProvider.searchResults) {
            let markers = new Array<Marker>();
            for (let group of this.searchResultsProvider.searchResults.groups) {
                for (let item of group.items) {
                    if (item.latitude && item.longitude) {

                        if (item.latitude < this.southWestCornerLatLng[0]) { this.southWestCornerLatLng[0] = item.latitude; }
                        if (item.longitude < this.southWestCornerLatLng[1]) { this.southWestCornerLatLng[1] = item.longitude; }

                        if (item.latitude > this.northEastCornerLatLng[0]) { this.northEastCornerLatLng[0] = item.latitude; }
                        if (item.longitude > this.northEastCornerLatLng[1]) { this.northEastCornerLatLng[1] = item.longitude; }

                        let marker = L.marker(
                            [item.latitude, item.longitude],
                            {
                                title: item.imageName,
                                icon: this.markerIcon
                            });

                        marker.on('mouseover', () => {
                            this.popup.setLatLng([item.latitude, item.longitude]);
                            this.popup.setContent(
                                '<div> '
                                + `<img src="${item.thumbUrl}" (click)="showItem(${item})" >`
                                + ' </div>');
                            this.map.openPopup(this.popup);
                        });

                        marker.on('click', () => {
                            this.popup.setLatLng([item.latitude, item.longitude]);
                            this.popup.setContent(
                                '<div> '
                                + `<img src="${item.thumbUrl}" >`
                                + ' </div>');
                            this.map.openPopup(this.popup);
                        });
                        markers.push(marker);
                    }
                }
            }

            this.cluster.addLayer(L.layerGroup(markers));


            let results = this.searchResultsProvider.searchResults;
            let request = this.searchResultsProvider.searchRequest;


            // Only fit bounds after the first search - otherwise, the map will jump around, which is unpleasant.
            if (request.first == 1) {
                this.fitBounds();
            }

            this.totalMatches = results.totalMatches;
            this.matchesRetrieved = request.first + results.resultCount - 1;

            if (this.matchesRetrieved < this.totalMatches) {
                this.searchResultsProvider.searchRequest.first = request.first + request.pageCount;
                this.searchResultsProvider.search(null);
            } else {
                this.isLoading = false;
            }
        }
    }

    showItem(item: SearchItem) {
        console.log(`showItem: ${item.imageName}`);
    }

    fitBounds() {
        this.map.fitBounds([this.southWestCornerLatLng, this.northEastCornerLatLng], null);
    }

    removeHighlight() {
        if (this.selectedMarker) {
            this.selectedMarker.setIcon(this.markerIcon)
        }
    }

    initializeMap() {
        if (this.map) { return }

        this.map = L.map('map', {
            center: [20, 0],
            zoom: 3,
            minZoom: 3,
            zoomControl: false
        });

        L.control.zoom({ position: "topright" }).addTo(this.map);

        L.tileLayer('http://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            maxZoom: 19,
            attribution: '&copy; <a href="http://openstreetmap.org">OpenStreetMap</a> ' +
                'contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>'
        }).addTo(this.map);

        L.control.scale({ position: "bottomright" }).addTo(this.map);

        this.map.on('click', () => { this.removeHighlight() });


        this.cluster = L.markerClusterGroup( { showCoverageOnHover: false } );
        this.map.addLayer(this.cluster);


        this.popup = L.popup();
        // this.popup.on('click', () => {
        //     console.log("popup click")
        // })
    }
}
