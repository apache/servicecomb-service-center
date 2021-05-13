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
import { FormControl, FormGroup } from '@angular/forms';
import { TranslateService } from '@ngx-translate/core';
import { envOptions } from 'src/config/global.config';
import { ServiceService } from '../../../../common/service.service';

@Component({
  selector: 'app-create',
  templateUrl: './create.component.html',
  styleUrls: ['./create.component.less'],
})
export class CreateComponent implements OnInit {
  @Input() data!: {
    onClose: (data?: any) => void;
  };

  constructor(
    private service: ServiceService,
    private translate: TranslateService
  ) {
    this.translate.get('service.create').subscribe((i18n) => {
      this.formRules.versionRules.validators[4].message = i18n.versionMessage;
    });

    this.translate.get('common.empty').subscribe((empty) => {
      this.envOpetions = JSON.parse(
        JSON.stringify(
          envOptions.map((item) => {
            item.label = item.label || empty;
            return item;
          })
        )
      );
    });
  }

  formGroup = new FormGroup({
    serviceName: new FormControl(''),
    appId: new FormControl(''),
    version: new FormControl(''),
    environment: new FormControl(''),
    description: new FormControl(''),
  });
  formRules = {
    rule: { message: 'The form verification failed, please check.' },
    usernameRules: {
      validators: [
        { required: true },
        { minlength: 1 },
        { whitespace: true },
        { maxlength: 128 },
      ],
    },
    appIdRules: {
      validators: [
        { required: true },
        { whitespace: true },
        { minlength: 1 },
        { maxlength: 128 },
      ],
    },
    versionRules: {
      validators: [
        { required: true },
        { whitespace: true },
        { minlength: 3 },
        { maxlength: 46 },
        {
          pattern: /^\d{1,}\.(\d{1,}\.){1,2}\d{1,}$/,
          message: '',
        },
      ],
    },
    environmentRules: {
      validators: [],
    },
    descriptionRules: {
      validators: [],
    },
  };

  envOpetions!: {
    id: string;
    label: string;
  }[];

  ngOnInit(): void {
    this.formGroup.controls.environment.setValue(this.envOpetions[0]);
  }

  async onCreateBtn(): Promise<void> {
    const parmas = {
      service: {
        appId: this.formGroup.controls.appId.value,
        description: this.formGroup.controls.description.value,
        environment: this.formGroup.controls.environment.value?.id,
        serviceName: this.formGroup.controls.serviceName.value,
        version: this.formGroup.controls.version.value,
      },
    };
    const service = await this.service.postService(parmas).toPromise();
    this.data.onClose(service?.serviceId);
  }

  onCancel(): void {
    this.data.onClose();
  }
}
