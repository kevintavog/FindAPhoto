import { Pipe, PipeTransform } from 'angular2/core';

@Pipe({
    name: 'DateStringToLocaleDatePipe'
})

export class DateStringToLocaleDatePipe implements PipeTransform {
    transform(value, args?) : any {
        return new Date(value).toLocaleString()
    }
}
