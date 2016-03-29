import { Pipe, PipeTransform } from 'angular2/core';

@Pipe({
    name: 'DateStringToLocaleDatePipe'
})

export class DateStringToLocaleDatePipe implements PipeTransform {
    transform(value, args?) : any {
        if (args[0] == 'dateOnly') {
            var date = new Date(value)
            return date.toLocaleDateString()
        } else if (args[0] == 'shortDateAndTime') {
            var date = new Date(value)
            return date.toLocaleDateString() + " " + date.toLocaleTimeString()
        } else {
            return new Date(value).toLocaleString()
        }
    }
}
