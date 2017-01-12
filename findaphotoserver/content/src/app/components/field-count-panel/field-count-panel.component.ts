import { Component, Input } from '@angular/core';

import { FieldValueProvider } from '../../providers/field-values.provider'

@Component({
    selector: 'app-field-count-panel',
    templateUrl: './field-count-panel.component.html',
    styles: [`
        h3 {
          margin-top: 0;
          margin-bottom: 0;
          margin-left: 0;
        }
        .field-panel {
            clear: both;
            margin-left: 1em;
            margin-bottom: 0.2em;
        }
        .fieldValueInfo {
            background: #484848;
            float:left;
            margin-bottom:10px;
            margin-right:10px;
            padding: 10px 10px 10px 10px;
        }
    `],
})


export class FieldCountPanelComponent {
    @Input() field: string;

    constructor(private fieldValuesProvider: FieldValueProvider) { }

    nameDisplay() {
        return this.fieldValuesProvider.fieldNameDisplay(this.field);
    }

    showField() {
        return this.fieldValuesProvider.showField(this.field);
    }

    toggleField() {
        this.fieldValuesProvider.toggleField(this.field);
    }

    fieldAndCounts() {
        return this.fieldValuesProvider.fieldAndCounts(this.field);
    }
}
