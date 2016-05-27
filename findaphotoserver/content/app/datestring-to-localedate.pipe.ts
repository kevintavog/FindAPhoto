import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
    name: 'DateStringToLocaleDatePipe'
})

export class DateStringToLocaleDatePipe implements PipeTransform {
    transform(value: string, args: string[]) : any {
        if (args != null) {
            if (args[0] == 'dateOnly') {
                var date = new Date(value)
                return date.toLocaleDateString()
            } else if (args[0] == 'shortDateAndTime') {
                var date = new Date(value)
                return date.toLocaleDateString() + " " + date.toLocaleTimeString()
            }
        }

        return new Date(value).toLocaleString()
    }
}
