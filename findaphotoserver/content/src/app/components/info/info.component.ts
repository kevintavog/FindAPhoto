import { Component, OnInit } from '@angular/core';

import { DataDisplayer } from '../../providers/data-displayer';
import { NavigationProvider } from '../../providers/navigation.provider';
import { SearchService } from '../../services/search.service';


export interface PathAndDate {
    path: string;
    lastIndexed: Date;
}

export class SearchHints {
    description: string;
    query: string;

    constructor(d: string, q: string) {
      this.description = d;
      this.query = q;
    }
}

@Component({
    selector: 'app-info',
    templateUrl: './info.component.html',
    styleUrls: ['./info.component.css']
})

export class InfoComponent implements OnInit {

    serverError: string;

    fields: [string]
    fieldNameWithValues: string;
    fieldNameServerError: string;
    fieldValues: [string];

    imageCount: number;
    paths: [PathAndDate]
    versionNumber: string;
    videoCount: number;
    warningCount: number;

    searchHints: SearchHints[]



    constructor(private searchService: SearchService, private displayer: DataDisplayer, protected navigationProvider: NavigationProvider) {
      this.searchHints = [
        new SearchHints('All warnings', 'warnings:*'),
        new SearchHints('Items with a keyword of "trip"', 'keywords:trip'),
        new SearchHints('Everything from 2016, January-December', 'date:2016*'),
        new SearchHints('Range, everything from 2015 and 2016', 'date:[2015* TO date:2016*]'),
        new SearchHints('All trips from 2015 and 2016', 'keywords:trip AND date:[2015* TO date:2016*]'),
        new SearchHints('The placename is more than 10 meters from the location', 'cachedlocationdistancemeters:>10'),
      ];
    }

    ngOnInit() {
      this.serverError = null;
      this.searchService.indexStats('imageCount,paths,versionNumber,videoCount,warningCount').subscribe(
            results => {
              this.imageCount = results.imageCount;
              this.paths = results.paths;
              this.versionNumber = results.versionNumber;
              this.videoCount = results.videoCount;
              this.warningCount = results.warningCount;

              this.searchService.indexFields().subscribe(
                  fieldResults => {
                      this.fields = fieldResults.fields;
                  },
                  error => {
                    this.serverError = error;
                  }
              );
            },
            error => {
                this.serverError = error;
            }
      );
    }

    hasPaths() {
      return this.paths != null && this.paths.length > 0;
    }

    getValuesForField(fieldName : string) {
        this.fieldValues = null;
        this.fieldNameServerError = null;
        this.fieldNameWithValues = fieldName;

        this.searchService.indexFieldValues(fieldName).subscribe(
            results => {
                this.fieldValues = results.values.values;
            },
            error => {
              this.fieldNameServerError = error;
            }
        );

    }
}
