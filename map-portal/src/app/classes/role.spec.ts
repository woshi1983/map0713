import { handleObject, PermValidator, UserInfo } from './role';

describe('PermValidator', () => {
  const orgId = 'organization-1';

  it('allows read-only navigation without granting write access', () => {
    for (const role of ['ADMIN', 'USER', 'GUEST']) {
      const permissions = new PermValidator(new UserInfo({admin: false}), {[orgId]: role});
      expect(permissions.canRead(orgId, handleObject.OVERVIEW)).toBeTrue();
      expect(permissions.can(orgId, handleObject.OVERVIEW)).toBeFalse();
    }
  });

  it('does not expose restricted system actions to a guest', () => {
    const permissions = new PermValidator(new UserInfo({admin: false}), {[orgId]: 'GUEST'});
    expect(permissions.canRead(orgId, handleObject.CLUSTER_PROVIDER)).toBeFalse();
    expect(permissions.can(orgId, handleObject.CLUSTER)).toBeFalse();
    expect(permissions.can('NULL', handleObject.ORGANIZATION)).toBeFalse();
  });

  it('does not grant access without an organization role', () => {
    const permissions = new PermValidator(new UserInfo({admin: false}), {});
    expect(permissions.canRead(orgId, handleObject.OVERVIEW)).toBeFalse();
  });
});
