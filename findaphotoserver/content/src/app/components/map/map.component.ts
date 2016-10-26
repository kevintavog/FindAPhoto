import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

import { Map } from "leaflet";

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
    map: Map
    thumbsInStrip = new Array<SearchItem>()
    mapSearchResultsProvider: SearchResultsProvider
    distanceMarkerGroup: L.FeatureGroup
    markerIcon: L.Icon
    highlightedMarkerIcon: L.Icon
    selectedMarker: L.Marker


    constructor(
        private route: ActivatedRoute,
        private navigationProvider: NavigationProvider,
        searchRequestBuilder: SearchRequestBuilder,
        searchService: SearchService) {

            this.mapSearchResultsProvider = new SearchResultsProvider(searchService, route, searchRequestBuilder)
            this.mapSearchResultsProvider.searchStartingCallback = (context) => {}
            this.mapSearchResultsProvider.searchCompletedCallback = (context) => this.mapSearchCompleted()
    }

    ngOnInit() {
        this.navigationProvider.initialize()
        this.mapSearchResultsProvider.initializeRequest(SearchResultsProvider.QueryProperties, 's')
        this.initializeMap()


        this.route.queryParams.subscribe(params => {
            if ('q' in params || 't' in params) {
                this.mapSearchResultsProvider.search(null)
            }
        })

// Will want to use the leaflet markercluster
// https://github.com/Leaflet/Leaflet.markercluster
// https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/leaflet-markercluster



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
//
//
// L.marker([51.51195,-0.1322], { icon: markerIcon }).addTo(this.map);
// let marker = L.marker([51.5, -0.09]).addTo(this.map);
let circle = L.circle([51.508, -0.11], {
    color: 'red',
    fillColor: '#f03',
    fillOpacity: 0.5,
    radius: 500
}).addTo(this.map);

    }

    mapSearchCompleted() {
        this.thumbsInStrip = new Array<SearchItem>()
        if (this.mapSearchResultsProvider.searchResults) {
            for (let group of this.mapSearchResultsProvider.searchResults.groups) {
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

            let markers = new Array<L.Marker>();
            for (let group of this.mapSearchResultsProvider.searchResults.groups) {
                for (let item of group.items) {
                    let marker = L.marker(
                        [item.latitude, item.longitude],
                        {
                            title: item.imageName,
                            icon: this.markerIcon
                        })

                    marker.on('click', () => {
                        this.removeHighlight()
                        marker.setIcon(this.highlightedMarkerIcon)
                        this.selectedMarker = marker
                    })
                    markers.push(marker)
                }
            }

            let group = L.featureGroup([L.layerGroup(markers)]).addTo(this.map);
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
    }
}
