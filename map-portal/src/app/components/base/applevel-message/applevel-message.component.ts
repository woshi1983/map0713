import { Component, OnDestroy } from '@angular/core';
import { Subscription } from 'rxjs';
import { MessagesService } from '../../../services/message.service';
import { Message } from '../../../classes';
import { ALERTTYPE } from '../../../utils/static-entries/shared.const';

@Component({
  selector: 'app-applevel-messages',
  templateUrl: './applevel-message.component.html',
  styleUrls: ['./applevel-message.component.css']
})

export class ApplevelMessageComponent implements OnDestroy {
  messages: Message;
  alertType: ALERTTYPE;
  subscription: Subscription;
  timeId: any;

  constructor (private messagesService: MessagesService) {
    // subscribe to app component messages
    this.subscription = this.messagesService.getMessage().subscribe(message => {
      this.alertType = message.type;
      this.messages = message;
      this.timeId = setTimeout(this.clearMsg, 3600000, this);
    });
  }

  ngOnDestroy (): void {
    // unsubscribe to ensure no memory leaks
    this.subscription.unsubscribe();
  }

  clearMsg (): void {
    this.messages = null;
    this.timeId = null;
  }
}
