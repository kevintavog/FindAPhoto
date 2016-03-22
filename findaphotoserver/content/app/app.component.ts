import { Component, provide } from 'angular2/core';
import { RouteConfig, ROUTER_DIRECTIVES, ROUTER_PROVIDERS } from 'angular2/router';
import { LocationStrategy, HashLocationStrategy } from 'angular2/router';
import { HTTP_PROVIDERS } from 'angular2/http'

import { SearchService } from './search.service';
import { SearchComponent } from './search.component';
import { SlideComponent } from './slide.component';
import { TodayComponent } from './today.component';

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
    path: '/today',
    name: 'Today',
    component: TodayComponent
  }

])

export class AppComponent {
  title = 'Find A Photo';
}
