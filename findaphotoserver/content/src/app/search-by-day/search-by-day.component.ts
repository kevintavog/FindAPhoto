import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, NavigationExtras, Router } from '@angular/router';
import { Location } from '@angular/common';

import { BaseSearchComponent } from '../base-search/base-search.component';
import { SearchRequestBuilder } from '../models/search.request.builder';
import { ByDayResult } from '../models/search-results';

import { SearchService } from '../services/search.service';

@Component({
  selector: 'app-search-by-day',
  templateUrl: '../search/search.component.html',
  styleUrls: ['../search/search.component.css']
})

export class SearchByDayComponent extends BaseSearchComponent implements OnInit {
    private static monthNames: string[] = [ "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December" ];

    activeDate: Date;


    constructor(
      router: Router,
      route: ActivatedRoute,
      location: Location,
      searchRequestBuilder: SearchRequestBuilder,
      searchService: SearchService) {
          super("/byday", router, route, location, searchRequestBuilder, searchService);
    }

    ngOnInit() {
        this.showSearch = false
        this.showResultCount = true
        this.initializeSearchRequest('d')
        this.activeDate = new Date(2016, this.searchRequest.month - 1, this.searchRequest.day, 0, 0, 0, 0)
        this.internalSearch(false)
    }

    processSearchResults() {
        // DOES NOT honor locale...
        this.pageMessage = "Pictures from " + SearchByDayComponent.monthNames[this.activeDate.getMonth()] + "  " + this.activeDate.getDate()

        this.typeLeftButtonClass = "fa fa-arrow-left"
        this.typeLeftButtonText = this.byDayString(this.searchResults.previousAvailableByDay)
        this.typeRightButtonClass = "fa fa-arrow-right"
        this.typeRightButtonText = this.byDayString(this.searchResults.nextAvailableByDay)
    }

    byDayString(byday: ByDayResult) {
        if (byday == undefined)
            return null
        return SearchByDayComponent.monthNames[byday.month - 1] + " " + byday.day
    }

    typeLeftButton() {
        let navigationExtras: NavigationExtras = {
            queryParams: { m:this.searchResults.previousAvailableByDay.month, d:this.searchResults.previousAvailableByDay.day }
        };

        this._router.navigate( ['byday'], navigationExtras);
    }

    typeRightButton() {
        let navigationExtras: NavigationExtras = {
            queryParams: { m:this.searchResults.nextAvailableByDay.month, d:this.searchResults.nextAvailableByDay.day }
        };

        this._router.navigate( ['byday'], navigationExtras);
    }
}
