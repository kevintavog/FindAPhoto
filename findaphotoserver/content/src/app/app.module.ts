import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { HttpModule } from '@angular/http';

import { AppComponent } from './app.component';
import { SearchComponent } from './search/search.component';
import { SearchByDayComponent } from './search-by-day/search-by-day.component';
import { SearchByLocationComponent } from './search-by-location/search-by-location.component';
import { SingleItemComponent } from './single-item/single-item.component';

import { SearchResultsProvider } from './providers/search-results.provider';
import { SearchService } from './services/search.service';
import { SearchRequestBuilder } from './models/search.request.builder';

import { DataDisplayer } from './providers/data-displayer';
import { NavigationProvider } from './providers/navigation.provider';

import { routing } from './app.routes';
import { MapComponent } from './map/map.component';
import { PagingComponent } from './components/paging/paging.component';
import { HeaderComponent } from './components/header/header.component';
import { AlertsComponent } from './components/alerts/alerts.component';


@NgModule({
    declarations: [
        AppComponent,
        MapComponent,
        PagingComponent,
        SearchComponent,
        SearchByDayComponent,
        SearchByLocationComponent,
        SingleItemComponent,
        HeaderComponent,
        AlertsComponent
    ],
    imports: [
        BrowserModule,
        FormsModule,
        HttpModule,
        routing
    ],
    providers: [
        DataDisplayer,
        NavigationProvider,
        SearchRequestBuilder,
        SearchResultsProvider,
        SearchService
    ],
    bootstrap: [
        AppComponent
    ]
})

export class AppModule { }
