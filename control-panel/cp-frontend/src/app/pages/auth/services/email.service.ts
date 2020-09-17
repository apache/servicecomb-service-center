import {Injectable} from '@angular/core';
import {Observable, of} from 'rxjs';

import {Email} from '../models';

@Injectable({
  providedIn: 'root'
})
export class EmailService {
  public loadEmails(): Observable<Email[]> {
    return of([
      {name: 'Alec Z', time: '9:32', message: 'Hey! How is it going?'},
      {name: 'Hannah C', time: '9:18', message: 'Check out my new Dashboard'}
    ]);
  }
}
