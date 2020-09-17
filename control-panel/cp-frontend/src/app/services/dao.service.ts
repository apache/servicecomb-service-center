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

import {Injectable} from '@angular/core';
import {ServiceModel} from '../models/service.model';
import {HttpClient, HttpHeaders} from '@angular/common/http';
import {apiUrls} from '../consts';
import {SchemaModel} from '../models/schema.model';
import {InstanceModel} from '../models/instance.model';
import {AppModel} from '../models/app.model';
import {combineLatest, Observable, Subject} from 'rxjs';
import {concatAll, map, tap} from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class DaoService {
  app: AppModel;
  public app$: Subject<AppModel> = new Subject<AppModel>();
  private httpOptions;

  constructor(private http: HttpClient) {
    this.app = new AppModel();
    this.app.appId = 'default';
    const httpOptions = {
      headers: new HttpHeaders({
        'x-domain-name': 'default',
      })
    };
    this.httpOptions = httpOptions;
  }

  public updateService(serviceIndex: number): Observable<any[]> {
    const url = apiUrls.GET_SERVICE.replace(':appId', this.app.appId).replace(':serviceId', this.app.services[serviceIndex].serviceId);
    return this.http.get(url, this.httpOptions).pipe(tap((data: any) => {
      this.app.services[serviceIndex] = data as ServiceModel;
      this.app$.next(this.app);
    }));
  }

  public indexServices(withDetail: boolean = false): Observable<any> {
    const url = apiUrls.GET_SERVICES.replace(':appId', this.app.appId);
    const options = {
      ...this.httpOptions,
      params: {options: 'all'}
    };
    return this.http.get(url, options).pipe(tap((data: any) => {
      this.app.retrieveFromRemote(data);
      this.app$.next(this.app);
    }));
  }

  public isNewService(serviceId: string): boolean {
    return !this.app.services.some((s) => s.serviceId === serviceId);
  }
}
