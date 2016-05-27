import { Component, provide } from '@angular/core';
import { RouteConfig, ROUTER_DIRECTIVES, ROUTER_PROVIDERS } from '@angular/router-deprecated';
import { LocationStrategy, HashLocationStrategy } from '@angular/common';
import { HTTP_PROVIDERS } from '@angular/http'

import { SearchService } from './search.service';
import { SearchComponent } from './search.component';
import { SearchRequestBuilder } from './search.request.builder';
import { SlideComponent } from './slide.component';
import { SlideshowComponent } from './slideshow.component';
import { ByDayComponent } from './byday.component';
import { ByLocationComponent } from './bylocation.component';

@Component({
  selector: 'find-a-photo',
  template: `
    <router-outlet></router-outlet>
  `,
  styleUrls: ['app/app.component.css'],
  directives: [
      ROUTER_DIRECTIVES
  ],
  providers: [
    ROUTER_PROVIDERS,
    HTTP_PROVIDERS,
    SearchService,
    SearchRequestBuilder,
    provide(LocationStrategy, {useClass: HashLocationStrategy})
  ]
})

@RouteConfig([
  {
    path: '/search',
    name: 'Search',
    component: SearchComponent,
    useAsDefault: true
  },
  {
    path: '/slide/:id',
    name: 'Slide',
    component: SlideComponent
  },
  {
    path: '/slideshow',
    name: 'Slideshow',
    component: SlideshowComponent
  },
  {
    path: '/byday',
    name: 'ByDay',
    component: ByDayComponent
  },
  {
    path: '/byloc',
    name: 'ByLocation',
    component: ByLocationComponent
  }

])

export class AppComponent {
  title = 'Find A Photo';
}
