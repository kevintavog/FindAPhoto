import { Component, OnInit } from '@angular/core';

import { DataDisplayer } from '../../providers/data-displayer';
import { FieldsProvider } from '../../providers/fields.provider';
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

    duplicateCount: number;
    imageCount: number;
    paths: [PathAndDate]
    versionNumber: string;
    videoCount: number;
    warningCount: number;

    searchHints: SearchHints[]



    constructor(
            private searchService: SearchService,
            private displayer: DataDisplayer,
            protected navigationProvider: NavigationProvider,
            private fieldsProvider: FieldsProvider) {
      this.searchHints = [
        new SearchHints('All warnings', 'warnings:*'),
        new SearchHints('Items with a keyword of "trip"', 'keywords:trip'),
        new SearchHints('Everything since a date: October 4, 2016', 'date:>=20161004'),
        new SearchHints('Everything from 2016, January-December', 'date:2016*'),
        new SearchHints('Range, everything from 2015 and 2016', 'date:[2015* TO date:2016*]'),
        new SearchHints('All trips from 2015 and 2016', 'keywords:trip AND date:[2015* TO date:2016*]'),
        new SearchHints('The placename is more than 10 meters from the location', 'cachedlocationdistancemeters:>10'),
        new SearchHints('Everything from outside of Washington state', 'statename:* AND NOT statename:Washington'),
      ];
    }

    ngOnInit() {
      this.fieldsProvider.initialize();
      this.serverError = null;
      this.searchService.indexStats('duplicateCount,imageCount,paths,versionNumber,videoCount,warningCount').subscribe(
            results => {
              this.duplicateCount = results.duplicateCount;
              this.imageCount = results.imageCount;
              this.paths = results.paths;
              this.versionNumber = results.versionNumber;
              this.videoCount = results.videoCount;
              this.warningCount = results.warningCount;
            },
            error => {
                this.serverError = error;
            }
      );
    }

    hasPaths() {
      return this.paths != null && this.paths.length > 0;
    }
}
