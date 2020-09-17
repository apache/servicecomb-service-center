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

import {Component, OnInit} from '@angular/core';
import {Observable} from 'rxjs';

import {DaoService} from '../../../../services/dao.service';
import {AppModel} from '../../../../models/app.model';
import {ServiceModel} from '../../../../models/service.model';
import {map} from 'rxjs/operators';

@Component({
  selector: 'app-dashboard-page',
  templateUrl: './service-overview-page.component.html',
  styleUrls: ['./service-overview-page.component.scss']
})
export class ServiceOverviewPageComponent implements OnInit {
  public app$: Observable<AppModel>;
  public services$: Observable<ServiceModel[]>;
  serviceN: number;
  instanceN: number;

  constructor(private daoService: DaoService) {
    this.app$ = this.daoService.app$;
    this.services$ = this.daoService.app$.pipe(map(a => a.services));
  }

  ngOnInit(): void {
    this.daoService.indexServices(true).subscribe(d => {
      this.serviceN = this.daoService.app.servicesN;
      this.instanceN = this.daoService.app.instanceN;
    });
  }
}
