import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { HttpModule } from '@angular/http';

import { AppComponent } from './app.component';
import { MapComponent } from './components/map/map.component';
import { SearchComponent } from './components/search/search.component';
import { SearchByDayComponent } from './components/search/search-by-day.component';
import { SearchByLocationComponent } from './components/search/search-by-location.component';
import { SingleItemComponent } from './components/single-item/single-item.component';

import { SearchResultsProvider } from './providers/search-results.provider';
import { SearchService } from './services/search.service';
import { SearchRequestBuilder } from './models/search.request.builder';

import { DataDisplayer } from './providers/data-displayer';
import { NavigationProvider } from './providers/navigation.provider';
import { FieldsProvider } from './providers/fields.provider';

import { routing } from './app.routes';
import { AlertsComponent } from './components/alerts/alerts.component';
import { InfoComponent } from './components/info/info.component';
import { HeaderComponent } from './components/header/header.component';
import { PagingComponent } from './components/paging/paging.component';
import { CategoryTreeViewComponent, CategoryDetailsTreeViewComponent } from
    './components/category-tree-view-component/category-tree-view.component';


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
        AlertsComponent,
        InfoComponent,
        CategoryTreeViewComponent,
        CategoryDetailsTreeViewComponent
    ],
    imports: [
        BrowserModule,
        FormsModule,
        HttpModule,
        routing
    ],
    providers: [
        DataDisplayer,
        FieldsProvider,
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
