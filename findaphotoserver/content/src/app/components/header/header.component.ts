import { Component } from '@angular/core';

import { NavigationProvider } from '../../providers/navigation.provider';

@Component({
    selector: 'app-header',
    templateUrl: './header.component.html',
    styleUrls: ['./header.component.css']
})

export class HeaderComponent {
    menuOpen: boolean;

    constructor(protected navigationProvider: NavigationProvider) { }

    toggleMenu() {
        this.menuOpen = !this.menuOpen;
    }

    autoCloseForDropdownMenu(event) {
        if (!event.target.closest('.drop-down-menu-container')) {
            this.menuOpen = false;
        }
    }
}
