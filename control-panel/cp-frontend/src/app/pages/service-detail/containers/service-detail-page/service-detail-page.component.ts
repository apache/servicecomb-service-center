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
import {ActivatedRoute} from '@angular/router';
import {DaoService} from '../../../../services/dao.service';
import {ServiceModel} from '../../../../models/service.model';
import * as _ from 'lodash';
import {Observable, of} from 'rxjs';
import {InstanceModel} from '../../../../models/instance.model';

@Component({
  selector: 'app-map-page',
  templateUrl: './service-detail-page.component.html',
  styleUrls: ['./service-detail-page.component.scss']
})
export class ServiceDetailPageComponent implements OnInit {
  service: ServiceModel;
  instances$: Observable<InstanceModel[]> = of([]);
  consumers$: Observable<ServiceModel[]> = of([]);
  providers$: Observable<ServiceModel[]> = of([]);

  constructor(private route: ActivatedRoute, private daoService: DaoService) {

  }

  ngOnInit() {
    const serviceId = this.route.snapshot.paramMap.get('serviceId');
    this.service = _.find(this.daoService.app.services, s => s.serviceId === serviceId);
    this.instances$ = of(this.service.instances);
    this.consumers$ = of(this.service.consumers);
    this.providers$ = of(this.service.providers);
  }
}
