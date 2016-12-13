import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import 'rxjs/Rx';

export class FPLocation {
    latitude: number;
    longitude: number;
}

export interface FPLocationCallback {
    (location: FPLocation): void;
}

interface FPLocationErrorCallback {
    (errorMessage: string): void;
}

@Injectable()
export class LocationProvider {

    getCurrentLocation(successCallback: FPLocationCallback, errorCallback: FPLocationErrorCallback) {
        if (window.navigator.geolocation) {
            window.navigator.geolocation.getCurrentPosition(
                (position: Position) => {
                    let location = new FPLocation();
                    location.latitude = position.coords.latitude;
                    location.longitude = position.coords.longitude;
                    successCallback(location);
                },
                (error: PositionError) => {
                    errorCallback('Unable to get location: ' + error.message + ' (' + error.code + ')');
                },
                { timeout: 5000 });
        } else {
            let error = new PositionError()
            errorCallback('Unable to get location service from browser.');
        }
    }
}
