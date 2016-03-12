import { Component, OnInit } from 'angular2/core';
import { Router, ROUTER_DIRECTIVES, RouteParams, Location } from 'angular2/router';

@Component({
  selector: 'root',
  template: '',
  directives: [ROUTER_DIRECTIVES],
})

// This class only exists because I clearly don't understand something.
// The problem I believe it solves is that it passes along query parameters
// to the default search component. It seems to be an issue with the way
// I've setup the routing, but I'm clueless on a proper solution.
export class RootComponent implements OnInit {
    constructor(
      private _router: Router,
      private _routeParams: RouteParams,
      private _location: Location) { }

    ngOnInit() {
        let path = this._location.path()
        let query = ""
        let i = path.indexOf("?")
        if (i >= 0) {
            query = path.substring(i)
        }
        this._router.navigateByUrl("/search" + query)
    }
}
