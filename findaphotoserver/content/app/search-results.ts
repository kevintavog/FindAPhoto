export interface SearchResults {
  totalMatches: number;
  resultCount: number;
  groups: SearchGroup[];
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
