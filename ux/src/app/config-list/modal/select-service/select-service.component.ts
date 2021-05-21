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
import { cloneDeep } from 'lodash';
import { TableWidthConfig } from 'ng-devui';
import { getTabelData } from 'src/app/shared/toolFunction/tabel.pagination';
import { ServiceService } from 'src/common/service.service';
import { ConfigService } from '../../../../common/config.service';

@Component({
  selector: 'app-select-service',
  templateUrl: './select-service.component.html',
  styleUrls: ['./select-service.component.less'],
})
export class SelectServiceComponent implements OnInit {
  @Input() data!: {
    onClose: (rowItem?: any, version?: string) => void;
  };
  constructor(
    private service: ConfigService,
    private serviceService: ServiceService
  ) {}

  selectService: any;
  private basicDataSource!: any[];
  dataSource!: any[];
  pager = {
    total: 0,
    pageIndex: 1,
    pageSize: 10,
    pageSizeOptions: [5, 10],
  };

  tableWidthConfig: TableWidthConfig[] = [
    {
      field: 'radio',
      width: '50px',
    },
    {
      field: 'serviceName',
      width: '120px',
    },
    {
      field: 'appId',
      width: '120px',
    },
    {
      field: 'environment',
      width: '120px',
    },
  ];

  ngOnInit(): void {
    this.serviceService.getServiceByGovern().subscribe(
      (res) => {
        this.basicDataSource = simplify(res);
        this.pager.total = this.basicDataSource.length;
        this.dataSource = getTabelData(this.basicDataSource, this.pager);
        this.selectService = this.dataSource.find(
          (item: any, index: number) => index === 0
        );
      },
      (err) => {
        // todo 提示
      }
    );

    function simplify(data: any): any[] {
      const result = data.allServicesDetail.reduce((obj: any, item: any) => {
        if (!obj[item.microService.serviceName + item.microService.appId]) {
          obj[item.microService.serviceName + item.microService.appId] = {
            serviceName: item.microService.serviceName,
            appId: item.microService.appId,
            environment: item.microService.environment || '',
            versions: [item.microService],
          };
        } else {
          obj[
            item.microService.serviceName + item.microService.appId
          ].versions.push(item.microService);
        }
        return obj;
      }, {});

      const arr: any[] = [];
      Object.keys(result).map((key) => {
        arr.push(result[key]);
      });

      return arr;
    }
  }

  onConfirm(): void {
    this.data.onClose(this.selectService);
  }

  onCancel(): void {
    this.data.onClose();
  }

  public onPaginationChange(pageIndex: number, pageSize: number): void {
    this.dataSource = getTabelData(this.basicDataSource, {
      ...cloneDeep(this.pager),
      pageIndex,
      pageSize,
    });
  }
}
