/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import {ServiceModel} from './service.model';
import * as _ from 'lodash';

export class AppModel {
  services: ServiceModel[] = [];
  appId: string;
  servicesN: number;
  instanceN: number;

  public retrieveFromRemote(data: any) {
    this.services = [];
    this.servicesN = data.statistics.services.count;
    this.instanceN = data.statistics.instances.count;
    for (const sData of data.allServicesDetail) {
      const s = new ServiceModel();
      Object.assign(s, sData.microService);
      s.app = this;
      this.services.push(s);

    }

    for (let i = 0; i < data.allServicesDetail.length; i++) {
      this.services[i].retrieveFromRemote(data.allServicesDetail[i]);
    }


    const sMap: any = {};
    _(this.services).each((s: ServiceModel) => {
      if (sMap[s.serviceName] === undefined || sMap[s.serviceName] === null) {
        sMap[s.serviceName] = [];
      }
      sMap[s.serviceName].push(s);
    });
    _.forEach(sMap, (v: ServiceModel[], k: string) => {
      sMap[k] = v.sort((a, b) => a.compareTo(b)).reverse();
    });
    _(this.services).each((s: ServiceModel) => {
      s.relativeServices = sMap[s.serviceName];
      s.isLatest = (s.serviceId === s.relativeServices[0].serviceId);
    });
    this.services.sort((a, b) => (a.consumers.length - b.consumers.length) * 1e12
      + (a.instances.length - b.instances.length) * 1e6
      + (a.relativeVersionsNum() - b.relativeVersionsNum())).reverse();
  }
}
