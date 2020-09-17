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

import {InstanceModel} from './instance.model';
import {SchemaModel} from './schema.model';
import {AppModel} from './app.model';
import * as _ from 'lodash';
import {ApiModel} from './api.model';

export class ServiceModel {
  instances: InstanceModel[];
  schemas: SchemaModel[];
  consumers: ServiceModel[];
  providers: ServiceModel[];
  isLatest: boolean;
  serviceId: string;
  project: AppModel;
  serviceName: string;
  version: string;
  description: string;
  status: string;
  properties: any;
  app: AppModel;

  relativeServices: ServiceModel[];

  public retrieveFromRemote(data: any) {
    this.instances = [];
    if (data.instances) {
      data.instances.forEach(d => {
        const i = new InstanceModel();
        Object.assign(i, d);
        i.service = this;
        this.instances.push(i);
      });
    }
    this.schemas = [];
    if (!_.isEmpty(data.schemaInfos)) {
      data.schemaInfos.forEach(d => {
        const s = new SchemaModel();
        Object.assign(s, d);
        s.parseSchema();
        this.schemas.push(s);
      });
    }
    this.consumers = [];
    if (data.consumers) {
      data.consumers.forEach(d => {
        this.consumers.push(_.find(this.app.services, s => s.serviceId === d.serviceId));
      });
    }
    this.providers = [];
    if (data.providers) {
      data.providers.forEach(d => {
        this.providers.push(_.find(this.app.services, s => s.serviceId === d.serviceId));
      });
    }

  }

  public apis(): ApiModel[] {
    const apis = _(this.schemas).map(sch => sch.apis).flatten().value();

    return apis;
  }

  public compareTo(b: ServiceModel): number {
    const aV = this.serviceId.split('.');
    const bV = b.serviceId.split('.');

    for (let i = 0; i < aV.length; i++) {
      if (i > bV.length - 1) {
        return 1;
      }
      if (_.toNumber(aV[i]) < _.toNumber(bV[i])) {
        return -1;
      } else if (_.toNumber(aV[i]) > _.toNumber(bV[i])) {
        return 1;
      }
    }
    if (aV.length === bV.length) {
      return 0;
    } else {
      return -1;
    }
  }

  public relativeVersionsNum(): number {
    return this.relativeServices.length;
  }

  public previousVersionsNum(): number {
    return this.relativeVersionsNum() - _.findIndex(this.relativeServices, s => s.version === this.version) - 1;
  }

  public consumersOverview(): string {
    return _(this.consumers).map(s => s.serviceName + ' ' + s.version).join('\n');
  }

  public instancesOverview(): string {
    return _(this.instances).map(i => i.endpoints.join('\n')).join(',\n');
  }

  public previousVersionsOverview(): string {
    const i = _.findIndex(this.relativeServices, s => s.version === this.version);
    return _(this.relativeServices).takeRight(this.relativeVersionsNum() - i - 1).map(s => s.version).join('\n');
  }

  public providersOverview(): string {
    return _(this.providers).map(s => s.serviceName + ' ' + s.version).join('\n');
  }

  // TODO: change the mock!
  public kieConfOverview(): string {
    return 'chassis.longwaiting=1' + '\n' +
      'mesher.retry=2';
  }
}
