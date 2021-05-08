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
  selector: 'app-service-contract',
  templateUrl: './service-contract.component.html',
  styleUrls: ['./service-contract.component.less'],
})
export class ServiceContractComponent implements OnInit {
  constructor(private service: ServiceService, private route: ActivatedRoute) {
    this.serviceId = this.route.snapshot.paramMap.get('id');
  }

  selectBtn = 'swagger';
  serviceId: string | null;
  options = [];

  monacoOptions = {
    theme: 'vs',
    language: 'yaml',
    readOnly: true,
  };

  code!: string;

  selected = { schema: '' };

  ngOnInit(): void {
    if (this.serviceId) {
      this.service.getSchemas(this.serviceId).subscribe(
        (res) => {
          this.options = res;
          this.selected = res[0];
          this.code = res[0].schema;
        },
        (err) => {
          // todo 提示
        }
      );
    }
  }

  onSelectLanguage(type: 'swagger' | 'yaml'): void {
    this.selectBtn = type;
  }
}
