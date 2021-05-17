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
import { Component, Input, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { cloneDeep } from 'lodash';
import { getTabelData } from 'src/app/shared/toolFunction/tabel.pagination';
import { ServiceService } from 'src/common/service.service';

@Component({
  selector: 'app-invoked-service',
  templateUrl: './invoked-service.component.html',
  styleUrls: ['./invoked-service.component.less'],
})
export class InvokedServiceComponent implements OnInit {
  @Input() serviceName = '';
  @Input() appId = '';
  @Input() type: 'consumers' | 'providers' = 'consumers';

  constructor(private service: ServiceService, private route: ActivatedRoute) {
    this.serviceId = this.route.snapshot.paramMap.get('id');
  }

  serviceId: string | null;
  private basicDataSource: any[] = [];
  dataSource: any;
  pager = {
    total: 0,
    pageIndex: 1,
    pageSize: 5,
    pageSizeOptions: [5, 10],
  };

  columns = [
    {
      field: 'serviceName',
      fieldType: 'text',
      width: '300px',
    },
    {
      field: 'appId',
      fieldType: 'text',
      width: '150px',
    },
    {
      field: 'version',
      fieldType: 'text',
      width: '150px',
    },
    {
      field: 'timestamp',
      fieldType: 'date',
      width: '300px',
    },
  ];

  ngOnInit(): void {
    this.initData();
  }

  initData(): void {
    if (this.serviceId) {
      const params = {
        serviceName: this.serviceName,
        appId: this.appId,
      };
      this.service.getDependencies(params).subscribe(
        (res) => {
          this.basicDataSource = res[this.type];
          this.pager.total = this.basicDataSource.length;
        },
        (err) => {
          // todo 提示
        }
      );
    }
  }

  public onPaginationChange(pageIndex: number, pageSize: number): void {
    this.dataSource = getTabelData(this.basicDataSource, {
      ...cloneDeep(this.pager),
      pageIndex,
      pageSize,
    });
  }
}
