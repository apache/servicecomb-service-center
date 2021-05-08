/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
import { NgModule } from '@angular/core';
import { SharedModule } from '../shared/shared.module';
import { InstanceListComponent } from './components/instance-list/instance-list.component';
import { InvokedServiceComponent } from './components/invoked-service/invoked-service.component';
import { ServiceContractComponent } from './components/service-contract/service-contract.component';
import { ServiceDetailRoutingModule } from './service-detail-routing.module';
import { ServiceDetailComponent } from './service-detail.component';
import { SwaggerComponent } from './components/service-contract/swagger/swagger.component';
import { OriginalDataComponent } from './components/original-data/original-data.component';

@NgModule({
  declarations: [
    ServiceDetailComponent,
    InstanceListComponent,
    InvokedServiceComponent,
    ServiceContractComponent,
    SwaggerComponent,
    OriginalDataComponent,
  ],
  imports: [SharedModule, ServiceDetailRoutingModule],
  exports: [ServiceDetailRoutingModule],
})
export class ServiceDetailModule {}
