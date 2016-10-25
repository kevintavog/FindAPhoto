import { Component } from '@angular/core';

import { NavigationProvider } from '../../providers/navigation.provider';
import { SearchResultsProvider } from '../../providers/search-results.provider';

@Component({
    selector: 'app-alerts',
    templateUrl: './alerts.component.html',
    styleUrls: ['./alerts.component.css']
})
export class AlertsComponent {

    constructor(
        private searchResultsProvider: SearchResultsProvider,
        private navigationProvider: NavigationProvider) { }
}
