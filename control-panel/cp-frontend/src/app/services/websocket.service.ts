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
import {Observable, Observer, Subject} from 'rxjs';
import {map} from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class WebsocketService {
  constructor() {
  }

  connect(url): Subject<any> {
    const ws = new WebSocket(url);
    const observable = new Observable<MessageEvent>((obs: Observer<MessageEvent>) => {
      ws.onmessage = obs.next.bind(obs);
      ws.onclose = obs.complete.bind(obs);
      ws.onerror = obs.error.bind(obs);
      return ws.close.bind(ws);
    });
    const subject = new Subject<MessageEvent>();
    const resSubject = subject.pipe<any>(map((me: MessageEvent) => {
      return JSON.parse(me.data) as any;
    })) as Subject<any>
    observable.subscribe(subject);
    return resSubject;
  }
}
