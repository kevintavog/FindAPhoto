import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, NavigationExtras, Router } from '@angular/router';
import { Title } from '@angular/platform-browser';


import { FieldAndCount, FieldValue, SearchCategory } from '../../models/search-results';

import { SearchRequestBuilder } from '../../models/search.request.builder';

import { DataDisplayer } from '../../providers/data-displayer';
import { FieldValueProvider } from '../../providers/field-values.provider'
import { NavigationProvider } from '../../providers/navigation.provider';
import { SearchResultsProvider } from '../../providers/search-results.provider';

import { SearchService } from '../../services/search.service';


@Component({
    selector: 'app-field-values',
    templateUrl: './field-values.component.html',
    styleUrls: ['./field-values.component.css']
})

export class FieldValuesComponent implements OnInit {

    currentSearch: string;

    dateCategories: SearchCategory;
    locationCategories: SearchCategory;

    maxValueCount = 400;
    apertureField = 'Aperture';
    cameraMakeField = 'Camera Make';
    cameraModelField = 'Camera Model';
    cityNameField = 'City';
    countryNameField = 'Country';
    displayNameField = 'Display name';
    durationSecondsField = 'Duration seconds';
    exposureTimeField = 'Exposure time';
    focalLengthField = 'Focal length';
    hierarchicalNameField = 'Hierarchical name'
    isoField = 'ISO'
    keywordsField = 'Keywords';
    lensModelField = 'Lens model';
    placeNameField = 'Place name';
    siteNameField = 'Site name';
    stateNameField = 'State name';
    tagsField = 'Tags';

    constructor(
        private router: Router,
        private route: ActivatedRoute,
        private searchResultsProvider: SearchResultsProvider,
        private searchRequestBuilder: SearchRequestBuilder,
        private navigationProvider: NavigationProvider,
        private displayer: DataDisplayer,
        private titleService: Title,
        private searchService: SearchService,
        private fieldValuesProvider: FieldValueProvider) { }

    ngOnInit() {
        this.titleService.setTitle('Field values - FindAPhoto');
        this.searchResultsProvider.initializeRequest(SearchResultsProvider.QueryProperties, 's');

        this.currentSearch = this.searchRequestBuilder.toReadableString(this.searchResultsProvider.searchRequest);

        this.route.queryParams.subscribe(params => {
            if ('q' in params || 't' in params) {
                this.startSearch(false);
            }
        });
    }

    userSearch() {
        this.searchResultsProvider.searchRequest.searchType = 's';
        this.startSearch(true);
    }


    datesFieldName() {
        if (this.dateCategories) {
            return 'Dates  (' + this.countCategories(this.dateCategories) + ')';
        }

        return 'Dates';
    }

    locationFieldName() {
        if (this.locationCategories) {
            return 'Location  (' + this.countCategories(this.locationCategories) + ')';
        }

        return 'Location';
    }

    countCategories(category: SearchCategory) {
        let count = 0;
        for (let c of category.details) {
            count += c.count;
        }

        return count        
    }

    startSearch(updateUrl: boolean) {
        this.currentSearch = this.searchRequestBuilder.toReadableString(this.searchResultsProvider.searchRequest);

        if (updateUrl) {
            let params = this.searchRequestBuilder.toLinkParametersObject(this.searchResultsProvider.searchRequest);
            let navigationExtras: NavigationExtras = { queryParams: params };

            // If the params are the same, navigating won't change anything, so fall through to the search invocation
            if (!this.navigationProvider.hasSameQueryParams(params)) {
                this.router.navigate( ['fieldvalues'], navigationExtras);
                return;
            }
        }

        this.searchResultsProvider.search(null);

        this.getRegularFields();
        this.getLocationsAndDates();
    }

    getRegularFields() {
        if (this.searchResultsProvider.searchRequest.searchType === 's') {
            this.searchService.indexFieldValues(
                    ['aperture', 'cameramake', 'cameramodel', 'cityname', 'countryname', 'displayname', 
                      'durationseconds', 'exposuretimestring', 'focallengthmm', 'hierarchicalname', 
                      'iso', 'keywords', 'lensmodel', 'placename', 'sitename', 'statename', 'tags'],
                    this.searchResultsProvider.searchRequest.searchText,
                    this.searchResultsProvider.searchRequest.drilldown,
                    this.maxValueCount).subscribe(
                        results => {
                            for (let fv of <[FieldValue]>results['fields']) {
                                switch (fv.name) {
                                    case 'aperture':
                                        this.fieldValuesProvider.fieldData.set(this.apertureField, fv.values);
                                        break;
                                    case 'cameramake':
                                        this.fieldValuesProvider.fieldData.set(this.cameraMakeField, fv.values);
                                        break;
                                    case 'cameramodel':
                                        this.fieldValuesProvider.fieldData.set(this.cameraModelField, fv.values);
                                        break;
                                    case 'cityname':
                                        this.fieldValuesProvider.fieldData.set(this.cityNameField, fv.values);
                                        break;
                                    case 'countryname':
                                        this.fieldValuesProvider.fieldData.set(this.countryNameField, fv.values);
                                        break;
                                    case 'displayname':
                                        this.fieldValuesProvider.fieldData.set(this.displayNameField, fv.values);
                                        break;
                                    case 'durationseconds':
                                        this.fieldValuesProvider.fieldData.set(this.durationSecondsField, fv.values);
                                        break;
                                    case 'exposuretimestring':
                                        this.fieldValuesProvider.fieldData.set(this.exposureTimeField, fv.values);
                                        break;
                                    case 'focallengthmm':
                                        this.fieldValuesProvider.fieldData.set(this.focalLengthField, fv.values);
                                        break;
                                    case 'hierarchicalname':
                                        this.fieldValuesProvider.fieldData.set(this.hierarchicalNameField, fv.values);
                                        break;
                                    case 'iso':
                                        this.fieldValuesProvider.fieldData.set(this.isoField, fv.values);
                                        break;
                                    case 'keywords':
                                        this.fieldValuesProvider.fieldData.set(this.keywordsField, fv.values);
                                        break;
                                    case 'lensmodel':
                                        this.fieldValuesProvider.fieldData.set(this.lensModelField, fv.values);
                                        break;
                                    case 'placename':
                                        this.fieldValuesProvider.fieldData.set(this.placeNameField, fv.values);
                                        break;
                                    case 'sitename':
                                        this.fieldValuesProvider.fieldData.set(this.siteNameField, fv.values);
                                        break;
                                    case 'statename':
                                        this.fieldValuesProvider.fieldData.set(this.stateNameField, fv.values);
                                        break;
                                    case 'tags':
                                        this.fieldValuesProvider.fieldData.set(this.tagsField, fv.values);
                                        break;

                                    default:
                                        console.log('Unhandled field: %o', fv.name);
                                }
                            }
                        },
                        error => {
                            console.log('keywords and tags error: %s', error);
                        });
        } else {
console.log('unsupported search type: %o', this.searchResultsProvider.searchRequest.searchType)
        }
    }

    getLocationsAndDates() {
        if (this.searchResultsProvider.searchRequest.searchType === 's') {
            this.searchService.searchByText(
                this.searchResultsProvider.searchRequest.searchText,
                '',
                1,
                1,
                this.searchResultsProvider.searchRequest.drilldown).subscribe(
                    results => {
                        for (let category of results.categories) {
                            if (category.field === 'dateYear') { this.dateCategories = category; }
                            if (category.field === 'countryName') { this.locationCategories = category; }
                        }
                    },
                    error => {
                        console.log('location and dates error: %s', error);
                    });
        } else {
console.log('unsupported search type: %o', this.searchResultsProvider.searchRequest.searchType)
        }
    }
}
