import { Routes, RouterModule } from '@angular/router';

import { MapComponent } from './map/map.component';
import { SearchComponent } from './search/search.component';
import { SearchByDayComponent } from './search-by-day/search-by-day.component';
import { SearchByLocationComponent } from './search-by-location/search-by-location.component';
import { SingleItemComponent } from './single-item/single-item.component';


const routes: Routes = [
    { path: '', redirectTo: 'search', pathMatch : 'full' },
    { path: 'byday', component: SearchByDayComponent },
    { path: 'bylocation', component: SearchByLocationComponent },
    { path: 'map', component: MapComponent },
    { path: 'search', component: SearchComponent },
    { path: 'singleitem', component: SingleItemComponent },
];

export const routing = RouterModule.forRoot(routes);
