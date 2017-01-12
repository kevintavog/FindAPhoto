import { Injectable } from '@angular/core';

@Injectable()
export class FieldValueProvider {
    toggleState = new Map();  // key is a field name, value is a boolean
    fieldData = new Map();    // key is a field name, value is a FieldAndCount[]


    fieldAndCounts(name: string) {
        return this.fieldData.get(name);
    }

    fieldNameDisplay(name: string) {
        let data = this.fieldAndCounts(name);
        if (data) {
            return name + '  (' + data.length + ')'
        }
        return name;
    }

    toggleField(name: string) {
        if (this.toggleState.has(name)) {
            this.toggleState.set(name, !this.toggleState.get(name))
        } else {
            this.toggleState.set(name, true);
        }
    }

    showField(name: string) {
        if (this.toggleState.has(name)) {
            return this.toggleState.get(name) === true;
        }
        return false;
    }
}
