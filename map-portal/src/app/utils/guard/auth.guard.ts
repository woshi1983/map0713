import { Injectable } from '@angular/core';
import { CanActivate, CanActivateChild } from '@angular/router';
import { Observable } from 'rxjs';
import { AuthService } from '../../services/auth.service';

@Injectable({
  providedIn: 'root'
})
export class AuthGuard implements CanActivate, CanActivateChild {
  constructor (private authService: AuthService) {}

  canActivate (): Observable<boolean> | boolean {
    if (!this.authService.isLoggedIn) {
      console.log('Auth Guard Failed.');
      this.authService.login();
      return false;
    }
    return true;
  }

  canActivateChild (): Observable<boolean> | boolean {
    return this.canActivate();
  }
}
