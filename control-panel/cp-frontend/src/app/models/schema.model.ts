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

import * as yaml from 'js-yaml';
import * as _ from 'lodash';
import {ApiModel} from './api.model';

export class SchemaModel {
  schemaId: string;
  schema: string;
  summary: string;
  schemaObj: any;
  apis: ApiModel[];

  public parseSchema(): void {
    if (this.schema) {
      this.schemaObj = yaml.safeLoad(this.schema);
      const paths = this.schemaObj.paths;
      this.apis = [];
      _.keys(this.schemaObj.paths).forEach(p => {
        _.keys(paths[p]).forEach(m => {
          const method = m;
          const pathArray = _(p).split('/');
          let shortPath = '';
          if (pathArray.size() > 3) {
            shortPath = '.../' + pathArray.takeRight(2).join('/');
          } else {
            shortPath = p;
          }
          const overview = [method + ' ' + p, paths[p][method].description].join('\n');
          this.apis.push({method: method.toUpperCase(), path: p, shortPath, overview} as ApiModel);
        });
      });
    }
  }
}
