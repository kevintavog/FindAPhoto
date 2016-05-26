export interface SearchResults {
  totalMatches: number;
  resultCount: number;
  groups: SearchGroup[];
  categories: SearchCategory[];
  previousAvailableByDay: ByDayResult;
  nextAvailableByDay: ByDayResult;
}

export interface ByDayResult {
    month: number;
    day: number;
}

export interface SearchGroup {
  name: string;
  items: SearchItem[];
  resultIndex: number;
}

export interface SearchItem {
  city: string;
  createdDate: Date;
  id: string;
  imageName: string;
  keywords: string[];
  latitude: number;
  locationName: string;
  locationDisplayName: string;
  longitude: number;
  mediaType: string;
  mediaUrl: string
  mimeType: string;
  path: string;
  slideUrl: string;
  thumbUrl: string;
  warnings: string[];
  distanceKm: number;
}

export interface SearchCategory {
    name: string;
    values: SearchCategoryValues[];
}

export interface SearchCategoryValues {
    count: number;
    value: string;
    subCategories: SearchCategoryValues[];
}
