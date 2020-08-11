import {Injectable} from '@angular/core';
import {Observable, Observer, Subject} from 'rxjs';
import {ServerEventModel} from '../model/server-event.model';
import {map} from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class WebsocketService {
  constructor() {
  }

  connect(url): Subject<ServerEventModel> {
    const ws = new WebSocket(url);
    const observable = new Observable<MessageEvent>((obs: Observer<MessageEvent>) => {
      ws.onmessage = obs.next.bind(obs);
      ws.onclose = obs.complete.bind(obs);
      ws.onerror = obs.error.bind(obs);
      return ws.close.bind(ws);
    });
    const subject = new Subject<MessageEvent>();
    const resSubject = subject.pipe<ServerEventModel>(map((me: MessageEvent) => {
      return JSON.parse(me.data) as ServerEventModel;
    })) as Subject<ServerEventModel>
    observable.subscribe(subject);
    return resSubject;
  }
}
