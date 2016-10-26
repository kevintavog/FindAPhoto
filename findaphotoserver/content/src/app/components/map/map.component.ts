
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

import { FeatureGroup, Icon, Layer, LayerGroup, Map, Marker } from "leaflet";
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

    map: Map;
    cluster: MarkerClusterGroup;
    thumbsInStrip = new Array<SearchItem>();
    distanceMarkerGroup: FeatureGroup;
    markerIcon: Icon;
    highlightedMarkerIcon: Icon;
    selectedMarker: Marker;


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
        this.thumbsInStrip = new Array<SearchItem>();
        this.searchResultsProvider.search(null);
    }

    mapSearchCompleted() {
        if (this.searchResultsProvider.searchResults) {
            for (let group of this.searchResultsProvider.searchResults.groups) {
                if (this.thumbsInStrip.length >= 50) {
                    break
                }
                for (let item of group.items) {
                    this.thumbsInStrip.push(item)

                    if (this.thumbsInStrip.length >= 50) {
                        break
                    }
                }
            }

            let markers = new Array<Marker>();
            for (let group of this.searchResultsProvider.searchResults.groups) {
                for (let item of group.items) {
                    if (item.latitude && item.longitude) {
                        let marker = L.marker(
                            [item.latitude, item.longitude],
                            {
                                title: item.imageName,
                                icon: this.markerIcon
                            });

                        marker.on('click', () => {
                            this.removeHighlight();
                            marker.setIcon(this.highlightedMarkerIcon);
                            this.selectedMarker = marker;
                        })
                        markers.push(marker);
                    }
                }
            }

            this.cluster.addLayer(L.layerGroup(markers));

            let results = this.searchResultsProvider.searchResults;
            let request = this.searchResultsProvider.searchRequest;
            let totalMatches = results.totalMatches;
            let retrieved = request.first + results.resultCount - 1;
            if (retrieved < totalMatches) {
                this.searchResultsProvider.searchRequest.first = request.first + request.pageCount;
                this.searchResultsProvider.search(null);
            }
        }
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

        this.map.on('click', () => { this.removeHighlight() })

        this.cluster = L.markerClusterGroup( { showCoverageOnHover: false } );
        this.map.addLayer(this.cluster);
    }
}
