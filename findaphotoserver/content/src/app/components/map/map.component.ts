import { Component, OnInit } from '@angular/core';

import { Map } from "leaflet";


@Component({
    selector: 'app-map',
    templateUrl: './map.component.html',
    styleUrls: ['./map.component.css']
})

export class MapComponent implements OnInit {
    map: Map


    constructor() { }

    ngOnInit() {

// Will want to use the leaflet markercluster
// https://github.com/Leaflet/Leaflet.markercluster
// https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/leaflet-markercluster

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


        let markerIcon = L.icon({
            iconUrl: 'assets/leaflet/marker-icon.png',
            iconRetinaUrl: 'assets/leaflet/marker-icon-2x.png',
            shadowUrl: 'assets/leaflet/marker-shadow.png',
            iconSize:    [25, 41],
    		iconAnchor:  [12, 41],
    		popupAnchor: [1, -34],
    		shadowSize:  [41, 41]
        });


L.marker([51.51195,-0.1322], { icon: markerIcon }).addTo(this.map);
let marker = L.marker([51.5, -0.09]).addTo(this.map);
let circle = L.circle([51.508, -0.11], {
    color: 'red',
    fillColor: '#f03',
    fillOpacity: 0.5,
    radius: 500
}).addTo(this.map);

    }

}
