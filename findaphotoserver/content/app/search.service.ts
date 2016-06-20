import { Injectable } from '@angular/core';
import { Http, RequestOptionsArgs, Response, Headers, ResponseType } from '@angular/http'
import { Observable } from 'rxjs/Observable';
import 'rxjs/Rx';

import { SearchRequest } from './search-request';

@Injectable()
export class SearchService {
    constructor(private _http: Http) {}

    search(request: SearchRequest) {
        switch (request.searchType) {
            case 's':
                return this.searchByText(request.searchText, request.properties, request.first, request.pageCount, request.drilldown)
            case 'd':
                return this.searchByDay(request.month, request.day, request.properties, request.first, request.pageCount, false, request.drilldown)
            case 'l':
                return this.searchByLocation(request.latitude, request.longitude, request.properties, request.first, request.pageCount, request.drilldown)
        }

        return Observable.throw("Unknown search type: " + request.searchType)
    }

  searchByText(searchText: string, properties: string, first: number, pageCount: number, drilldown: string) {
      var url = "/api/search?q=" + searchText + "&first=" + first + "&count=" + pageCount + "&properties=" + properties + "&categories=keywords,placename,date"
      if (drilldown != undefined && drilldown.length > 0) {
          url += "&drilldown=" + drilldown
      }
      return this._http.get(url)
                  .map(response => response.json())
                  .catch(this.handleError);
  }

  searchByDay(month: number, day: number, properties: string, first: number, pageCount: number, random: boolean, drilldown: string) {
      var url = "/api/by-day?month=" + month + "&day=" + day + "&first=" + first + "&count=" + pageCount + "&properties=" + properties + "&categories=keywords,placename,year"
      if (drilldown != undefined && drilldown.length > 0) {
          url += "&drilldown=" + drilldown
      }
      if (random) {
          url += "&random=" + random
      }

      return this._http.get(url)
                  .map(response => response.json())
                  .catch(this.handleError);
  }

  searchByLocation(lat: number, lon: number, properties: string, first: number, pageCount: number, drilldown: string) {
      var url = "/api/nearby?lat=" + lat + "&lon=" + lon + "&first=" + first + "&count=" + pageCount + "&properties=" + properties + "&categories=keywords,date"
      if (drilldown != undefined && drilldown.length > 0) {
          url += "&drilldown=" + drilldown
      }
      return this._http.get(url)
                    .map(response => response.json())
                    .catch(this.handleError);
  }

  private handleError(response: Response) {
      if (response.type == ResponseType.Error) {
          return Observable.throw("Server not accessible");
      }

      let error = response.json()
      if (!error) {
          return Observable.throw("The server returned: " + response.text());
      }

      return Observable.throw("The server failed with: " + error.errorCode + "; " + error.errorMessage);
  }
}
