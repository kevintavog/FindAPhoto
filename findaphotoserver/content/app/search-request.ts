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

    // A nearby search
    latitude: number;
    longitude: number;

    drilldown: string;
}
