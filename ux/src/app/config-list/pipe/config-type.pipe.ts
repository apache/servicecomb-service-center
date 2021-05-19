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
import { Pipe, PipeTransform } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';

@Pipe({
  name: 'ConfigTypePipe',
})
export class ConfigTypePipe implements PipeTransform {
  constructor(private translate: TranslateService) {
    this.translate
      .get('common')
      .subscribe(({ appConfig, customConfig, serviceConfig }) => {
        this.types = {
          app: appConfig,
          custom: customConfig,
          service: serviceConfig,
        };
      });
  }

  types = {
    app: '',
    custom: '',
    service: '',
  };

  transform(value: Lables): string {
    return this.types[configTypeFn(value)];
  }
}

export const configTypeFn = (value: Lables): type => {
  if (!value) {
    return 'custom';
  }
  if (
    value.app &&
    value.service &&
    value.version &&
    value.environment !== undefined
  ) {
    return 'service';
  }
  if (value.app && value.environment !== undefined) {
    return 'app';
  }
  return 'custom';
};

interface Lables {
  [lable: string]: string;
}

interface Types {
  app: string;
  custom: string;
  service: string;
}

type type = keyof Types;
