import {PreloadAllModules, RouterModule, Routes} from '@angular/router';
import {NgModule} from '@angular/core';
import {ServiceOverviewPageComponent} from './pages/service-overview/containers';
import {NotFoundComponent} from './pages/not-found/not-found.component';
import {AuthGuard} from './pages/auth/guards';
import {ServiceListPageComponent} from './pages/service-list/containers';
import {TopologyPageComponent} from './pages/topology/containers';
import {InstanceListPageComponent} from './pages/instance-list/containers';
import {ServiceDetailPageComponent} from './pages/service-detail/containers';

const routes: Routes = [
  {
    path: 'service_overview',
    pathMatch: 'full',
    canActivate: [AuthGuard],
    component: ServiceOverviewPageComponent
  },
  {
    path: 'service_list',
    pathMatch: 'full',
    canActivate: [AuthGuard],
    component: ServiceListPageComponent
  },
  {
    path: 'service_topology',
    pathMatch: 'full',
    canActivate: [AuthGuard],
    component: TopologyPageComponent
  },
  {
    path: 'instance_list',
    pathMatch: 'full',
    canActivate: [AuthGuard],
    component: InstanceListPageComponent
  },
  {
    path: 'service_detail/:serviceId',
    pathMatch: 'full',
    canActivate: [AuthGuard],
    component: ServiceDetailPageComponent
  },
  {
    path: '404',
    component: NotFoundComponent
  },
  {
    path: 'login',
    loadChildren: () => import('./pages/auth/auth.module').then(m => m.AuthModule)
  },
  {
    path: '**',
    redirectTo: '404'
  }
];

@NgModule({
  imports: [
    RouterModule.forRoot(routes, {
      useHash: true,
      preloadingStrategy: PreloadAllModules
    })
  ],
  exports: [RouterModule]
})

export class AppRoutingModule {
}
