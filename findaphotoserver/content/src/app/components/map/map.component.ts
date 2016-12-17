
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { Title } from '@angular/platform-browser';

import { Circle, Icon, LatLngTuple, Map, Marker } from 'leaflet';
import { MarkerClusterGroup } from 'leaflet.markercluster';

import { SearchItem } from '../../models/search-results';

import { DataDisplayer } from '../../providers/data-displayer';
import { FPLocationAccuracy, LocationProvider } from '../../providers/location.provider';
import { NavigationProvider } from '../../providers/navigation.provider';
import { SearchResultsProvider } from '../../providers/search-results.provider';


@Component({
    selector: 'app-map',
    templateUrl: './map.component.html',
    styleUrls: ['./map.component.css']
})

export class MapComponent implements OnInit {
    public static QueryProperties: string = 'createdDate,id,imageName,latitude,longitude,locationDisplayName,thumbUrl';

    markerIcon: Icon;
    selectedMarkerIcon: Icon;

    map: Map;
    cluster: MarkerClusterGroup;
    selectedMarker: Marker;
    currentLocationCircle: Circle;

    currentItem: SearchItem;

    southWestCornerLatLng: LatLngTuple;
    northEastCornerLatLng: LatLngTuple;

    isLoading: boolean;
    pageError: string;
    totalMatches: number;
    matchesRetrieved: number;
    maxMatchesAllowed: number;      // Only set for a nearby search
    fitBoundsOnFirstResults: boolean;

    locationAccuracy: FPLocationAccuracy;


    get percentageLoadedWidth() {
        return this.percentageLoaded.toString() + '%';
    }

    get percentageLoaded() {
        if (!this.isLoading) { return 100; }
        return Math.round(this.matchesRetrieved * 100 / this.totalMatches);
    }


    constructor(
        private route: ActivatedRoute,
        private navigationProvider: NavigationProvider,
        private searchResultsProvider: SearchResultsProvider,
        private displayer: DataDisplayer,
        private titleService: Title,
        private locationProvider: LocationProvider) {

            searchResultsProvider.searchStartingCallback = (context) => {};
            searchResultsProvider.searchCompletedCallback = (context) => this.mapSearchCompleted();
            this.fitBoundsOnFirstResults = true;
    }

    ngOnInit() {
        this.titleService.setTitle('Map - FindAPhoto');
        this.navigationProvider.initialize();
        this.searchResultsProvider.initializeRequest(MapComponent.QueryProperties, 's');
        this.searchResultsProvider.searchRequest.pageCount = 100;
        this.initializeMap();


        this.route.queryParams.subscribe(params => {
            if (('q' in params && params['q'] !== '') || ('t' in params && params['t'] !== 's')) {
                this.startSearch();
            }
        });

        this.markerIcon = L.icon({
            iconUrl: 'assets/leaflet/marker-icon.png',
            iconRetinaUrl: 'assets/leaflet/marker-icon-2x.png',
            shadowUrl: 'assets/leaflet/marker-shadow.png',
            iconSize:    [25, 41],
            iconAnchor:  [12, 41],
            popupAnchor: [1, -34],
            shadowSize:  [41, 41]
        });

        this.selectedMarkerIcon = L.icon({
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
        this.currentItem = null;
        this.pageError = null;
        this.southWestCornerLatLng = [90, 180];
        this.northEastCornerLatLng = [-90, -180];

        this.totalMatches = this.matchesRetrieved = 0;
        this.searchResultsProvider.searchRequest.first = 1;
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
                                icon: this.markerIcon
                            });

                        marker.on('mouseover', () => {
                            this.currentItem = item;
                            this.selectMarker(marker);
                        });

                        marker.on('click', () => {
                            this.currentItem = item;
                            this.selectMarker(marker);
                        });
                        markers.push(marker);
                    }
                }
            }

            this.cluster.addLayer(L.layerGroup(markers));


            let results = this.searchResultsProvider.searchResults;
            let request = this.searchResultsProvider.searchRequest;


            // Only fit bounds after the first search - otherwise, the map will jump around, which is unpleasant.
            if (this.fitBoundsOnFirstResults && request.first === 1) {
                this.fitBounds();
            }

            if (this.maxMatchesAllowed > 0) {
                this.totalMatches = Math.min(results.totalMatches, this.maxMatchesAllowed);
            } else {
                this.totalMatches = results.totalMatches;
            }

            this.matchesRetrieved = request.first + results.resultCount - 1;

            if (results.resultCount > 0 && this.matchesRetrieved < this.totalMatches) {
                this.searchResultsProvider.searchRequest.first = request.first + request.pageCount;
                this.searchResultsProvider.search(null);
            } else {
                this.isLoading = false;
            }
        }
    }

    fitBounds() {
        this.map.fitBounds([this.southWestCornerLatLng, this.northEastCornerLatLng], null);
    }

    nearby() {
        this.cluster.clearLayers();
        this.currentItem = null;
        this.isLoading = true;
        this.maxMatchesAllowed = 2000;
        this.fitBoundsOnFirstResults = false;
        this.pageError = 'Getting current location...';

        this.locationProvider.getCurrentLocation(
            location => {
                this.pageError = '';
                this.locationAccuracy = location.accuracy;
                this.searchResultsProvider.searchRequest.searchType = 'l';
                this.searchResultsProvider.searchRequest.latitude = location.latitude;
                this.searchResultsProvider.searchRequest.longitude = location.longitude;
                this.searchResultsProvider.searchRequest.maxKilometers = 10.5;

                if (this.currentLocationCircle) {
                    this.currentLocationCircle.remove();
                }

                let radius = 50;
                let circleProperties = { color: '#0000FF', fillColor: '#00f' };
                if (this.locationAccuracy !== FPLocationAccuracy.FromDevice) {
                    radius = 150;
                    circleProperties = { color: '#DC143C', fillColor: '#FF0000' };
                }

                this.currentLocationCircle = L.circle([location.latitude, location.longitude], radius, circleProperties);
                this.currentLocationCircle.addTo(this.map);
                this.map.setView([location.latitude, location.longitude], 17);

                this.startSearch();
            },
            error => {
                this.pageError = 'Unable to get current location: ' + error;
            });
    }

    selectMarker(marker: L.Marker) {
        if (this.selectedMarker) {
            this.selectedMarker.setIcon(this.markerIcon);
            this.selectedMarker = null;
        }

        if (marker) {
            this.selectedMarker = marker;
            this.selectedMarker.setIcon(this.selectedMarkerIcon);
        }
    }

    initializeMap() {
        if (this.map) { return; }

        this.map = L.map('map', {
            center: [20, 0],
            zoom: 3,
            minZoom: 3,
            zoomControl: false
        });

        L.control.zoom({ position: 'topright' }).addTo(this.map);

        L.tileLayer('http://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            maxZoom: 19,
            attribution: '&copy; <a href="http://openstreetmap.org">OpenStreetMap</a> ' +
                'contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>'
        }).addTo(this.map);

        this.cluster = L.markerClusterGroup( { showCoverageOnHover: false } );
        this.map.addLayer(this.cluster);
    }
}
