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

            _navigationProvider.updateSearchCallback = () => { console.log('via updateSearchCallBack'); this.internalSearch(true) };
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
// console.log('updating URL')
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

    internalSearch(updateUrl: boolean) {
        let result = this.generateDrilldown();

console.log('internalSearch(%s)', updateUrl)

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

        // On a refresh with drilldown, we need to select the proper categories
        if (this.drilldownFromUrl) {
            selectedCategories = this.drilldownToCategories(this.drilldownFromUrl);
            this.drilldownFromUrl = null;
        }

        selectedCategories.forEach((value, key) => {
            this.selectSavedCategories(key.split('~'), value);
        });

        if (context['updateUrl']) { this.updateUrl(); }
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

    removeSelectedCategory(scd: SearchCategoryDetail) {
        scd.selected = false;
        this.internalSearch(true);
    }

    selectSavedCategories(categoryPath: string[], valueArray: string[]) {
        if (this._searchResultsProvider.searchResults === undefined 
            || this._searchResultsProvider.searchResults.categories === undefined) {
            console.log('No categories in search results, cannot do anything...')
            return;
        }

        // valueArray item can be a path, separated by ~
        var valuePath: string[][] = [];
        valueArray.forEach( val => {
            var pathArray = val.split("~");
            valuePath.push(pathArray);
        });

        for (let category of this._searchResultsProvider.searchResults.categories) {
            if (category.field === categoryPath[0] && category.details !== undefined) {
                valuePath.forEach( vp => {
                    this.selectSavedCategoryChildren(category.details, categoryPath.slice(1), vp, '');
                });
            }
        }
    }

    selectSavedCategoryChildren(details: SearchCategoryDetail[], childPath: string[], valuePath: string[], displayPathPrefix: string) {
        for (let d of details) {
            if (valuePath[0] === d.value) {
                d.selected = true;

                if (childPath.length == 0) {
                    d.displayPath = displayPathPrefix + valuePath[0]
                    this.selectedCategories.push(d);
                } else {
                    if (childPath[0] === d.field) {
                        this.selectSavedCategoryChildren(
                            d.details, 
                            childPath.slice(1), 
                            valuePath.slice(1), 
                            displayPathPrefix + valuePath[0] + '/');
                        return;
                    }

                    console.log('Unable to find child path: %s', childPath[0])
                }
            }
        }
    }

    saveSelectedCategories(
            parent: SearchCategoryDetail, 
            selectionPrefix: string, 
            valuePrefix: string, 
            selectedCategories: Map<string, string[]>): boolean {

        if (!parent.details) { return false; }

        let hasSelection = false;
        for (let scd of parent.details) {
            if (scd.selected) {
                hasSelection = true;
                let key = selectionPrefix + parent.field;
                let value = valuePrefix + scd.value;

                let isChildSelected = this.saveSelectedCategories(scd, key + '~', value + '~', selectedCategories)
                if (!isChildSelected) {
                    if (selectedCategories.has(key)) {
                        selectedCategories.get(key).push(value);
                    } else {
                        selectedCategories.set(key, [value]);
                    }
                }

            } else {
                this.saveSelectedCategories(scd, '', '', selectedCategories);
            }
        }

        return hasSelection;
    }

    // Generate the drilldown from selected categories. There are two formats, basically:
    //  1) keywords: <category name>:val1,val2
    //  2) heirarchecal (dates, locations): <parent name>~<child name>:<parent value>~<child value>
    // These are separated by '_'.
    //      Example: countryName~stateName~cityName:Canada~British Columbia~Vancouver_keywords:soccer,flower
    generateDrilldown() {
        let selectedCategories = new Map<string, string[]>();
        if (this._searchResultsProvider.searchResults != null && this._searchResultsProvider.searchResults.categories != null) {
            for (let cat of this._searchResultsProvider.searchResults.categories) {

                let scd = new SearchCategoryDetail();
                scd.selected = false;
                scd.details = cat.details;
                scd.field = cat.field;

                this.saveSelectedCategories(scd, '', '', selectedCategories);
            }
        }

        let url = this.categoriesToDrilldown(selectedCategories);
        return { drilldown: url, categories: selectedCategories };
    }

    categoriesToDrilldown(selectedCategories: Map<string, string[]>): string {
        let drilldown = '';
        selectedCategories.forEach((value, key) => {
            if (drilldown.length > 0) {
                drilldown += '_';
            }

            if (key.includes('~')) {
                drilldown += key + ':' + value;
            } else {
                drilldown += key + ':' + value.join(',');
            }
        });

        return drilldown;
    }

    drilldownToCategories(drilldown: string): Map<string, string[]> {
        let selectedCategories = new Map<string, string[]>();
        if (!drilldown) {
            return selectedCategories;
        }
        if (!this._searchResultsProvider.searchResults || !this._searchResultsProvider.searchResults.categories) {
            console.log('There are no categories in the search results - unable to select any categories: "%s"', drilldown);
            return selectedCategories;
        }

        for (let keyAndValue of drilldown.split('_')) {
            let tokens = keyAndValue.split(':')
            if (tokens.length != 2) {
                console.log('Ignoring unexpected drilldown key/value: "%s", cannot parse it', keyAndValue)
            } else {
                let categoryName = tokens[0]
                for (let value of tokens[1].split(',')) {
                    if (selectedCategories.has(categoryName)) {
                        selectedCategories.get(categoryName).push(value);
                    } else {
                        selectedCategories.set(categoryName, [value]);
                    }
                }
            }
        }

        return selectedCategories;
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
