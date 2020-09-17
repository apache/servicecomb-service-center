import {BrowserModule} from '@angular/platform-browser';
import {NgModule} from '@angular/core';
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';
import {RouterModule} from '@angular/router';
import {ToastrModule} from 'ngx-toastr';
import {MatCardModule} from '@angular/material/card';
import {MatButtonModule} from '@angular/material/button';

import {AppComponent} from './app.component';
import {AppRoutingModule} from './app-routing.module';
import {NotFoundComponent} from './pages/not-found/not-found.component';
import {AuthModule} from './pages/auth/auth.module';
import {MatTabsModule} from '@angular/material/tabs';
import {ServiceOverviewPageComponent} from './pages/service-overview/containers';
import {InstanceListPageComponent} from './pages/instance-list/containers';
import {ServiceDetailPageComponent} from './pages/service-detail/containers';
import {TopologyPageComponent} from './pages/topology/containers';
import {ApiShortComponent} from './shared/components/api-short/api-short.component';
import {FooterComponent} from './shared/footer/footer.component';
import {HeaderModule} from './shared/header/header.module';
import {LayoutComponent} from './shared/layout/layout.component';
import {SidebarComponent} from './shared/sidebar/sidebar.component';
import {EditKieConfDialogComponent} from './shared/ui-elements/edit-kie-conf-dialog/edit-kie-conf-dialog.component';
import {SettingsMenuComponent} from './shared/ui-elements/settings-menu/settings-menu.component';
import {ServicesTableComponent} from './pages/service-list/components';
import {ServiceListPageComponent} from './pages/service-list/containers';
import {ServiceOverviewComponent} from './pages/service-overview/components';
import {DaoService} from './services/dao.service';
import {MatMenuModule} from '@angular/material/menu';
import {MatTooltipModule} from '@angular/material/tooltip';
import {MatListModule} from '@angular/material/list';
import {MatPaginatorModule} from '@angular/material/paginator';
import {MatProgressBarModule} from '@angular/material/progress-bar';
import {MatToolbarModule} from '@angular/material/toolbar';
import {MatGridListModule} from '@angular/material/grid-list';
import {MatIconModule} from '@angular/material/icon';
import {MatSelectModule} from '@angular/material/select';
import {MatInputModule} from '@angular/material/input';
import {MatFormFieldModule} from '@angular/material/form-field';
import {CommonModule} from '@angular/common';
import {FormsModule} from '@angular/forms';
import {MatSidenavModule} from '@angular/material/sidenav';
import {MatDialog, MatDialogModule} from '@angular/material/dialog';
import {MatCheckboxModule} from '@angular/material/checkbox';
import {MatTableModule} from '@angular/material/table';
import {InstancesTableComponent} from './pages/instance-list/components';
import {HttpClientModule} from '@angular/common/http';
import {SwaggerUiComponent} from './pages/service-detail/components/swagger-ui/swagger-ui.component';
import {MatBadgeModule} from '@angular/material/badge';

@NgModule({
  declarations: [
    AppComponent,
    NotFoundComponent,
    ServiceOverviewPageComponent,
    InstanceListPageComponent,
    ServiceDetailPageComponent,
    TopologyPageComponent,
    ApiShortComponent,
    FooterComponent,
    LayoutComponent,
    SidebarComponent,
    EditKieConfDialogComponent,
    SettingsMenuComponent,
    ServicesTableComponent,
    ServiceListPageComponent,
    ServiceOverviewComponent,
    InstancesTableComponent,
    SwaggerUiComponent
  ],
  imports: [
    AuthModule,
    BrowserModule,
    BrowserAnimationsModule,
    CommonModule,
    RouterModule,
    AppRoutingModule,
    HeaderModule,
    ToastrModule.forRoot(),
    MatCardModule,
    MatButtonModule,
    MatTabsModule,
    MatMenuModule,
    MatTooltipModule,
    MatListModule,
    MatPaginatorModule,
    MatIconModule,
    MatProgressBarModule,
    MatToolbarModule,
    MatGridListModule,
    MatSelectModule,
    MatInputModule,
    MatFormFieldModule,
    FormsModule,
    MatSidenavModule,
    MatDialogModule,
    MatCheckboxModule,
    MatTableModule,
    HttpClientModule,
    MatBadgeModule
  ],
  providers: [DaoService],
  bootstrap: [AppComponent]
})
export class AppModule {
}
