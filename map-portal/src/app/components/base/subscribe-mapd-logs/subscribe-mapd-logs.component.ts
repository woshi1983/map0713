import { Component } from '@angular/core';
import { Apollo, gql } from 'apollo-angular';
import { Subscription } from 'rxjs';
import { saveAs } from 'file-saver';

@Component({
  selector: 'app-subscribe-mapd-logs',
  templateUrl: './subscribe-mapd-logs.component.html',
  styleUrls: ['./subscribe-mapd-logs.component.css']
})
export class SubscribeMapdLogsComponent {
  mapdLogs: any[] = [];
  logsModelShow = false;
  loading = false;

  sub = gql`subscription{
      ConsoleLog
  }`

  subscription: Subscription;
  enable = true;

  constructor (private apollo: Apollo) {}

  open (): void {
    this.logsModelShow = true;
    this.loading = true;
    this.subscription = this.listen().subscribe(data => {
      this.loading = false;
      this.mapdLogs.push(data.data['ConsoleLog']);
    }, (e) => {
      console.log(e)
    })
  }

  listen () {
    return this.apollo.subscribe({
      query: this.sub,
      fetchPolicy: 'no-cache'
    })
  }

  close (): void {
    this.logsModelShow = false;
    this.mapdLogs = [];
    this.subscription.unsubscribe();
  }

  exportLogs (): void {
    const blob = new Blob([this.mapdLogs.join('')], { type: 'text/text' });
    saveAs(blob, 'mapd.log');
  }

  changeLogsSubscription (enable: boolean): void {
    this.enable = enable;
    if (enable) {
      this.loading = true;
      this.mapdLogs = [];
      this.subscription = this.listen().subscribe(data => {
        this.loading = false;
        this.mapdLogs.push(data.data['ConsoleLog']);
      }, (e) => {
        console.log(e)
      })
    } else {
      this.subscription.unsubscribe();
    }
  }
}
