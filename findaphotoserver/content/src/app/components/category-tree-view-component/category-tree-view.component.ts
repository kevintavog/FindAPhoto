import { Component, Input } from '@angular/core';
import { SearchCategory, SearchCategoryDetail } from '../../models/search-results'

@Component({
  selector: 'app-category-details-tree-view',
  styles: [`
    .details-tree-ul {
        margin-left:5px;
    }
    .details-tree-text {
        display:inline;
        cursor:default;
    }
  `],
  template: `
    <ul class="c-tree details-tree-ul" >
        <li *ngFor="let scd of details" [ngClass]="getClassName(scd)" >
                <input type="checkbox" value="{{scd.value}}" [(ngModel)]="scd.selected" (change)="onSelectionChange(scd)" >
                    <div class="details-tree-text" (click)="toggleOpen(scd)">
                        {{ getDisplayValue(scd) }} ({{scd.count}})
                    </div>

            <app-category-details-tree-view *ngIf="scd.isOpen == true" 
                [details]=scd.details [field]=scd.field [parentComponent]=this [parentDetail]=scd >
            </app-category-details-tree-view>
        </li>
    </ul>
  `,
})

export class CategoryDetailsTreeViewComponent {
    @Input() field: string;
    @Input() details: SearchCategoryDetail[];
    @Input() parentComponent: CategoryDetailsTreeViewComponent;
    @Input() parentDetail: SearchCategoryDetail;

    getDisplayValue(scd: SearchCategoryDetail) : string {
        if (scd.displayValue != undefined && scd.displayValue != null) {
            return scd.displayValue
        }
        return scd.value
    }

    hasDetails(scd: SearchCategoryDetail) : boolean {
        return scd != undefined && scd.details != null && scd.details.length > 0
    }

    toggleOpen(scd: SearchCategoryDetail) {
        scd.isOpen = !scd.isOpen;
    }

    onSelectionChange(scd: SearchCategoryDetail) {
        let beingSelected = !scd.selected
        console.log('onSelectionChange for %s (parent: %o), %o', scd.value, this.parentDetail, beingSelected);
        if (beingSelected) {
            this.selectParents(this.parentComponent, this.parentDetail);
        } else {
            this.deselectChildren(this.details);
        }
    }

    selectParents(component: CategoryDetailsTreeViewComponent, detail: SearchCategoryDetail) {
        if (component && detail) {
            detail.selected = true;
            this.selectParents(component.parentComponent, component.parentDetail);
        }
    }

    deselectChildren(children: SearchCategoryDetail[]) {
        if (children) {
            for (let child of children) {
                child.selected = false;
                this.deselectChildren(child.details);
            }
        }
    }

    getClassName(scd: SearchCategoryDetail) : string {
        if (this.hasDetails(scd)) {
            if (scd.isOpen) {
                return 'c-tree__item c-tree__item--expanded';
            } else {
                return 'c-tree__item c-tree__item--expandable';
            }
        }
        return 'c-tree__item';
    }
}


@Component({
  selector: 'app-category-tree-view',
  template: `
    <div>{{caption}}
        <app-category-details-tree-view [details]=category.details [field]=category.field ></app-category-details-tree-view>
    </div>
  `,
})

export class CategoryTreeViewComponent {
    @Input() caption: any
    @Input() category: any
}
