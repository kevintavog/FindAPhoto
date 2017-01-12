import { Injectable } from '@angular/core';

import { FieldValue } from '../models/search-results';
import { SearchService } from '../services/search.service';

@Injectable()
export class FieldsProvider {
    initialized: boolean = false;
    fields: [string];

    serverError: string;
    nameWithValues: string;
    values: string[];

    constructor(private searchService: SearchService) {
    }

    initialize() {
        this.nameWithValues = null;
        this.values = null;
        if (this.initialized) { return; }

        this.searchService.indexFields().subscribe(
            fieldResults => {
                this.fields = fieldResults.fields;
                this.initialized = true;
            },
            error => {
                this.serverError = error;
            }
        );
    }

    refreshFieldValues(searchText: string, drilldown: string) {
        if (this.nameWithValues) {
            this.getValuesForFieldWithSearch(this.nameWithValues, searchText, drilldown);
        }
    }

    getValuesForField(fieldName: string) {
        this.values = null;
        this.serverError = null;
        this.nameWithValues = fieldName;

        this.searchService.indexFieldValues([fieldName], null, null, null).subscribe(
            results => {
                for (let fv of <[FieldValue]>results['fields']) {
                    if (fv.name === fieldName) {
                        this.values = fv.values.map( (fc) => {
                            return fc.value;
                        });
                        break;
                    }
                }
            },
            error => {
                this.serverError = error;
            }
        );
    }

    getValuesForFieldWithSearch(fieldName: string, searchText: string, drilldown: string) {
        this.values = null;
        this.serverError = null;
        this.nameWithValues = fieldName;

        this.searchService.indexFieldValues([fieldName], searchText, drilldown, null).subscribe(
            results => {;
                for (let fv of <[FieldValue]>results['fields']) {
                    if (fv.name === fieldName) {
                        this.values = fv.values.map( (fc) => {
                            return fc.value;
                        });
                        break;
                    }
                }
            },
            error => {
                this.serverError = error;
            }
        );
    }

    getValuesForFieldByDay(fieldName: string, month: number, day: number, drilldown: string) {
        this.values = null;
        this.serverError = null;
        this.nameWithValues = fieldName;

        this.searchService.indexFieldValuesByDay([fieldName], month, day, drilldown, null).subscribe(
            results => {
                for (let fv of <[FieldValue]>results['fields']) {
                    if (fv.name === fieldName) {
                        this.values = fv.values.map( (fc) => {
                            return fc.value;
                        });
                        break;
                    }
                }
            },
            error => {
                this.serverError = error;
            }
        );
    }
}
