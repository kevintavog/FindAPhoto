import { SearchRequest } from './search-request';
import { Injectable } from 'angular2/core';
import { Http, RequestOptionsArgs, Response, Headers, ResponseType } from 'angular2/http'
import { Observable } from 'rxjs/Observable';
import 'rxjs/Rx';

@Injectable()
export class SearchService {
 constructor(private _http: Http) {}

  search(request: SearchRequest) {
      var url = "/api/search?q=" + request.searchText + "&properties=" + request.properties + "&first=" + request.first + "&count=" + request.pageCount
      return this._http.get(url)
                  .map(response => response.json())
                  .catch(this.handleError);
  }

  today(month: number, day: number, properties: string) {
      var url = "/api/by-day?month=" + month + "&day=" + day + "&properties=" + properties
      return this._http.get(url)
                  .map(response => response.json())
                  .catch(this.handleError);
  }

  nearby(lat: number, lon: number, count: number, properties: string) {
      var url = "/api/nearby?lat=" + lat + "&lon=" + lon + "&count=" + count + "&properties=" + properties
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
