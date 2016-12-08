import { OnDestroy } from '@angular/core';
import { ActivatedRoute, NavigationExtras, Router } from '@angular/router';
import { Location } from '@angular/common';


import { SortType } from '../../models/search-request';
import { SearchRequestBuilder } from '../../models/search.request.builder';
import { SearchCategoryDetail, SearchItem } from '../../models/search-results';
import { UIState } from '../../models/ui-state';

import { FieldsProvider } from '../../providers/fields.provider';
import { NavigationProvider } from '../../providers/navigation.provider';
import { SearchResultsProvider } from '../../providers/search-results.provider';


export abstract class BaseSearchComponent implements OnDestroy {

    public DatesCaption: string = 'Dates:';
    public KeywordsCaption: string = 'Keywords:';
    public LocationsCaption: string = 'Locations:';

    uiState = new UIState();

    selectedCategories: SearchCategoryDetail[] = [];

    pageMessage: string;
    pageSubMessage: string;
    typeLeftButtonText: string;
    typeRightButtonText: string;
    typeLeftButtonClass: string;
    typeRightButtonClass: string;

    drilldownFromUrl: string;

    constructor(
        private _pageRoute: string,
        protected _router: Router,
        protected _route: ActivatedRoute,
        protected _location: Location,
        protected _searchRequestBuilder: SearchRequestBuilder,
        protected _searchResultsProvider: SearchResultsProvider,
        protected _navigationProvider: NavigationProvider,
        protected _fieldsProvider: FieldsProvider) {
            this.uiState.showGroup = true;
            this.uiState.sortMenuDisplayText = 'Date: Newest';

            _searchResultsProvider.searchStartingCallback = (context) => this.searchStartingCallback(context);
            _searchResultsProvider.searchCompletedCallback = (context) => this.searchCompletedCallback(context);

            _navigationProvider.updateSearchCallback = () => this.internalSearch(true);
    }

    ngOnDestroy() {
        this._searchResultsProvider.searchStartingCallback = null;
        this._searchResultsProvider.searchCompletedCallback = null;
        this._navigationProvider.updateSearchCallback = null;
    }


    singleItemSearchLinkParameters(item: SearchItem, imageIndex: number, groupIndex: number) {
        let properties = this._searchRequestBuilder.toLinkParametersObject(this._searchResultsProvider.searchRequest);
        properties['id'] = item.id;
        properties['i'] = imageIndex + groupIndex + this._searchResultsProvider.searchRequest.first;
        return properties;
    }

    updateUrl() {
        let params = this._searchRequestBuilder.toLinkParametersObject(this._searchResultsProvider.searchRequest);
        let drilldown = this.generateDrilldown().drilldown;
        if (drilldown.length > 0) {
            params['drilldown'] = drilldown;
        }

        if (this._searchResultsProvider.currentPage > 1) {
            params['p'] = this._searchResultsProvider.currentPage;
        }

        let navigationExtras: NavigationExtras = { queryParams: params };
        this._router.navigate( [this._pageRoute], navigationExtras);
    }

    sortByDateNewest() { this.sortBy(SortType.DateNewest, 'Date: Newest'); }
    sortByDateOldest() { this.sortBy(SortType.DateOldest, 'Date: Oldest'); }
    sortByLocationAscending() { this.sortBy(SortType.LocationAZ, 'Location: A-Z'); }
    sortByLocationDescending() { this.sortBy(SortType.LocationZA, 'Location: Z-A'); }
    sortByFolderAscending() { this.sortBy(SortType.FolderAZ, 'Folder: A-Z'); }
    sortByFolderDescending() { this.sortBy(SortType.FolderZA, 'Folder: Z-A'); }
    sortBy(sortType: string, sortDisplayName: string) {
console.log('sort by %o', sortType);
        this.uiState.sortMenuDisplayText = sortDisplayName;
    }

    searchStartingCallback(context: Map<string, any>) {
        if (!context) { return; }

        this.uiState.sortMenuShowing = false;
        this.pageMessage = undefined;
        this.selectedCategories = [];
    }

    searchCompletedCallback(context: Map<string, any>) {
        this.processSearchResults();

        if (!context) { return; }

        let selectedCategories: Map<string, string[]> = context['selectedCategories'];
        selectedCategories.forEach((value, key) => {
            this.selectSavedCategories(key.split('/'), value);
        });

        // On a refresh with drilldown, we need to select the proper categories
        if (this.drilldownFromUrl) {
            this.selectCategoriesFromDrilldown(this.drilldownFromUrl);
            this.drilldownFromUrl = null;
        }

        if (context['updateUrl']) { this.updateUrl(); }
    }

    internalSearch(updateUrl: boolean) {
        let result = this.generateDrilldown();

        if (this.drilldownFromUrl) {
            this._searchResultsProvider.searchRequest.drilldown = this.drilldownFromUrl;
        } else {
            this._searchResultsProvider.searchRequest.drilldown = result.drilldown;
        }

        let context = new Map<string, any>();
        context['updateUrl'] = updateUrl;
        context['selectedCategories'] = result.categories;

        this._searchResultsProvider.search(context);
    }

    removeSelectedCategory(scd: SearchCategoryDetail) {
        scd.selected = false;
        this.internalSearch(true);
    }

    selectSavedCategories(categoryPath: string[], valueArray: string[]) {
        if (this._searchResultsProvider.searchResults === undefined 
            || this._searchResultsProvider.searchResults.categories === undefined) {
            return;
        }

        for (let category of this._searchResultsProvider.searchResults.categories) {
            if (category.field === categoryPath[0] && category.details !== undefined) {
                this.selectSavedCategoryChildren(category.details, categoryPath.slice(1), valueArray);
            }
        }
    }

    selectSavedCategoryChildren(details: SearchCategoryDetail[], childPath: string[], valueArray: string[]) {
        if (childPath.length === 0) {
            for (let d of details) {
                if (valueArray.indexOf(d.value) >= 0) {
                    d.selected = true;
                    this.selectedCategories.push(d);
                }
            }
        } else {
            for (let d of details) {
                if (childPath[0] === d.field) {
                    this.selectSavedCategoryChildren(d.details, childPath.slice(1), valueArray);
                }
            }
        }
    }

    saveSelectedCategories(field: string, details: SearchCategoryDetail[], selectedCategories: Map<string, string[]>) {
        for (let scd of details) {
            if (scd.selected) {
                if (selectedCategories.has(field)) {
                    selectedCategories.get(field).push(scd.value);
                } else {
                    selectedCategories.set(field, [scd.value]);
                }
            }

            if (scd.details !== undefined) {
                this.saveSelectedCategories(field + '/' + scd.field, scd.details, selectedCategories);
            }
        }
    }

    // Generate the drilldown from selected categories. The format is 'category name':val1,val2' - each category is
    // separated by '_'. For heirarchecal categories, the 'category name' is the selected value
    //      Example: "countryName:Canada_stateName:Washington,Ile-de-France_keywords:trip,flower"
    generateDrilldown() {
        let selectedCategories = new Map<string, string[]>();
        if (this._searchResultsProvider.searchResults != null && this._searchResultsProvider.searchResults.categories != null) {
            for (let cat of this._searchResultsProvider.searchResults.categories) {
               this.saveSelectedCategories(cat.field, cat.details, selectedCategories);
            }
        }
        return { drilldown: this.generateDrilldownWithCategories(selectedCategories), categories: selectedCategories };
    }

    generateDrilldownWithCategories(selectedCategories: Map<string, string[]>): string {
        let drilldown = '';
        selectedCategories.forEach((value, key) => {
            let categories = key.split('/');
            if (drilldown.length > 0) {
                drilldown += '_';
            }

            drilldown += categories[categories.length - 1] + ':' + value.join(',');
        });

        return drilldown;
    }

    populateFieldValues(fieldName: string) {
        if (this._searchResultsProvider.searchRequest.searchType === 'd') {
            this._fieldsProvider.getValuesForFieldByDay(
                fieldName,
                this._searchResultsProvider.searchRequest.month,
                this._searchResultsProvider.searchRequest.day,
                this._searchResultsProvider.searchRequest.drilldown);
        } else {
            this._fieldsProvider.getValuesForFieldWithSearch(
                fieldName,
                this._searchResultsProvider.searchRequest.searchText,
                this._searchResultsProvider.searchRequest.drilldown);
        }
    }

    // Generate the selected categories from the drilldown. The format is 'category name':val1,val2' - each category is
    // separated by '_'. For heirarchecal categories, the 'category name' is the selected value
    //      Example: "countryName:Canada_stateName:Washington,Ile-de-France_keywords:trip,flower"
    selectCategoriesFromDrilldown(drilldown: string) {
        if (!drilldown) { return; }
        if (!this._searchResultsProvider.searchResults || !this._searchResultsProvider.searchResults.categories) {
            console.log('There are no categories in the search results - unable to select any categories: "%s"', drilldown);
        }

        for (let categoryValue of drilldown.split('_')) {
            let tokens = categoryValue.split(':')
            if (tokens.length != 2) {
                console.log('Ignoring unexpected category value: "%s", cannot parse it', categoryValue)
            } else {
                let categoryName = tokens[0]
                for (let value of tokens[1].split(',')) {
                    let valueSet = false;
                    for (let cat of this._searchResultsProvider.searchResults.categories) {
                        valueSet = this.selectCategoryAndValue(categoryName, value, cat.field, cat.details)
                        if (valueSet) {
                            break;
                        } 
                    }

                    if (!valueSet) {
                        console.log('Didn\'t find value "%s", cannot set it for the category "%s"', value, categoryName);
                    }
                }
            }
        }
    }

    // Recursively walk all categories/sub-categories to find a match for a category and value - select it
    selectCategoryAndValue(categoryName: string, value: string, field: string, details: SearchCategoryDetail[]): boolean {
        if (categoryName == field) {
            for (let scd of details) {
                if (value == scd.value) {
                    scd.selected = true;
                    this.selectedCategories.push(scd);
                    return true;
                }
            }
        }

        // Unable to find it directly in the given category - perhaps it's a sub-detail?
        for (let scd of details) {
            if (scd.details != null) {
                if (this.selectCategoryAndValue(categoryName, value, scd.field, scd.details)) {
                    return true;
                }
            }
        } 

        return false;
    }

    logCategoryDetails(details: SearchCategoryDetail[], prefix: string) {
        if (details === undefined || details === null) { return; }
        for (let d of details) {
            console.log(prefix + d.value + '; ' + d.count + '; ' + d.field);
            this.logCategoryDetails(d.details, prefix + '  ');
        }
    }

    abstract processSearchResults(): void;
    typeLeftButton() {};
    typeRightButton() {};
}
