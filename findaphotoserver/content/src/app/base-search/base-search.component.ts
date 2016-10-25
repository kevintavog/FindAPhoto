import { Component, Input, Output } from '@angular/core';
import { ActivatedRoute, NavigationExtras, Router } from '@angular/router';
import { Location } from '@angular/common';


import { BaseComponent } from '../base/base.component';
import { SearchRequest, SortType } from '../models/search-request';
import { SearchRequestBuilder } from '../models/search.request.builder';
import { SearchCategory, SearchCategoryDetail, SearchGroup, SearchItem, SearchResults } from '../models/search-results';
import { UIState } from '../models/ui-state'

import { NavigationProvider } from '../providers/navigation.provider';
import { SearchResultsProvider } from '../providers/search-results.provider';


export abstract class BaseSearchComponent extends BaseComponent  {


    public DatesCaption: string = "Dates:"
    public KeywordsCaption: string = "Keywords:"
    public LocationsCaption: string = "Locations:"

    uiState = new UIState()


    pageMessage: string;
    pageSubMessage: string;
    typeLeftButtonText: string;
    typeRightButtonText: string;
    typeLeftButtonClass: string;
    typeRightButtonClass: string;

    extraProperties: string;

    constructor(
        private _pageRoute: string,
        protected _router: Router,
        protected _route: ActivatedRoute,
        protected _location: Location,
        protected _searchRequestBuilder: SearchRequestBuilder,
        protected _searchResultsProvider: SearchResultsProvider,
        protected _navigationProvider: NavigationProvider) {
            super()
            this.uiState.showGroup = true
            this.uiState.sortMenuDisplayText = "Date: Newest"

            _searchResultsProvider.searchStartingCallback = (context) => this.searchStartingCallback(context)
            _searchResultsProvider.searchCompletedCallback = (context) => this.searchCompletedCallback(context)

            _navigationProvider.updateSearchCallback = () => this.internalSearch(true)
    }

    initializeSearchRequest(searchType: string) {
        let queryProps = SearchResultsProvider.QueryProperties
        if (this.extraProperties != undefined) {
            queryProps += "," + this.extraProperties
        }

        this._navigationProvider.locationError = undefined
        this._searchResultsProvider.initializeRequest(queryProps, searchType);
    }

    singleItemSearchLinkParameters(item: SearchItem, imageIndex: number, groupIndex: number) {
        let properties = this._searchRequestBuilder.toLinkParametersObject(this._searchResultsProvider.searchRequest)
        properties['id'] = item.id
        properties['i'] = imageIndex + groupIndex + this._searchResultsProvider.searchRequest.first
        return properties
    }

    updateUrl() {
        let params = this._searchRequestBuilder.toLinkParametersObject(this._searchResultsProvider.searchRequest);
        let drilldown = this.generateDrilldown()
        if (drilldown.length > 0) {
            console.log("!!! handle drilldown...")
            params['drilldown'] = drilldown;
        }

        if (this._searchResultsProvider.currentPage > 1) {
            params['p'] = this._searchResultsProvider.currentPage;
        }

        let navigationExtras: NavigationExtras = { queryParams: params };
        this._router.navigate( [this._pageRoute], navigationExtras);
    }

    sortByDateNewest() { this.sortBy(SortType.DateNewest, "Date: Newest") }
    sortByDateOldest() { this.sortBy(SortType.DateOldest, "Date: Oldest") }
    sortByLocationAscending() { this.sortBy(SortType.LocationAZ, "Location: A-Z") }
    sortByLocationDescending() { this.sortBy(SortType.LocationZA, "Location: Z-A") }
    sortByFolderAscending() { this.sortBy(SortType.FolderAZ, "Folder: A-Z") }
    sortByFolderDescending() { this.sortBy(SortType.FolderZA, "Folder: Z-A") }
    sortBy(sortType: string, sortDisplayName: string) {
        console.log("sort by %o", sortType)
        this.uiState.sortMenuDisplayText = sortDisplayName
    }

    searchStartingCallback(context: Map<string,any>) {
        this.uiState.sortMenuShowing = false
        var selectedCategories = new Map<string,string[]>()
        if (this._searchResultsProvider.searchResults != null) {
            for (var cat of this._searchResultsProvider.searchResults.categories) {
               this.saveSelectedCategories(cat.field, cat.details, selectedCategories)
            }

            this._searchResultsProvider.searchRequest.drilldown = this.generateDrilldown()
        }

        this.pageMessage = undefined

        context['selectedCategories'] = selectedCategories
    }

    searchCompletedCallback(context: Map<string,any>) {
        this.processSearchResults()

        let selectedCategories: Map<string,string[]> = context['selectedCategories']
        selectedCategories.forEach((value, key) => {
            this.selectSavedCategories(key.split("/"), value)
        })

        if (context['updateUrl']) { this.updateUrl() }
    }

    internalSearch(updateUrl: boolean) {
        let context = new Map<string,any>()
        context['updateUrl'] = updateUrl
        this._searchResultsProvider.search(context)
    }

    selectSavedCategories(categoryPath:string[], valueArray:string[]) {
        if (this._searchResultsProvider.searchResults == undefined || this._searchResultsProvider.searchResults.categories == undefined) {
            return
        }

        for (let category of this._searchResultsProvider.searchResults.categories) {
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
        if (this._searchResultsProvider.searchResults != null) {
            for (var cat of this._searchResultsProvider.searchResults.categories) {
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

    abstract processSearchResults() : void
    typeLeftButton() {}
    typeRightButton() {}

}
