import { Component, OnInit, OnDestroy } from '@angular/core';
import { Router } from '@angular/router';

const defaultInterval = 1000;
const defaultLeftTime = 3;

@Component({
  selector: 'app-no-permission',
  templateUrl: './no-permission.component.html',
  styleUrls: ['./no-permission.component.css']
})
export class NoPermissionComponent implements OnInit, OnDestroy {
    leftSeconds: number = defaultLeftTime;
    timeInterval: any = null;

    constructor (private router: Router) {}

    ngOnInit (): void {
      if (!this.timeInterval) {
        this.timeInterval = setInterval(() => {
          this.leftSeconds--;
          if (this.leftSeconds <= 0) {
            this.router.navigate(['portal']);
            clearInterval(this.timeInterval);
          }
        }, defaultInterval);
      }
    }

    ngOnDestroy (): void {
      if (this.timeInterval) {
        clearInterval(this.timeInterval);
      }
    }
}
