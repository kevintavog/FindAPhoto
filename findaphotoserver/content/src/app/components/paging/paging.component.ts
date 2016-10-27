import { Component } from '@angular/core';

import { NavigationProvider } from '../../providers/navigation.provider';
import { SearchResultsProvider } from '../../providers/search-results.provider';

@Component({
    selector: 'app-paging',
    templateUrl: './paging.component.html',
    styleUrls: ['./paging.component.css']
})


export class PagingComponent {

    constructor(protected searchResultsProvider: SearchResultsProvider, protected navigationProvider: NavigationProvider) { }

    get currentPage(): number { return this.searchResultsProvider.currentPage; }
    get totalPages(): number { return this.searchResultsProvider.totalPages; }

    gotoPage(pageNumber: number) {
        this.navigationProvider.gotoPage(pageNumber);
    }
}
