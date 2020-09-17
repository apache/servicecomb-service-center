/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
