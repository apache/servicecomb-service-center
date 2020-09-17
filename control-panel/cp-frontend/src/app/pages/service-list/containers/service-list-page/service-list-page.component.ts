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

import {Component} from '@angular/core';
import {Observable, of} from 'rxjs';


import {DaoService} from '../../../../services/dao.service';
import {ServiceModel} from '../../../../models/service.model';
import {map} from 'rxjs/operators';

@Component({
  selector: 'app-tables-page',
  templateUrl: './service-list-page.component.html',
  styleUrls: ['./service-list-page.component.scss']
})
export class ServiceListPageComponent {
  public servicesTableData$: Observable<ServiceModel[]>;

  constructor(private daoService: DaoService) {

    this.servicesTableData$ = this.daoService.app$.pipe(map(a => a.services));
    this.daoService.indexServices().subscribe();
  }
}
