export class SortType {
    static DateOldest = 'do';
    static DateNewest = 'dn';
    static LocationAZ = 'la';
    static LocationZA = 'lz';
    static FolderAZ = 'fa';
    static FolderZA = 'fz';
}

export interface SearchRequest {
    searchType: string;

    properties: string;
    first: number;
    pageCount: number;

    // A text search, the most common search
    searchText: string;

    // A by date search
    month: number;
    day: number;
    byDayRandom: boolean;

    // A nearby search
    latitude: number;
    longitude: number;
    maxKilometers: number;

    drilldown: string;
}
