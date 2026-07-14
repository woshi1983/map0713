import { Injectable } from '@angular/core';
import {ActivatedRouteSnapshot, CanActivate, CanActivateChild, Router} from '@angular/router';
import { Observable } from 'rxjs';
import {environment} from "../../../environments/environment";

@Injectable({
  providedIn: 'root'
})
export class FeatureGuard implements CanActivate, CanActivateChild {

  constructor (
    private router: Router
  ){}

  canActivate(route: ActivatedRouteSnapshot){
    const param = route.data.permissionParam.resource;
    if (environment.features.hasOwnProperty(param)) {
      if(!environment.features[param]){
        this.router.navigate(['/portal']);
      }
      return environment.features[param];
    }
    return true;
  }

  canActivateChild (route: ActivatedRouteSnapshot): Observable<boolean> | boolean {
    return this.canActivate(route);
  }

}
