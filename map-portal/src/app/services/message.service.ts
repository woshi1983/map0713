import { Injectable } from '@angular/core';
import { Message } from '../classes';
import { Observable, Subject } from 'rxjs';
import { ALERTTYPE } from '../utils/static-entries/shared.const';

@Injectable({
  providedIn: 'root'
})
export class MessagesService {
  private subject = new Subject<any>();

  sendMessage (text: string, type?: ALERTTYPE): void {
    if (type === undefined) {
      type = ALERTTYPE.INFO;
    }
    this.subject.next(new Message(text, type));
  }

  getMessage (): Observable<any> {
    return this.subject.asObservable();
  }
}
