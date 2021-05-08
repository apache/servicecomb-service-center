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
import { HttpClientXsrfModule } from '@angular/common/http';
import { NgModule } from '@angular/core';
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
  ],
  exports: [
    ...devUIModules,
    ...angularModule,
    ...cusModule,
    ...pipes,
    ...derective,
    MonacoEditorModule,
  ],
})
export class SharedModule {}
