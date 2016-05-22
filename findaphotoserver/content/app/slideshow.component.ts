import { Component, OnInit } from 'angular2/core';
import { Router,RouteParams, ROUTER_DIRECTIVES, Location } from 'angular2/router';

import { BaseComponent } from './base.component';
import { SearchRequest } from './search-request';
import { SearchResults, SearchGroup, SearchItem } from './search-results';
import { SearchService } from './search.service';
import { BaseSearchComponent } from './base.search.component';
import { SearchRequestBuilder } from './search.request.builder';

import { DateStringToLocaleDatePipe } from './datestring-to-localedate.pipe';

@Component({
  selector: 'slideshow',
  templateUrl: 'app/slideshow.component.html',
  styleUrls:  ['app/slideshow.component.css'],
  directives: [ROUTER_DIRECTIVES],
  pipes: [DateStringToLocaleDatePipe]
})

export class SlideshowComponent extends BaseComponent implements OnInit {
    private static QueryProperties: string = "id,mediaUrl,createdDate,keywords,locationName,mimeType,mediaType,path"
    private static SlideIntervalMsecs: number = 5000;

    slideIndex: number;
    totalSearchMatches: number;
    searchRequest: SearchRequest;
    slideInfo: SearchItem;
    error: string;

    constructor(
        private _router: Router,
        private _routeParams: RouteParams,
        private _searchService: SearchService,
        private _searchRequestBuilder: SearchRequestBuilder) { super() }

    ngOnInit() {
        this.slideInfo = undefined
        this.error = undefined

        this.searchRequest = this._searchRequestBuilder.createRequest(this._routeParams, 1, SlideshowComponent.QueryProperties, 's')
        this.searchRequest.searchType = 's'
        this.searchRequest.searchText = "keywords:mural"
        this.loadSlide()
    }

    previousSlide() {
        if (this.slideIndex > 1) {
            let index = this.searchRequest.first - 1
            this._router.navigate( ['Slideshow', this.slideshowLinkParameters(this.slideInfo, index)] );
        }
    }

    nextSlide() {
        if (this.slideIndex < this.totalSearchMatches) {
            let index = this.searchRequest.first + 1
            this._router.navigate( ['Slideshow', this.slideshowLinkParameters(this.slideInfo, index)] );
        }
    }

    loadSlide() {
        this.slideIndex = this.searchRequest.first
        this._searchService.search(this.searchRequest).subscribe(
            results => {
                if (results.groups.length > 0 && results.groups[0].items.length > 0) {
                    this.slideInfo = results.groups[0].items[0]
                    this.totalSearchMatches = results.totalMatches
                } else {
                    this.error = "The slide cannot be found"
                }
            },
            error => this.error = "The server returned an error: " + error
        );
    }

    // private resetTimer() {
    //     if (this.currentInterval) {
    //         clearInterval(this.currentInterval);
    //         this.currentInterval = null;
    //     }
    // }
    //
    // private startTimer() {
    //     this.resetTimer();
    //     let interval = +this.interval;
    //     if (!isNaN(interval) && interval > 0) {
    //         this.currentInterval = setInterval(() => {
    //             let nInterval = +this.interval;
    //             if (this.isPlaying && !isNaN(this.interval) && nInterval > 0 && this.slides.length) {
    //                 this.next();
    //             } else {
    //                 this.pause();
    //             }
    //         }, interval);
    //     }
    // }

    slideshowLinkParameters(item: SearchItem, imageIndex: number) {
        let properties = this._searchRequestBuilder.toLinkParametersObject(this.searchRequest)
        properties['i'] = imageIndex
        return properties
    }

}
