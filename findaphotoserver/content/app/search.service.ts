import { SearchRequest } from './search-request';
import { Injectable } from 'angular2/core';
import { Http, RequestOptionsArgs, Response, Headers } from 'angular2/http'
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

  private handleError(response: Response) {
      console.error("Server returned an error: ", response.statusText, ": ", response.text())

      let error = response.json()
      if (!error) {
          return Observable.throw(response.text());
      }

      return Observable.throw(error.errorCode + "; " + error.errorMessage);
  }
}
