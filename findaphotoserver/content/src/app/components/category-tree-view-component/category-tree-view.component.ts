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
                <input type="checkbox" value="{{scd.value}}" [(ngModel)]="scd.selected" >
                    <div class="details-tree-text" (click)="scd.isOpen=!scd.isOpen">
                        {{ getDisplayValue(scd) }} ({{scd.count}})
                    </div>

            <app-category-details-tree-view *ngIf="scd.isOpen == true" [details]=scd.details [field]=scd.field >
            </app-category-details-tree-view>
        </li>
    </ul>
  `,
})

export class CategoryDetailsTreeViewComponent {
    @Input() field: string
    @Input() details: SearchCategoryDetail[]

    getDisplayValue(scd: SearchCategoryDetail) : string {
        if (scd.displayValue != undefined && scd.displayValue != null) {
            return scd.displayValue
        }
        return scd.value
    }

    hasDetails(scd: SearchCategoryDetail) : boolean {
        return scd != undefined && scd.details != null && scd.details.length > 0
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
