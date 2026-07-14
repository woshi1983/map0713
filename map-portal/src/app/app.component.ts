import { Component, OnInit } from '@angular/core';
import { BnNgIdleService } from 'bn-ng-idle';
import { Router } from '@angular/router';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {
  constructor (
    private bnIdle: BnNgIdleService,
    private router: Router
  ) {}

  ngOnInit (): void {
    console.log('The Application is Starting.');

    // set 10 min idle timeout
    this.bnIdle.startWatching(1800).subscribe((isTimeOut: boolean) => {
      if (isTimeOut) {
        this.router.navigate(['/portal/logout']);
      } else {
        this.bnIdle.resetTimer();
      }
    });
  }
}
