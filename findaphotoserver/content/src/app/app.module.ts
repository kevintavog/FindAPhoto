import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { HttpModule } from '@angular/http';

import { AppComponent } from './app.component';
import { SearchComponent } from './search/search.component';
import { SearchByDayComponent } from './search-by-day/search-by-day.component';
import { SearchByLocationComponent } from './search-by-location/search-by-location.component';
import { SingleItemComponent } from './single-item/single-item.component';

import { SearchService } from './services/search.service';
import { SearchRequestBuilder } from './models/search.request.builder';

import { routing } from './app.routes';
import { MapComponent } from './map/map.component';


@NgModule({
    declarations: [
        AppComponent,
        SearchComponent,
        SearchByDayComponent,
        SearchByLocationComponent,
        SingleItemComponent,
        MapComponent
    ],
    imports: [
        BrowserModule,
        FormsModule,
        HttpModule,
        routing
    ],
    providers: [
        SearchRequestBuilder,
        SearchService
    ],
    bootstrap: [
        AppComponent
    ]
})

export class AppModule { }
