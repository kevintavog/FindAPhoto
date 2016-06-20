import { Component, Input } from '@angular/core'
import { SearchCategory, SearchCategoryDetail } from './search-results'


@Component({
  selector: 'category-details-tree-view',
  template: `
    <div *ngFor="let scd of details" >
        <label>
            <input type="checkbox" value="{{scd.value}}" [(ngModel)]="scd.selected" >
            {{ getDisplayValue(scd) }} ({{scd.count}})
        </label>
        <div *ngIf="hasDetails(scd)" >

            <div *ngIf="scd.isOpen != true">
                <span (click)="scd.isOpen=true">&#x025BD;</span>
                <label>
                    <input type="checkbox" value="{{scd.details[0].value}}" [(ngModel)]="scd.details[0].selected" >
                    {{ getDisplayValue(scd.details[0]) }} ({{scd.details[0].count}})
                </label>
            </div>

            <div *ngIf="scd.isOpen == true">
                <span (click)="scd.isOpen=false">&#x025B3;</span>
            </div>
            <ul>
                <category-details-tree-view *ngIf="scd.isOpen == true" [details]=scd.details [field]=scd.field ></category-details-tree-view>
            </ul>
        </div>
    </div>
  `,
  directives: [CategoryDetailsTreeView]
})

export class CategoryDetailsTreeView {
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
}


@Component({
  selector: 'category-tree-view',
  template: `
    <div>{{caption}}
        <category-details-tree-view [details]=category.details [field]=category.field ></category-details-tree-view>
    </div>
  `,
  directives: [CategoryDetailsTreeView]
})

export class CategoryTreeView {
    @Input() caption: any
    @Input() category: any
}
