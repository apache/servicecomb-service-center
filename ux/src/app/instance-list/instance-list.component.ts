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
import { DialogService, ModalService, TableWidthConfig } from 'ng-devui';
import { ServiceService } from 'src/common/service.service';
import { ActionItem } from '../shared/action-menu/action-menu.module';
import { getTabelData } from '../shared/toolFunction/tabel.pagination';

@Component({
  selector: 'app-instance-list',
  templateUrl: './instance-list.component.html',
  styleUrls: ['./instance-list.component.less'],
})
export class InstanceListComponent implements OnInit {
  constructor(
    private service: ServiceService,
    private dialog: DialogService,
    private module: ModalService
  ) {}
  title = '实例列表';

  private basicDataSource = [];
  dataSource: any[] = [];
  pager = {
    total: 0,
    pageIndex: 1,
    pageSize: 5,
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
      header: '状态',
      fieldType: 'text',
      width: '100px',
    },
    {
      field: 'environment',
      header: '环境',
      fieldType: 'text',
      width: '100px',
    },
    {
      field: 'serviceName',
      header: '所属服务',
      fieldType: 'text',
      width: '100px',
    },
    {
      field: 'endpoints',
      header: 'Endpoints',
      fieldType: 'text',
      width: '200px',
    },
    {
      field: 'version',
      header: '所属版本',
      fieldType: 'text',
      width: '150px',
    },
    {
      field: 'modTimestamp',
      header: '更新时间',
      fieldType: 'text',
      width: '250px',
    },
  ];

  ngOnInit(): void {
    this.initData();
  }

  initData(): void {
    this.dataSource = [];
    this.service.getInstances().subscribe(
      (res) => {
        this.basicDataSource = res;
        this.pager.total = res.length;
        this.dataSource = getTabelData(this.basicDataSource, this.pager);
      },
      (err) => {
        // todo 提示
      }
    );
  }

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
              .setInstanceStatus(rowItem.serviceId, rowItem.instanceId, e.id)
              .subscribe(
                (res: any) => {
                  this.initData();
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
}
