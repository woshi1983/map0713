import { ActivatedRouteSnapshot, Router } from '@angular/router';
import { handleObject } from '../../classes/role';
import { AuthService } from '../../services/auth.service';
import { OrganizationService } from '../../services/organization.service';
import { OrgRolePermissionGuard } from './org-role-permission.guard';

describe('OrgRolePermissionGuard', () => {
  const org = 'project-id';
  let router: jasmine.SpyObj<Router>;
  let auth: jasmine.SpyObj<AuthService>;
  let guard: OrgRolePermissionGuard;

  beforeEach(() => {
    router = jasmine.createSpyObj<Router>('Router', ['navigate']);
    auth = jasmine.createSpyObj<AuthService>('AuthService', ['can', 'canRead']);
    guard = new OrgRolePermissionGuard(router, auth, {getSelectedOrgId: org} as OrganizationService);
  });

  function route(resource: handleObject): ActivatedRouteSnapshot {
    return {data: {permissionParam: {resource}}} as unknown as ActivatedRouteSnapshot;
  }

  it('lets a project role open the read-only overview', () => {
    auth.can.and.returnValue(false);
    auth.canRead.and.returnValue(true);

    expect(guard.canActivateChild(route(handleObject.OVERVIEW))).toBeTrue();
    expect(auth.canRead).toHaveBeenCalledWith(org, handleObject.OVERVIEW);
    expect(router.navigate).not.toHaveBeenCalled();
  });

  it('continues to block the restricted system menu', () => {
    auth.can.and.returnValue(false);
    auth.canRead.and.returnValue(true);

    expect(guard.canActivateChild(route(handleObject.CLUSTER_PROVIDER))).toBeFalse();
    expect(auth.canRead).not.toHaveBeenCalled();
    expect(router.navigate).toHaveBeenCalledWith(['/portal/no-permission']);
  });
});
