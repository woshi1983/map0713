import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { CookieService} from 'ngx-cookie';
import { Observable, throwError } from 'rxjs';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { catchError } from 'rxjs/operators';
import { handleObject, PermValidator, UserInfo } from '../classes';
import {
  USER_INFO_COOKIE_KEY,
  TOKEN_COOKIE_KEY, CURRENT_ORG_KEY, USER_ROLE_KEY, cookieOption
} from '../utils/static-entries/shared.const';
import { UsersService } from './users.service';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private tokenCookieKey = TOKEN_COOKIE_KEY;
  private userInfoCookieKey = USER_INFO_COOKIE_KEY;
  private userRoleCookieKey = USER_ROLE_KEY;
  private authAPI = environment.authAPI;
  private permValidator: PermValidator;

  private authenticated: boolean;
  loginUserDisplayName = '';
  loginUserId = '';
  loginUserName = '';
  loginUserAdmin: boolean;
  loginUserManaged: boolean;

  constructor (
    private router: Router,
    private cookie: CookieService,
    private userService: UsersService,
    private http: HttpClient
  ) {
    this.getLoginSession();
  }

  // used for set user displayName
  public setUserDisplayName (displayName: string): void {
    this.loginUserDisplayName = displayName;
  }

  login (): void {
    const href = window.location.href;
    window.location.href = window.location.origin + '/api/auth/.well-known/login?redirectURI=' + encodeURI(href);
  }

  private handleError = (error: any) => throwError(error);

  requestLogout (): Observable<any> {
    const logoutURL = this.authAPI + '/.well-known/logout';
    const httpHeaders = new HttpHeaders(
      {
        'Content-Type': 'application/json'
      }
    );
    return this.http.get(logoutURL, {
      headers: httpHeaders,
      responseType: 'json',
      observe: 'response' as 'body'
    }
    ).pipe(
      catchError(this.handleError)
    );
  }

  logout (): void {
    // remove token cookie
    this.requestLogout().subscribe(() => {
      this.cookie.remove(CURRENT_ORG_KEY, { path: '/' });
      this.cookie.remove(this.userInfoCookieKey, { path: '/' });
      this.cookie.remove(this.userRoleCookieKey, { path: '/' });
      this.authenticated = false;
    }, (e) => {
      throwError(e);
    }
    )
  }

  private getLoginSession (): void {
    const accessToken = this.cookie.get(this.tokenCookieKey) ? this.cookie.get(this.tokenCookieKey).trim() : null;
    if (accessToken) {
      const userInfo = this.getUserInfo();
      this.loginUserId = userInfo.id;
      this.loginUserName = userInfo.name;
      this.userService.getUserInfoAndRoleById(userInfo.id).subscribe(data => {
        this.loginUserDisplayName = data.data.User.displayName;
        this.loginUserAdmin = data.data.User.admin;
        this.loginUserManaged = data.data.User.managed;
        let role: {[organizationId: string]: string} = {};
        if(!!data?.data?.User?.roleBindings?.edge?.nodes){
          for(let each of data.data.User.roleBindings.edge.nodes){
            role[each.organization.id] = each.role?.toUpperCase();
          }
        }
        this.cookie.put(USER_ROLE_KEY, JSON.stringify(role), cookieOption);
      })
      this.authenticated = true;
    } else {
      this.authenticated = false;
    }
  }

  get isLoggedIn (): boolean {
    // Check if current date is before token
    // expiration and user is signed in locally
    // return Date.now() < this.expiresAt && this.authenticated;
    return this.authenticated;
  }

  getUserInfo (): UserInfo {
    const str = this.getSecuredCookie(this.userInfoCookieKey);
    if (!str) { return null; }

    return new UserInfo(JSON.parse(str));
  }

  private getSecuredCookie (name: string): string {
    const rawValue = this.cookie.get(name);
    if (!rawValue) { return null; }

    const ss = rawValue.split('.');
    if (ss.length < 2) { return null; }
    return atob(ss[0]);
  }

  public can (org: string, obj: handleObject): boolean {
    const accessToken = this.cookie.get(this.tokenCookieKey) ? this.cookie.get(this.tokenCookieKey).trim() : null;
    const userRoleString = this.cookie.get(this.userRoleCookieKey) ? this.cookie.get(this.userRoleCookieKey).trim() : null;
    if (accessToken && userRoleString) {
      const userInfo = this.getUserInfo();
      const userRole: {[organizationId: string]: string} = JSON.parse(userRoleString);
      this.permValidator = new PermValidator(userInfo, userRole);
    }
    if (!this.permValidator) {
      return false;
    }
    return this.permValidator.can(org, obj);
  }
}
