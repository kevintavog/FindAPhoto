import { Input, Output } from "@angular/core"
import { Router, ROUTER_DIRECTIVES, RouteParams } from '@angular/router-deprecated';
import { Location } from '@angular/common';

import { BaseComponent } from './base.component';
import { SearchService } from './search.service';
import { SearchRequestBuilder } from './search.request.builder';
import { SearchRequest, SortType } from './search-request';
import { SearchResults, SearchGroup, SearchItem, SearchCategory, SearchCategoryDetail } from './search-results';


export abstract class BaseSearchComponent extends BaseComponent {


    protected static QueryProperties: string = "id,city,keywords,imageName,createdDate,latitude,longitude,thumbUrl,slideUrl,warnings"
    public static ItemsPerPage: number = 30

    public DatesCaption: string = "Dates:"
    public KeywordsCaption: string = "Keywords:"
    public LocationsCaption: string = "Locations:"

    sortMenuShowing: boolean
    sortMenuDisplayText: string
    showSearch: boolean
    showResultCount: boolean
    showGroup: boolean
    showDistance: boolean
    @Output() @Input() showFilters: boolean = false

    locationError: string
    serverError: string
    searchRequest: SearchRequest;
    searchResults: SearchResults;
    currentPage: number;
    totalPages: number;

    pageMessage: string
    pageSubMessage: string
    typeLeftButtonText: string
    typeRightButtonText: string
    typeLeftButtonClass: string
    typeRightButtonClass: string

    extraProperties: string


    constructor(
        private _pageRoute: string,
        protected _router: Router,
        protected _routeParams: RouteParams,
        protected _location: Location,
        protected _searchService: SearchService,
        protected _searchRequestBuilder: SearchRequestBuilder) {
            super()
            this.showGroup = true
            this.sortMenuDisplayText = "Date: Newest"
        }


    initializeSearchRequest(searchType: string) {
        let queryProps = BaseSearchComponent.QueryProperties
        if (this.extraProperties != undefined) {
            queryProps += "," + this.extraProperties
        }
        this.searchRequest = this._searchRequestBuilder.createRequest(this._routeParams, BaseSearchComponent.ItemsPerPage, queryProps, searchType)
    }


    slideSearchLinkParameters(item: SearchItem, imageIndex: number, groupIndex: number) {
        let properties = this._searchRequestBuilder.toLinkParametersObject(this.searchRequest)
        properties['id'] = item.id
        properties['i'] = imageIndex + groupIndex + this.searchRequest.first
        return properties
    }

    updateUrl() {
        let drilldown = this.generateDrilldown()
        if (drilldown.length > 0) {
            drilldown = "&drilldown=" + drilldown
        }

        this._location.go(this._pageRoute, this._searchRequestBuilder.toSearchQueryParameters(this.searchRequest) + drilldown + "&p=" + this.currentPage)
    }

    currentPageNumber() {
        if (this.searchResults != undefined) {
            return 1 + Math.round(this.searchRequest.first / this.searchRequest.pageCount)
        }
        return null;
    }

    gotoPage(pageOneBased: number) {
        if (this.currentPage != pageOneBased) {
            this.searchRequest.first = 1 + (pageOneBased - 1) * BaseSearchComponent.ItemsPerPage
            this.internalSearch(true)
        }
    }

    firstPage() {
        if (this.currentPage > 1) {
            this.searchRequest.first = 1
            this.internalSearch(true)
        }
    }

    lastPage() {
        if (this.currentPage < this.totalPages) {
            this.searchRequest.first = (this.totalPages - 1) * BaseSearchComponent.ItemsPerPage
            this.internalSearch(true)
        }
    }

    previousPage() {
        if (this.currentPage > 1) {
            let zeroBasedPage = this.currentPage - 1
            this.searchRequest.first = 1 + ((zeroBasedPage - 1) * BaseSearchComponent.ItemsPerPage)
            this.internalSearch(true)
        }
    }

    nextPage() {
        if (this.currentPage < this.totalPages) {
            let zeroBasedPage = this.currentPage - 1
            this.searchRequest.first = 1 + ((zeroBasedPage + 1) * BaseSearchComponent.ItemsPerPage)
            this.internalSearch(true)
        }
    }

    home() {
        this._router.navigate( ['Search'] )
    }

    toggleSortMenu() {
        if (this.sortMenuShowing != true) {
            this.sortMenuShowing = true
        } else {
            this.sortMenuShowing = false
        }
    }

    searchToday() {
        this._router.navigate( ['ByDay'] )
    }

    searchNearby() {
        this.locationError = undefined

        if (window.navigator.geolocation) {
            window.navigator.geolocation.getCurrentPosition(
                (position: Position) => {
                    this._router.navigate( ['ByLocation', { lat:position.coords.latitude, lon:position.coords.longitude }] );
                },
                (error: PositionError) => {
                    this.locationError = "Unable to get location: " + error.message + " (" + error.code + ")"
                })
        }
    }

    sortByDateNewest() { this.sortBy(SortType.DateNewest, "Date: Newest") }
    sortByDateOldest() { this.sortBy(SortType.DateOldest, "Date: Oldest") }
    sortByLocationAscending() { this.sortBy(SortType.LocationAZ, "Location: A-Z") }
    sortByLocationDescending() { this.sortBy(SortType.LocationZA, "Location: Z-A") }
    sortByFolderAscending() { this.sortBy(SortType.FolderAZ, "Folder: A-Z") }
    sortByFolderDescending() { this.sortBy(SortType.FolderZA, "Folder: Z-A") }
    sortBy(sortType: string, sortDisplayName: string) {
        console.log("sort by %o", sortType)
        this.sortMenuDisplayText = sortDisplayName
    }

    internalSearch(updateUrl: boolean) {
        this.sortMenuShowing = false
        var selectedCategories = new Map<string,string[]>()
        if (this.searchResults != null) {
            for (var cat of this.searchResults.categories) {
               this.saveSelectedCategories(cat.field, cat.details, selectedCategories)
            }

            this.searchRequest.drilldown = this.generateDrilldown()
        }

        this.searchResults = undefined
        this.serverError = undefined
        this.pageMessage = undefined

        this._searchService.search(this.searchRequest).subscribe(
            results => {
                this.searchResults = results

                let resultIndex = 0
                for (var group of this.searchResults.groups) {
                    group.resultIndex = resultIndex
                    resultIndex += group.items.length
                }

                let pageCount = this.searchResults.totalMatches / BaseSearchComponent.ItemsPerPage
                this.totalPages = ((pageCount) | 0) + (pageCount > Math.floor(pageCount) ? 1 : 0)
                this.currentPage = 1 + (this.searchRequest.first / BaseSearchComponent.ItemsPerPage) | 0

                this.processSearchResults()

                var dates = this.categoryDate()
                if (dates != null) {
                    for (var detail of dates.details) {
                        if (detail.value.length == 8 && Number.isInteger(Number(detail.value))) {
                            let year = Number(detail.value.substring(0, 4))
                            let month = Number(detail.value.substring(4, 6))
                            let day = Number(detail.value.substring(6, 8))

                            detail.displayValue = new Date(year, month - 1, day).toLocaleDateString()
                        }
                    }
                }

                selectedCategories.forEach((value, key) => {
                    this.selectSavedCategories(key.split("/"), value)
                })

                if (updateUrl) { this.updateUrl() }
            },
            error => this.serverError = error
       );
    }

    selectSavedCategories(categoryPath:string[], valueArray:string[]) {
        if (this.searchResults == undefined || this.searchResults.categories == undefined) {
            return
        }

        for (let category of this.searchResults.categories) {
            if (category.field == categoryPath[0] && category.details != undefined) {
                this.selectSavedCategoryChildren(category.details, categoryPath.slice(1), valueArray)
            }
        }
    }

    selectSavedCategoryChildren(details:SearchCategoryDetail[], childPath:string[], valueArray:string[]) {
        if (childPath.length == 0) {
            for (let d of details) {
                if (valueArray.indexOf(d.value) >= 0) {
                    d.selected = true
                }
            }
        } else {
            for (let d of details) {
                if (childPath[0] == d.field) {
                    this.selectSavedCategoryChildren(d.details, childPath.slice(1), valueArray)
                }
            }
        }
    }

    saveSelectedCategories(field: string, details: SearchCategoryDetail[], selectedCategories: Map<string,string[]>) {
        for (var scd of details) {
            if (scd.selected) {
                if (selectedCategories.has(field)) {
                    selectedCategories.get(field).push(scd.value)
                } else {
                    selectedCategories.set(field, [scd.value])
                }
            }

            if (scd.details != undefined) {
                this.saveSelectedCategories(field + "/" + scd.field, scd.details, selectedCategories)
            }
        }
    }

    // Generate the drilldown from selected categories. The format is 'category name':val1,val2' - each category is
    // separated by '_'. For heirarchecal categories, the 'category name' is the selected value
    //      Example: "countryName:Canada_stateName:Washington,Ile-de-France_keywords:trip,flower"
    generateDrilldown() : string {
        var selectedCategories = new Map<string,string[]>()
        if (this.searchResults != null) {
            for (var cat of this.searchResults.categories) {
               this.saveSelectedCategories(cat.field, cat.details, selectedCategories)
            }
        }
        return this.generateDrilldownWithCategories(selectedCategories)
    }

    generateDrilldownWithCategories(selectedCategories: Map<string,string[]>) : string {
        var drilldown = ""
        selectedCategories.forEach((value, key) => {
            let categories = key.split("/")
            if (drilldown.length > 0) {
                drilldown += "_"
            }

            drilldown += categories[categories.length - 1] + ':' + value.join(",")
        })
        return drilldown
    }

    logCategoryDetails(details: SearchCategoryDetail[], prefix: string) {
        if (details == undefined || details == null) { return }
        for (var d of details) {
            console.log(prefix + d.value + "; " + d.count + "; " + d.field)
            this.logCategoryDetails(d.details, prefix + "  ")
        }
    }

    categoryDate() : SearchCategory {
        return this.categoryByField("date")
    }

    categoryKeywords() : SearchCategory {
        return this.categoryByField("keywords")
    }

    categoryPlacenames() : SearchCategory {
        return this.categoryByField("countryName")
    }

    categoryByField(field: string) : SearchCategory {
        for (var category of this.searchResults.categories) {
            if (category.field == field) { return category }
        }
        return null
    }

    toggleFilterPanel() {
        if (this.showFilters != true) {
            this.showFilters = true
        } else {
            this.showFilters = false
        }
    }

    abstract processSearchResults() : void
    typeLeftButton() {}
    typeRightButton() {}

    firstResult() {
        if (this.searchResults != undefined && this.searchResults.totalMatches > 0) {
            return this.searchResults.groups[0].items[0]
        }
        return undefined
    }
}
