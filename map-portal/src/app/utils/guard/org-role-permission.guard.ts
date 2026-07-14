import { Injectable } from '@angular/core';
import { ActivatedRouteSnapshot, CanActivate, CanActivateChild, Router } from '@angular/router';
import { Observable } from 'rxjs';
import { AuthService } from '../../services/auth.service';
import { OrganizationService } from '../../services/organization.service';

@Injectable({
  providedIn: 'root'
})
export class OrgRolePermissionGuard implements CanActivate, CanActivateChild {
  constructor (
    private router: Router,
    private authService: AuthService,
    private orgService: OrganizationService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  orgId: string;

  canActivate (route: ActivatedRouteSnapshot): Observable<boolean> | boolean {
    const object = route.data.permissionParam.resource;
    if (!this.authService.can(this.orgId, object)) {
      this.router.navigate(['/portal/no-permission']);
      return false;
    }
    return true;
  }

  canActivateChild (route: ActivatedRouteSnapshot): Observable<boolean> | boolean {
    return this.canActivate(route);
  }
}
