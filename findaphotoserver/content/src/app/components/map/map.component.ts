
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, NavigationExtras, Router } from '@angular/router';
import { Title } from '@angular/platform-browser';

import * as leaflet from 'leaflet';
import { MarkerClusterGroup } from 'leaflet.markercluster';

import { SearchItem } from '../../models/search-results';
import { SearchRequestBuilder } from '../../models/search.request.builder';

import { DataDisplayer } from '../../providers/data-displayer';
import { FPLocationAccuracy, LocationProvider } from '../../providers/location.provider';
import { NavigationProvider } from '../../providers/navigation.provider';
import { SearchResultsProvider } from '../../providers/search-results.provider';

interface SelectedItem {
    key: string;
    selected: boolean;
}


@Component({
    selector: 'app-map',
    templateUrl: './map.component.html',
    styleUrls: ['./map.component.css']
})

export class MapComponent implements OnInit {
    public static QueryProperties: string = 'createdDate,id,imageName,latitude,longitude,locationDisplayName,thumbUrl';

    readableSearchString: string;

    searchBarVisible: boolean;
    resultsSearchText: string;  // The last search text

    markerIcon: leaflet.Icon;
    selectedMarkerIcon: leaflet.Icon;


    choosingLocation: boolean;
    choosingLocationMarker: leaflet.Marker;

    map: leaflet.Map;
    defaultDoubleClickZoom: leaflet.Handler;
    cluster: MarkerClusterGroup;
    selectedMarker: leaflet.Marker;
    currentLocationCircle: leaflet.Circle;

    activeClusterList: leaflet.Marker[];
    allClusterLayerGroups = new Map<string, leaflet.LayerGroup>();

    routeListOpen: boolean;
    selectedRouteKeys: SelectedItem[];
    selectedRoutesLayer: L.FeatureGroup;

    allRoutes = new Map<string, L.Polyline>();
    activeRouteList: leaflet.LatLngTuple[];
    lastRouteItem: SearchItem;

    currentItem: SearchItem;
    currentIndex: number;

    southWestCornerLatLng: leaflet.LatLngTuple;
    northEastCornerLatLng: leaflet.LatLngTuple;

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
        private router: Router,
        private route: ActivatedRoute,
        private navigationProvider: NavigationProvider,
        private searchResultsProvider: SearchResultsProvider,
        private searchRequestBuilder: SearchRequestBuilder,
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
                this.startSearch(false);
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

        this.choosingLocationMarker = L.marker(
            this.map.getCenter(),
            {
                icon: this.markerIcon
            }
        );
    }

    toggleRouteList() {
        this.routeListOpen = !this.routeListOpen;
    }

    autoCloseForDropdownMenu(event) {
        if (!event.target.closest('.route-list-container')) {
            this.routeListOpen = false;
        }
    }

    startSearch(updateUrl: boolean) {
        if (updateUrl) {
            let params = this.searchRequestBuilder.toLinkParametersObject(this.searchResultsProvider.searchRequest);
            let navigationExtras: NavigationExtras = { queryParams: params };

            // If the params are the same, navigating won't change anything, so fall through to the search invocation
            if (!this.navigationProvider.hasSameQueryParams(params)) {
                this.router.navigate( ['map'], navigationExtras);
                return;
            }
        }

        // ? If 'searchBarVisible' is set outside of this timer, the page is refreshed.
        // I'm currently blaming this on something funny with the way I'm using the variable
        // and the way Angular2 handles it. Will re-test once I update Angular
        let timer = setInterval( () => {
            clearTimeout(timer);
            this.searchBarVisible = false;
        });

        this.closeImage();
        this.cluster.clearLayers();

        this.selectedRouteKeys = [];
        this.allRoutes.clear();
        this.activeRouteList = [];
        this.lastRouteItem = null;

        this.activeClusterList = [];
        this.allClusterLayerGroups.clear();

        this.currentItem = null;
        this.pageError = null;
        this.southWestCornerLatLng = [90, 180];
        this.northEastCornerLatLng = [-90, -180];

        this.totalMatches = this.matchesRetrieved = 0;
        this.searchResultsProvider.searchRequest.first = 1;
        this.isLoading = true;
        this.readableSearchString = ' -- ' + this.searchRequestBuilder.toReadableString(this.searchResultsProvider.searchRequest);

        this.searchResultsProvider.search(null);
    }

    mapSearchCompleted() {
        if (this.searchResultsProvider.searchResults) {
            let index = 0;
            for (let group of this.searchResultsProvider.searchResults.groups) {
                for (let item of group.items) {

                    if (item.latitude && item.longitude) {
                        this.updateBounds(item);

                        let latLng: leaflet.LatLngTuple = [item.latitude, item.longitude];

                        let marker = this.createMarker(item, latLng, this.searchResultsProvider.searchRequest.first + index);
                        this.updateCluster(item, marker);

                        this.updateRoute(item, latLng);
                    }

                    ++index;
                }
            }

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
                this.addClusterFromActiveList(this.lastRouteItem)
                this.addRouteFromActiveList(this.lastRouteItem);

                for (let key of Array.from(this.allRoutes.keys())) {
                    this.selectedRouteKeys.push( { key: key, selected: false});
                }

                this.isLoading = false;
            }
        }
    }

    createMarker(item: SearchItem, latLng: leaflet.LatLngTuple, index: number) {
        let marker = L.marker(
            latLng,
            {
                icon: this.markerIcon
            });

        marker.on('mouseover', () => {
            this.currentItem = item;
            this.currentIndex = index;
            this.selectMarker(marker);
        });

        marker.on('click', () => {
            this.currentItem = item;
            this.currentIndex = index;
            this.selectMarker(marker);
        });

        return marker;
    }

    updateCluster(item: SearchItem, marker: leaflet.Marker) {
        let newCluster = !this.lastRouteItem;
        if (this.lastRouteItem) {
            newCluster = this.displayer.getItemLocaleDate(item) !== this.displayer.getItemLocaleDate(this.lastRouteItem);
        }

        if (newCluster) {
            this.addClusterFromActiveList(item);
        }

        this.activeClusterList.push(marker);
    }

    addClusterFromActiveList(item: SearchItem) {
        if (this.activeClusterList.length > 1) {
            let lg = L.layerGroup(this.activeClusterList);
            this.cluster.addLayer(lg);

            let key = this.routeKeyFromItem(item);
            this.allClusterLayerGroups.set(key, lg);
        }

        this.activeClusterList = [];
    }

    updateRoute(item: SearchItem, latLng: leaflet.LatLngTuple) {
        let newRoute = !this.lastRouteItem;
        if (this.lastRouteItem) {
            newRoute = this.displayer.getItemLocaleDate(item) !== this.displayer.getItemLocaleDate(this.lastRouteItem);
            if (this.lastRouteItem.latitude === item.latitude && this.lastRouteItem.longitude === item.longitude) {
                return;
            }
        }

        if (newRoute) {
            this.addRouteFromActiveList(item);
        }

        this.activeRouteList.push(latLng);
        this.lastRouteItem = item;
    }

    addRouteFromActiveList(item: SearchItem) {
        if (this.activeRouteList.length > 1) {
            let key = this.routeKeyFromItem(item);
            let route = L.polyline(this.activeRouteList, {color: 'red' } );
            this.allRoutes.set(key, route);
        }

        this.activeRouteList = [];
    }

    routeKeyFromItem(item: SearchItem) {
        let key = this.displayer.getItemLocaleDate(item);
        if (this.lastRouteItem) {
            key = this.displayer.getItemLocaleDate(this.lastRouteItem);
        }
        return key;
    }

    updateBounds(item: SearchItem) {
        if (item.latitude < this.southWestCornerLatLng[0]) {
            this.southWestCornerLatLng[0] = item.latitude;
        }
        if (item.longitude < this.southWestCornerLatLng[1]) {
            this.southWestCornerLatLng[1] = item.longitude;
        }

        if (item.latitude > this.northEastCornerLatLng[0]) {
            this.northEastCornerLatLng[0] = item.latitude;
        }
        if (item.longitude > this.northEastCornerLatLng[1]) {
            this.northEastCornerLatLng[1] = item.longitude;
        }
    }

    selectedRoutesChanged(si: SelectedItem) {
        const colors = ['#F44336', '#9C27B0', '#3F51B5', '#2196F3', '#009688', '#FF9800', '#795548'];

        this.selectedRoutesLayer.clearLayers();

        si.selected = !si.selected;

        // Remove all layers from Cluster
        this.cluster.clearLayers();

        let selectedCount = 0;
        for (let k of this.selectedRouteKeys) {
            if (k.selected) {
                let route = this.allRoutes.get(k.key);
                if (route) {
                    route.setStyle({color: colors[selectedCount % colors.length]});
                    this.selectedRoutesLayer.addLayer(route);
                    ++selectedCount;

                    let lg = this.allClusterLayerGroups.get(k.key);
                    if (lg) {
                        this.cluster.addLayer(lg);
                    }
                }
            }
        }

        if (selectedCount > 0) {
            this.map.fitBounds(this.selectedRoutesLayer.getBounds(), null);
        } else {
            // Nothing is selected, add all layergroups to cluster
            for (let lg of Array.from(this.allClusterLayerGroups.values())) {
                this.cluster.addLayer(lg);
            }
        }
    }

    chooseLocation() {
        this.closeImage();
        this.cluster.clearLayers();
        this.choosingLocation = true;
        this.map.doubleClickZoom.disable();
        this.pageError = 'Double click/tap to choose a location';
        if (this.currentLocationCircle) {
            this.currentLocationCircle.remove();
        }

        this.choosingLocationMarker.setLatLng(this.map.getCenter());
        this.choosingLocationMarker.addTo(this.map);

        this.map.on('dblclick', (le: leaflet.LocationEvent) => {
            this.map.panTo(le.latlng);
            this.choosingLocationMarker.setLatLng(le.latlng);
        });

        this.map.on('moveend', () => {
            this.choosingLocationMarker.setLatLng(this.map.getCenter());
        });
    }

    endChooseLocation() {
        this.map.off('dblclick');
        this.choosingLocationMarker.removeFrom(this.map);

        this.map.doubleClickZoom.enable();
        this.choosingLocation = false;
        this.pageError = '';

        this.locationAccuracy = FPLocationAccuracy.FromDevice;
        let center = this.map.getCenter();
        this.searchNear(center.lat, center.lng);
    }

    toggleSearchBar() {
        this.searchBarVisible = !this.searchBarVisible;
    }

    searchWithText() {
        this.searchResultsProvider.searchRequest.searchType = 's';
        this.startSearch(true);
    }

    fitBounds() {
        this.map.fitBounds([this.southWestCornerLatLng, this.northEastCornerLatLng], null);
    }

    nearby() {
        this.cluster.clearLayers();
        this.currentItem = null;
        this.isLoading = true;
        this.pageError = 'Getting current location...';

        this.locationProvider.getCurrentLocation(
            location => {
                this.locationAccuracy = location.accuracy;
                this.searchNear(location.latitude, location.longitude);
                this.map.setView([location.latitude, location.longitude], 17);
            },
            error => {
                this.pageError = 'Unable to get current location: ' + error;
            });
    }

    searchNear(latitude: number, longitude: number) {
        this.isLoading = true;
        this.maxMatchesAllowed = 2000;
        this.fitBoundsOnFirstResults = false;
        this.pageError = '';

        this.searchResultsProvider.searchRequest.searchType = 'l';
        this.searchResultsProvider.searchRequest.latitude = latitude;
        this.searchResultsProvider.searchRequest.longitude = longitude;
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

        this.currentLocationCircle = L.circle([latitude, longitude], radius, circleProperties);
        this.currentLocationCircle.addTo(this.map);

        this.startSearch(true);

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

    closeImage() {
        if (this.selectedMarker) {
            this.selectedMarker.setIcon(this.markerIcon);
            this.selectedMarker = null;
        }
        this.currentItem = null;
        this.currentIndex = -1;
    }

    singleItemSearchLinkParameters(item: SearchItem, imageIndex: number) {
        let properties = this.searchRequestBuilder.toLinkParametersObject(this.searchResultsProvider.searchRequest);
        properties['id'] = item.id;
        properties['i'] = imageIndex;
        return properties;
    }


    initializeMap() {
        if (this.map) { return; }

        this.map = L.map('map', {
            center: [20, 0],
            zoom: 3,
            minZoom: 3,
            zoomControl: false
        });

        this.defaultDoubleClickZoom = this.map.doubleClickZoom;

        L.control.zoom({ position: 'topright' }).addTo(this.map);

        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            maxZoom: 19,
            attribution: '&copy; <a href="http://openstreetmap.org">OpenStreetMap</a> ' +
                'contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>'
        }).addTo(this.map);

        this.cluster = L.markerClusterGroup( { showCoverageOnHover: false } );
        this.map.addLayer(this.cluster);

        this.selectedRoutesLayer = L.featureGroup();
        this.map.addLayer(this.selectedRoutesLayer);
    }
}
