import { SearchResults,SearchGroup,SearchItem } from './search-results';

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
}
