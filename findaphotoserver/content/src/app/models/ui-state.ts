import { Input, Output } from '@angular/core';

export class UIState {
    sortMenuShowing: boolean;
    sortMenuDisplayText: string;

    showSearch: boolean;
    showResultCount: boolean;

    showGroup: boolean;
    showDistance: boolean;
    @Output() @Input() showFilters: boolean = false;


    toggleSortMenu() {
        if (this.sortMenuShowing !== true) {
            this.sortMenuShowing = true;
        } else {
            this.sortMenuShowing = false;
        }
    }


    toggleFilterPanel() {
        if (this.showFilters !== true) {
            this.showFilters = true;
        } else {
            this.showFilters = false;
        }
    }
}
