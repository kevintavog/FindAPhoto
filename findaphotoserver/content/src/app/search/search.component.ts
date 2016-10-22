import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, NavigationExtras, Router } from '@angular/router';
import { Location } from '@angular/common';

import { Observable }         from 'rxjs/Observable';


import { BaseSearchComponent } from '../base-search/base-search.component';
import { SearchRequestBuilder } from '../models/search.request.builder';

import { SearchService } from '../services/search.service';

@Component({
    selector: 'app-search',
    templateUrl: './search.component.html',
    styleUrls: ['./search.component.css']
})

export class SearchComponent extends BaseSearchComponent implements OnInit {
    resultsSearchText: string;


    constructor(
            router: Router,
            route: ActivatedRoute,
            location: Location,
            searchRequestBuilder: SearchRequestBuilder,
            searchService: SearchService) {
        super("/search", router, route, location, searchRequestBuilder, searchService);
    }

    ngOnInit() {
        this.showSearch = true
        this.showResultCount = true
        this.initializeSearchRequest('s')


        this._route.queryParams.subscribe(params => {
            if ('q' in params || 't' in params) {
                this.internalSearch(false)
            }
        })
    }

    userSearch() {
        // If the search is new or different, navigate so we can use browser back to get to previous search results
        if (this.resultsSearchText && this.resultsSearchText != this.searchRequest.searchText) {
            let navigationExtras: NavigationExtras = {
                queryParams: { q: this.searchRequest.searchText }
            };

            this._router.navigate( ['search'], navigationExtras);
            return
        }

        this.internalSearch(true)
    }

    processSearchResults() {
        this.resultsSearchText = this.searchRequest.searchText
    }
}
