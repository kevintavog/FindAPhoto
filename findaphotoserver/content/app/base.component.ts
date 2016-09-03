import { SearchResults,SearchGroup,SearchItem } from './search-results';

interface DegreesMinutesSeconds {
    degrees: number
    minutes: number
    seconds: number
}


export abstract class BaseComponent {

    itemYear(item: SearchItem) {
        let date = this.getItemDate(item)
        if (date != null) {
            return date.getFullYear()
        }
        return -1
    }

    itemMonth(item: SearchItem) {
        let date = this.getItemDate(item)
        if (date != null) {
            return date.getMonth() + 1
        }
        return -1
    }

    itemDay(item: SearchItem) {
        let date = this.getItemDate(item)
        if (date != null) {
            return date.getDate()
        }
        return -1
    }

    getItemDate(item: SearchItem) {
        if (item.createdDate != null) {
            let date = item.createdDate
            if (typeof item.createdDate === 'string') {
                date = new Date(item.createdDate.toString())
            }
            return date
        }
        return undefined
    }

    getItemLocaleDate(item: SearchItem) {
        if (item.createdDate != null) {
            return new Date(item.createdDate).toLocaleDateString()
        }
        return undefined
    }

    getItemLocaleDateAndTime(item: SearchItem) {
        if (item.createdDate != null) {
            let d = new Date(item.createdDate)
            return d.toLocaleDateString() + "  " + d.toLocaleTimeString()
        }
        return undefined
    }

    lonDms(item: SearchItem) {
        return this.longitudeDms(item.longitude)
    }

    longitudeDms(longitude: number) {
        return this.convertToDms(longitude, ["E", "W"])
    }

    latDms(item: SearchItem) {
        return this.latitudeDms(item.latitude)
    }

    latitudeDms(latitude: number) {
        return this.convertToDms(latitude, ["N", "S"])
    }


    convertToDms(degrees: number, refValues: string[]) : string {
        var dms = this.degreesToDms(degrees)
        var ref = refValues[0]
        if (dms.degrees < 0) {
            ref = refValues[1]
            dms.degrees *= -1
        }
        return dms.degrees + "° " + dms.minutes + "' " + dms.seconds.toFixed(2) + "\" " + refValues[1]
    }

    degreesToDms(degrees: number):DegreesMinutesSeconds {

        var d = degrees
        if (d < 0) {
            d = Math.ceil(d)
        } else {
            d = Math.floor(d)
        }

        var minutesSeconds = Math.abs(degrees - d) * 60.0
        var m = Math.floor(minutesSeconds)
        var s = (minutesSeconds - m) * 60.0

        return {
            degrees: d,
            minutes: m,
            seconds: s};
    }
}
