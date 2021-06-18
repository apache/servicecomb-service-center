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
import { TranslateService } from '@ngx-translate/core';
import { cloneDeep, flatten, map, uniq, uniqBy } from 'lodash';
import { DialogService, ICategorySearchTagItem, ModalService } from 'ng-devui';
import {
  FilterItem,
  filterTabDataByCategory,
  getTabelData,
} from 'src/app/shared/toolFunction/tabel.pagination';
import { ConfigService, getTagsByObj } from '../../../../common/config.service';
import { CreateModalComponent } from '../../modal/create/create-modal.component';

@Component({
  selector: 'app-config-list',
  templateUrl: './config-list.component.html',
  styleUrls: ['./config-list.component.less'],
})
export class ConfigListComponent implements OnInit {
  constructor(
    private service: ConfigService,
    private modalService: ModalService,
    private dialogService: DialogService,
    private translate: TranslateService
  ) {}
  private basicDataSource: any;
  dataSource: any;
  category: Array<ICategorySearchTagItem> | any = [];

  columns!: {
    field: string;
    header: string;
    fieldType: string;
    width: string;
  }[];

  pager = {
    total: 0,
    pageIndex: 1,
    pageSize: 10,
    pageSizeOptions: [5, 10, 20, 50],
  };

  ngOnInit(): void {
    this.onRefresh();
    this.translate.get('kie.list').subscribe((i18n) => {
      this.columns = [
        {
          field: 'key',
          header: i18n.columns.key,
          fieldType: 'text',
          width: '100px',
        },
        {
          field: 'status',
          header: i18n.columns.status,
          fieldType: 'text',
          width: '50px',
        },
        {
          field: 'lables',
          header: i18n.columns.lables,
          fieldType: 'text',
          width: '200px',
        },
        {
          field: 'type',
          header: i18n.columns.type,
          fieldType: 'text',
          width: '100px',
        },
        {
          field: 'value_type',
          header: i18n.columns.value_type,
          fieldType: 'text',
          width: '100px',
        },
        {
          field: 'update_time',
          header: i18n.columns.update_time,
          fieldType: 'date',
          width: '200px',
        },
        {
          field: 'operator',
          header: i18n.columns.operator,
          fieldType: 'text',
          width: '100px',
        },
      ];

      this.category = [
        {
          field: 'key',
          label: i18n.columns.key,
          type: 'textInput',
        },
        {
          field: 'status',
          label: i18n.columns.status,
          options: [],
          type: 'label',
        },
        {
          field: 'labels_format',
          label: i18n.columns.lables,
          options: [],
          type: 'label',
        },
      ];
    });
  }

  async onRefresh(): Promise<void> {
    const operator = await this.getI18n('kie.list.operator');
    this.service.getAllKies().subscribe(
      (data) => {
        this.basicDataSource = data.data
          .map((item) => {
            item.labels_format = getTagsByObj(item.labels);
            return item;
          })
          .sort((a, b) => b.update_time - a.update_time);
        this.pager.total = this.basicDataSource.length;
        this.dataSource = getTabelData(this.basicDataSource, this.pager);
        this.category[1].options = uniqBy(
          map(this.basicDataSource, (item: any) => ({
            id: item.status,
            label:
              item.status === 'enabled' ? operator.enabled : operator.forbidden,
          })),
          'label'
        );

        this.category[2].options = map(
          uniq(
            flatten(
              map(this.basicDataSource, (item: any) => item.labels_format)
            )
          ),
          (item) => ({
            label: item,
          })
        );
      },
      (err) => {
        console.log(err);
      }
    );
  }

  onCreate(): void {
    const results = this.modalService.open({
      id: 'modal-modal',
      width: '550px',
      backdropCloseable: true,
      component: CreateModalComponent,
      data: {
        onClose: () => {
          results.modalInstance.hide();
        },
      },
    });
  }

  async onForbidden(rowItem: { id: string; key: string }): Promise<void> {
    const common = await this.getI18n('common');
    const title = await this.getI18n('kie.list.tip');
    const content = await this.getI18n('kie.list.forbiddenContent', {
      key: rowItem.key,
    });
    const results = this.dialogService.open({
      id: 'forbidden',
      title,
      content,
      width: '400px',
      buttons: [
        {
          text: common.confirm,
          cssClass: 'danger',
          handler: async () => {
            await forbiddenFn(rowItem.id);
            results.modalInstance.hide();
            this.onRefresh();
          },
        },
        {
          text: common.cancel,
          cssClass: 'common',
          handler: () => {
            results.modalInstance.hide();
          },
        },
      ],
    });

    const forbiddenFn = (id: string) => {
      const pamars = {
        status: 'disabled',
      };
      return this.service.putKie(id, pamars).toPromise();
    };
  }

  async onEnable(rowItem: { id: string; key: string }): Promise<void> {
    const common = await this.getI18n('common');
    const title = await this.getI18n('kie.list.tip');
    const content = await this.getI18n('kie.list.enableContent', {
      key: rowItem.key,
    });
    const results = this.dialogService.open({
      id: 'forbidden',
      title,
      content,
      width: '400px',
      buttons: [
        {
          text: common.confirm,
          cssClass: 'danger',
          handler: async () => {
            await enableFn(rowItem.id);
            results.modalInstance.hide();
            this.onRefresh();
          },
        },
        {
          text: common.cancel,
          cssClass: 'common',
          handler: () => {
            results.modalInstance.hide();
          },
        },
      ],
    });

    const enableFn = (id: string) => {
      const pamars = {
        status: 'enabled',
      };
      return this.service.putKie(id, pamars).toPromise();
    };
  }

  async onDeleteItem(rowItem: { id: string; key: string }): Promise<void> {
    const common = await this.getI18n('common');
    const title = await this.getI18n('kie.list.tip');
    const content = await this.getI18n('kie.list.deleteContent', {
      key: rowItem.key,
    });
    const results = this.dialogService.open({
      id: 'deleteKie',
      width: '400px',
      showAnimate: true,
      title,
      content,
      buttons: [
        {
          text: common.confirm,
          cssClass: 'danger',
          handler: async () => {
            // todo
            await this.service.deleteKie(rowItem.id).toPromise();
            this.onRefresh();
            results.modalInstance.hide();
          },
        },
        {
          text: common.cancel,
          cssClass: 'common',
          handler: () => {
            // todo
            results.modalInstance.hide();
          },
        },
      ],
    });
  }

  onPaginationChange(pageIndex: number, pageSize: number): void {
    this.dataSource = getTabelData(this.basicDataSource, {
      ...cloneDeep(this.pager),
      pageIndex,
      pageSize,
    });
  }

  onSelectedTagsChange(e: FilterItem[]): void {
    const { data, tableData, pageination } = filterTabDataByCategory(
      this.basicDataSource,
      this.pager,
      e
    );
    this.pager = pageination;
    this.dataSource = tableData;
  }

  private async getI18n(key: string, param?: any): Promise<any> {
    return await this.translate.get(key, param).toPromise();
  }
}
