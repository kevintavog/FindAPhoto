import { Injectable } from '@angular/core';
import { Http, Response, ResponseType } from '@angular/http';
import { Observable } from 'rxjs/Observable';
import 'rxjs/Rx';


export class FPLocationAccuracy {
    static FromDevice = 1;
    static FromIPAddress = 2;
}

export class FPLocation {
    accuracy: FPLocationAccuracy;
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

    constructor(private _http: Http) { }

    getCurrentLocation(successCallback: FPLocationCallback, errorCallback: FPLocationErrorCallback) {
        let haveResponse = false;

        if (window.navigator.geolocation) {
            window.navigator.geolocation.getCurrentPosition(
                (position: Position) => {
                    haveResponse = true;
                    let location = new FPLocation();
                    location.accuracy = FPLocationAccuracy.FromDevice;
                    location.latitude = position.coords.latitude;
                    location.longitude = position.coords.longitude;
                    successCallback(location);
                },
                (error: PositionError) => {
                    if (!haveResponse) {
                        this.getLocationFromIp(
                            successCallback, errorCallback, 
                            'Unable to get location: ' + error.message + ' (' + error.code + ')');
                    }
                },
                { timeout: 5000 });
        } else {
            this.getLocationFromIp(successCallback, errorCallback, 'Unable to get location service from browser.');
        }
    }

    // When we can't get a location from the browser, we fall back to getting the (much less accurate) location from
    // the external IP address. Better than nothing, I hope.
    getLocationFromIp(successCallback: FPLocationCallback, errorCallback: FPLocationErrorCallback, originalError: string) {
        let url = 'http://freegeoip.net/json/';
        return this._http.get(url)
            .map(response => response.json())
            .catch(this.handleGeoFromIpError).subscribe(
                results => {
                    let location = new FPLocation();
                    location.accuracy = FPLocationAccuracy.FromIPAddress;
                    location.latitude = results.latitude;
                    location.longitude = results.longitude;
                    successCallback(location);
                },
                error => {
                    errorCallback(originalError);
                }
            );
    }

    private handleGeoFromIpError(response: Response) {
        if (response.type === ResponseType.Error) {
            console.log('GeoFromIp server is not accessible')
        } else {
            let error = response.json();
            if (error) {
                console.log('The server failed with: ' + error.errorCode + '; ' + error.errorMessage);
            } else {
                console.log('GeoFromIp server returned ' + response.text());
            }
        }

        return Observable.throw('failed');
    }
}
