import {NgModule} from '@angular/core';
import {CommonModule} from '@angular/common';
import {MatTabsModule} from '@angular/material/tabs';
import {MatButtonModule} from '@angular/material/button';
import {MatInputModule} from '@angular/material/input';
import {ReactiveFormsModule, FormsModule} from '@angular/forms';

import {AuthPageComponent} from './containers';
import {AuthRoutingModule} from './auth-routing.module';
import {YearPipe} from './pipes';
import {AuthService, EmailService} from './services';
import {LoginFormComponent} from './components';
import {AuthGuard} from './guards';

@NgModule({
  declarations: [
    AuthPageComponent,
    YearPipe,
    LoginFormComponent
  ],
  imports: [
    CommonModule,
    AuthRoutingModule,
    MatTabsModule,
    MatButtonModule,
    MatInputModule,
    ReactiveFormsModule,
    FormsModule
  ],
  providers: [
    AuthService,
    EmailService,
    AuthGuard
  ]
})
export class AuthModule {
}
