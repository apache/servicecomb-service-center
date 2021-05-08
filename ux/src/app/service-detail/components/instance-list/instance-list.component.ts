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
import { ActivatedRoute } from '@angular/router';
import { cloneDeep } from 'lodash';
import { DialogService } from 'ng-devui';
import { ActionItem } from 'src/app/shared/action-menu/action-menu.module';
import { getTabelData } from 'src/app/shared/toolFunction/tabel.pagination';
import { ServiceService } from 'src/common/service.service';

@Component({
  selector: 'app-instance-list',
  templateUrl: './instance-list.component.html',
  styleUrls: ['./instance-list.component.less'],
})
export class InstanceListComponent implements OnInit {
  constructor(
    private service: ServiceService,
    private route: ActivatedRoute,
    private dialog: DialogService
  ) {
    this.serviceId = this.route.snapshot.paramMap.get('id') as string;
  }

  serviceId: string;
  private basicDataSource: any[] = [];
  dataSource: any;
  pager = {
    total: 0,
    pageIndex: 1,
    pageSize: 10,
    pageSizeOptions: [5, 10, 20, 50],
  };

  columns = [
    {
      field: 'hostName',
      header: '实例名称',
      fieldType: 'text',
      width: '300px',
    },
    {
      field: 'status',
      header: '实例状态',
      fieldType: 'text',
      width: '150px',
    },
    {
      field: 'endpoints',
      header: 'Endpoints',
      fieldType: 'text',
      width: '300px',
    },
    {
      field: 'modTimestamp',
      header: '更新时间',
      fieldType: 'date',
      width: '300px',
    },
    {
      field: 'operation',
      header: '操作',
      fieldType: 'text',
      width: '200px',
    },
  ];

  ngOnInit(): void {
    this.initData(this.serviceId);
  }

  initData(serviceId: string): void {
    this.dataSource = [];
    if (serviceId) {
      this.service.getInstancesbyServiceId(serviceId).subscribe(
        (res) => {
          this.basicDataSource = res.instances;
          this.pager.total = res.instances?.length;
          this.dataSource = getTabelData(this.basicDataSource, this.pager);
        },
        (err) => {
          // todo 提示
        }
      );
    }
  }

  onToggle(e?: any): void {}

  actionFn(rowItem: any): ActionItem[] {
    // UP在线,OUTOFSERVICE摘机,STARTING正在启动,DOWN下线,TESTING拨测状态。
    const actions: ActionItem[] = [
      {
        id: 'DOWN',
        label: '下线',
        disabled: rowItem.status === 'DOWN',
      },
      {
        id: 'UP',
        label: '上线',
        disabled: rowItem.status === 'UP',
      },
      {
        id: 'OUTOFSERVICE',
        label: '摘机',
        disabled: rowItem.status === 'OUTOFSERVICE',
      },
      {
        id: 'TESTING',
        label: '拨测',
        disabled: rowItem.status === 'TESTING',
      },
    ];

    return actions;
  }

  actionClick(e: ActionItem, rowItem: any): void {
    const results = this.dialog.open({
      id: 'action-modal',
      title: '提示',
      content: `确认改变实例 "${rowItem.hostName}" 的状态为${e.label}?`,
      buttons: [
        {
          text: '确定',
          cssClass: 'danger',
          handler: () => {
            this.service
              .setInstanceStatus(this.serviceId, rowItem.instanceId, e.id)
              .subscribe(
                (res: any) => {
                  this.initData(this.serviceId);
                },
                (err) => {
                  // todo 错误提示
                }
              );
            results.modalInstance.hide();
          },
        },
        {
          text: '取消',
          cssClass: 'common',
          handler: () => {
            results.modalInstance.hide();
          },
        },
      ],
    });
  }

  public onPaginationChange(pageIndex: number, pageSize: number): void {
    this.dataSource = getTabelData(this.basicDataSource, {
      ...cloneDeep(this.pager),
      pageIndex,
      pageSize,
    });
  }
}
