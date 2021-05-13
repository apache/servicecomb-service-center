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
import { CommonModule } from '@angular/common';
import { HttpClient, HttpClientXsrfModule } from '@angular/common/http';
import { Inject, LOCALE_ID, NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import {
  CategorySearchModule,
  DevUIModule,
  FormModule,
  ModalModule,
  RadioModule,
  SelectModule,
  TagsModule,
  TextareaModule,
  ToggleModule,
  TooltipModule,
} from 'ng-devui';
import { CheckBoxModule } from 'ng-devui/checkbox';
import { DataTableModule } from 'ng-devui/data-table';
import { LayoutModule } from 'ng-devui/layout';
import { MonacoEditorModule } from 'ngx-monaco-editor';
import { ActionMenuModule } from './action-menu/action-menu.module';
import { AutoHidePaginationDirective } from './derective/auto-hide-pagination.derective';
import { FilterRefreshModule } from './filter-refresh/filter-refresh.module';
import { EnvironmentPipe } from './pipe/environment.pipe';
import { InstanceStatusPipe } from './pipe/instance-status.pipe';

import { environment } from 'src/environments/environment';

import {
  TranslateModule,
  TranslateLoader,
  TranslateService,
} from '@ngx-translate/core';
import { TranslateHttpLoader } from '@ngx-translate/http-loader';

export function createTranslateLoader(http: HttpClient): TranslateHttpLoader {
  return new TranslateHttpLoader(http, './assets/i18n/', '.json');
}

const devUIModules = [
  CategorySearchModule,
  CheckBoxModule,
  DevUIModule,
  DataTableModule,
  FormModule,
  LayoutModule,
  ModalModule,
  RadioModule,
  TagsModule,
  ToggleModule,
  SelectModule,
  TextareaModule,
  TooltipModule,
];
const angularModule = [
  CommonModule,
  FormsModule,
  HttpClientXsrfModule,
  ReactiveFormsModule,
];

const cusModule = [ActionMenuModule, FilterRefreshModule];

const pipes = [InstanceStatusPipe, EnvironmentPipe];

const derective = [AutoHidePaginationDirective];

@NgModule({
  declarations: [...pipes, ...derective],
  imports: [
    ...devUIModules,
    ...angularModule,
    ...cusModule,
    MonacoEditorModule.forRoot(),
    TranslateModule.forRoot({
      loader: {
        provide: TranslateLoader,
        useFactory: createTranslateLoader,
        deps: [HttpClient],
      },
    }),
  ],
  exports: [
    ...devUIModules,
    ...angularModule,
    ...cusModule,
    ...pipes,
    ...derective,
    MonacoEditorModule,
    TranslateModule,
  ],
})
export class SharedModule {
  constructor(
    private i18n: TranslateService,
    @Inject(LOCALE_ID) locale: string
  ) {
    if (environment.supportedLocale.indexOf(locale as never) === -1) {
      locale = 'zh_CN';
    }

    this.i18n.use(locale);
  }
}
