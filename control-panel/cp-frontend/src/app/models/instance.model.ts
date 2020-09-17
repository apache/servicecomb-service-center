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

export class InstanceModel {
  instanceId: string;
  service: ServiceModel;
  endpoints: string[];
  hostName: string;
  status: string;
  timeStamp: string;
  modTimestamp: string;

  public endpointsOverview(): string {
    if (this.endpoints.length > 1) {
      return _.join(this.endpoints, '\n');
    } else {
      return '';
    }
  }

  public endpointsSimple(): string {
    if (this.endpoints.length > 1) {
      return this.endpoints[0] + ', ...';
    } else {
      return this.endpoints[0];
    }
  }
}
