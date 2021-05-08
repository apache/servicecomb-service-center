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
import { ServiceService } from 'src/common/service.service';

@Component({
  selector: 'app-original-data',
  templateUrl: './original-data.component.html',
  styleUrls: ['./original-data.component.less'],
})
export class OriginalDataComponent implements OnInit {
  constructor(private service: ServiceService, private route: ActivatedRoute) {
    this.serivceId = this.route.snapshot.paramMap.get('id') || '';
  }

  serivceId: string;
  code = '';
  monacoOptions = {
    theme: 'vs',
    language: 'json',
    readOnly: true,
  };

  ngOnInit(): void {
    if (!this.serivceId) {
      return;
    }
    this.initData(this.serivceId);
  }

  initData(serivceId: string): void {
    this.service
      .getServiceDetalByGovern(serivceId, { options: 'instances' } as any)
      .subscribe(
        (res) => {
          this.code = JSON.stringify(res.service, null, 2);
        },
        (err) => {
          // todo 提示
        }
      );
  }
}
