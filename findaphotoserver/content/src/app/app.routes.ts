import { Routes, RouterModule } from '@angular/router';

import { InfoComponent } from './components/info/info.component';
import { FieldValuesComponent } from './components/field-values/field-values.component';
import { MapComponent } from './components/map/map.component';
import { SearchComponent } from './components/search/search.component';
import { SearchByDayComponent } from './components/search/search-by-day.component';
import { SearchByLocationComponent } from './components/search/search-by-location.component';
import { SingleItemComponent } from './components/single-item/single-item.component';


const routes: Routes = [
    { path: '', redirectTo: 'search', pathMatch : 'full' },
    { path: 'byday', component: SearchByDayComponent },
    { path: 'bylocation', component: SearchByLocationComponent },
    { path: 'fieldvalues', component: FieldValuesComponent },
    { path: 'info', component: InfoComponent },
    { path: 'map', component: MapComponent },
    { path: 'search', component: SearchComponent },
    { path: 'singleitem', component: SingleItemComponent },
];

export const routing = RouterModule.forRoot(routes);
