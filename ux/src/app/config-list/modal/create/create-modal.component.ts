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
import { Router } from '@angular/router';
import { TranslateService } from '@ngx-translate/core';

@Component({
  selector: 'app-create-modal',
  templateUrl: './create-modal.component.html',
  styleUrls: ['./create-modal.component.less'],
})
export class CreateModalComponent implements OnInit {
  @Input() data!: {
    onClose: () => void;
  };
  constructor(private router: Router, private translate: TranslateService) {
    this.i18nInit();
  }

  items!: {
    title: string;
    content: string;
    type: string;
  }[];

  ngOnInit(): void {}

  async i18nInit(): Promise<void> {
    const common = await this.translate.get('common').toPromise();
    const i18n = await this.translate.get('kie.modal').toPromise();
    this.items = [
      {
        title: common.appConfig,
        content: i18n.appConfigContent,
        type: 'app',
      },
      {
        title: common.serviceConfig,
        content: i18n.serviceConfigContent,
        type: 'service',
      },
      {
        title: common.customConfig,
        content: i18n.customConfigContent,
        type: 'custom',
      },
    ];
  }

  onCreateBtn(type: string): void {
    this.data.onClose();
    this.router.navigate(['/kie/create'], {
      queryParams: {
        configType: type,
      },
    });
  }
}
