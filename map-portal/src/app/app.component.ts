import { Component, OnInit } from '@angular/core';
import { BnNgIdleService } from 'bn-ng-idle';
import { Router } from '@angular/router';
import { BrandingService } from './services/branding.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {
  constructor (
    private bnIdle: BnNgIdleService,
    private router: Router,
    private branding: BrandingService
  ) {}

  ngOnInit (): void {
    this.branding.load().subscribe({ error: error => console.warn('Unable to load branding settings:', error) });
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
