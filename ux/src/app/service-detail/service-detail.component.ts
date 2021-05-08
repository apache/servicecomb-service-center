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
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { FormLayout, ModalService } from 'ng-devui';
import { getTagsByObj } from 'src/common/config.service';
import { ServiceService } from '../../common/service.service';

@Component({
  selector: 'app-service-detail',
  templateUrl: './service-detail.component.html',
  styleUrls: ['./service-detail.component.less'],
})
export class ServiceDetailComponent implements OnInit {
  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private service: ServiceService,
    private module: ModalService
  ) {
    this.serivceId = this.route.snapshot.paramMap.get('id') || '';
  }

  title!: string;
  serivceId: string;
  serviceData!: any;
  formLayout = FormLayout;
  acticeTabId!: any;

  tags!: string[];

  ngOnInit(): void {
    if (this.serivceId) {
      this.initData();
    }
  }

  initData(): void {
    this.service.getServiceById(this.serivceId).subscribe(
      (res) => {
        this.serviceData = res.service || {};
        this.title = res.service.serviceName;

        const params = {
          options: 'all',
          serviceName: this.title,
          appId: this.serviceData.appId,
        };
        this.service.getServiceByGovern(params).subscribe(
          (data) => {
            data.allServicesDetail.forEach((item: any) => {
              if (item?.microService?.serviceId === this.serivceId) {
                // this.title = item.microService.serviceName;
                this.serviceData.tags = item.tags || {};
                this.serviceData.instances = item.instances;
              }
            });
            this.tags = getTagsByObj(this.serviceData.tags);
          },
          (err) => {
            // todo 提示
            this.router.navigate(['servicelist']);
          }
        );
      },
      (err) => {
        // todo 提示
      }
    );
  }

  activeTabChange(event: any): void {
    console.log('switch to', event);
  }
}
