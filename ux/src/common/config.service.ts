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
import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { pick, reduce } from 'lodash';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import {
  CONFIG_CENTER_PREFIX,
  DOMAON_NAME,
  GOVERN_PREFIX,
  PROJECT_ID,
  REGISTRY_PREFIX,
} from 'src/config/global.config';

@Injectable({
  providedIn: 'root',
})
export class ConfigService {
  constructor(private http: HttpClient) {}

  getAllKies(): Observable<{ total: number; data: KieItem[] }> {
    return this.http
      .get<{ total: number; data: KieItem[] }>(`${CONFIG_CENTER_PREFIX}/kv`, {
        headers: {
          'Content-Type': 'application/json;charset=UTF-8',
        },
      })
      .pipe(map((item) => item));
  }

  getKie(id: string): Observable<any> {
    return this.http.get(`${CONFIG_CENTER_PREFIX}/kv/${id}`);
  }

  postKie(body: any): Observable<any> {
    return this.http.post(`${CONFIG_CENTER_PREFIX}/kv`, body, {
      headers: {
        'Content-Type': 'application/json',
      },
    });
  }

  putKie(id: string, body: any): Observable<any> {
    return this.http.put(`${CONFIG_CENTER_PREFIX}/kv/${id}`, body, {
      headers: {
        'Content-Type': 'application/json',
      },
    });
  }

  deleteKie(id: string): Observable<any> {
    return this.http.delete(`${CONFIG_CENTER_PREFIX}/kv/${id}`);
  }

  getApps(): Observable<any> {
    return this.http
      .get<any>(`${GOVERN_PREFIX}/microservices`, {
        headers: {
          'x-domain-name': DOMAON_NAME,
        },
      })
      .pipe(
        map((data) => {
          const appIdList: string[] = [];
          const results = (data.allServicesDetail || []).reduce(
            (list: any, item: any) => {
              if (!appIdList.includes(item.microService.appId)) {
                appIdList.push(item.microService.appId);
                list.push({
                  appId: item.microService.appId,
                  environment: item.microService.environment || '',
                });
              }
              return list;
            },
            []
          );
          return results;
        })
      );
  }

  getServices(): Observable<any> {
    return this.http.get(`${REGISTRY_PREFIX}/microservices`, {
      headers: {
        'X-Enterprise-Project-ID': PROJECT_ID,
      },
    });
  }
}

interface KieItem {
  [name: string]: any;
}

export const getTagsByObj = (labels: { [x: string]: any }): string[] => {
  const data = Object.keys(labels || {}).reduce(
    (list: string[], key: string) => {
      if (key) {
        list.push(`${key}=${labels[key]}`);
      }
      return list;
    },
    []
  );
  return data;
};
