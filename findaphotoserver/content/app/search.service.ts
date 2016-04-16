import { SearchRequest } from './search-request';
import { Injectable } from 'angular2/core';
import { Http, RequestOptionsArgs, Response, Headers, ResponseType } from 'angular2/http'
import { Observable } from 'rxjs/Observable';
import 'rxjs/Rx';

@Injectable()
export class SearchService {
    constructor(private _http: Http) {}

    search(request: SearchRequest) {
        switch (request.searchType) {
            case 's':
                return this.searchByText(request.searchText, request.properties, request.first, request.pageCount)
            case 'd':
                return this.searchByDay(request.month, request.day, request.properties, request.first, request.pageCount)
            case 'l':
                return this.searchByLocation(request.latitude, request.longitude, request.properties, request.first, request.pageCount)
        }

        console.log("Unknown search type: " + request.searchType)
        return Observable.throw("Unknown search type: " + request.searchType)
    }

  searchByText(searchText, properties: string, first, pageCount: number) {
      var url = "/api/search?q=" + searchText + "&first=" + first + "&count=" + pageCount + "&properties=" + properties
      return this._http.get(url)
                  .map(response => response.json())
                  .catch(this.handleError);
  }

  searchByDay(month: number, day: number, properties: string, first, pageCount: number) {
      var url = "/api/by-day?month=" + month + "&day=" + day + "&first=" + first + "&count=" + pageCount + "&properties=" + properties
      return this._http.get(url)
                  .map(response => response.json())
                  .catch(this.handleError);
  }

  searchByLocation(lat, lon: number, properties: string, first, pageCount: number) {
      var url = "/api/nearby?lat=" + lat + "&lon=" + lon + "&first=" + first + "&count=" + pageCount + "&properties=" + properties
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
