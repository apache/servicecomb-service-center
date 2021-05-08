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

@Component({
  selector: 'app-create-modal',
  templateUrl: './create-modal.component.html',
  styleUrls: ['./create-modal.component.less'],
})
export class CreateModalComponent implements OnInit {
  @Input() data!: {
    onClose: () => void;
  };
  constructor(private router: Router) {}

  items = [
    {
      title: '应用级配置',
      content: '将新建到配置关联到某一应用，并添加应用名称和所在环境到标签。',
      type: 'app',
    },
    {
      title: '微服务级配置',
      content:
        '将新建到配置关联到某一微服务，并添加微服务名称和所在环境到标签。',
      type: 'service',
    },
    {
      title: '自定义配置',
      content: '自定义一个新到配置文件。',
      type: 'custom',
    },
  ];

  ngOnInit(): void {}

  onCreateBtn(type: string): void {
    this.data.onClose();
    this.router.navigate(['/kie/create'], {
      queryParams: {
        configType: type,
      },
    });
  }
}
