import { Component } from '@angular/core';

import { NavigationProvider } from '../../providers/navigation.provider';

@Component({
    selector: 'app-header',
    templateUrl: './header.component.html',
    styleUrls: ['./header.component.css']
})
export class HeaderComponent {

    constructor(protected navigationProvider: NavigationProvider) { }

}
