import {NgModule} from '@angular/core';
import {MatToolbarModule} from '@angular/material/toolbar';
import {MatFormFieldModule} from '@angular/material/form-field';
import {CommonModule} from '@angular/common';
import {MatIconModule} from '@angular/material/icon';
import {MatMenuModule} from '@angular/material/menu';
import {MatButtonModule} from '@angular/material/button';
import {MatInputModule} from '@angular/material/input';
import {MatBadgeModule} from '@angular/material/badge';

import {HeaderComponent} from './containers';
import {NotificationsComponent} from './components/notifications/notifications.component';
import {SearchComponent} from './components/search/search.component';
import {ShortNamePipe} from './pipes';

@NgModule({
  declarations: [
    HeaderComponent,
    NotificationsComponent,
    SearchComponent,
    ShortNamePipe
  ],
  exports: [
    HeaderComponent
  ],
  imports: [
    CommonModule,
    MatToolbarModule,
    MatFormFieldModule,
    MatIconModule,
    MatMenuModule,
    MatButtonModule,
    MatInputModule,
    MatBadgeModule
  ]
})
export class HeaderModule {
}
