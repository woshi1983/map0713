import { Organization } from './organization';

export const STATICPERMISSION = {
  ADMIN: { // role name
    APPLICATION: 'FULL',
    OVERVIEW: 'READ',
    AUDITLOG: 'FULL',
    CLUSTER: 'FULL',
    CLUSTER_PROVIDER: 'WRITE',
    CLUSTER_RESOURCE: 'FULL',
    COMPONENT: 'FULL',
    IMAGE_REGISTRY: 'READ',
    INSTANCE: 'FULL',
    LOAD_BALANCER: 'READ',
    ORGANIZATION: 'WRITE', // ONLY WRITE OWNED
    ROUTE: 'FULL',
    USER: 'READ'
  },
  USER: {
    APPLICATION: 'FULL',
    OVERVIEW: 'READ',
    AUDITLOG: 'READ',
    CLUSTER: 'READ',
    CLUSTER_PROVIDER: 'WRITE',
    CLUSTER_RESOURCE: 'FULL',
    COMPONENT: 'FULL',
    IMAGE_REGISTRY: 'READ',
    INSTANCE: 'FULL',
    LOAD_BALANCER: 'READ',
    ORGANIZATION: 'READ',
    ROUTE: 'FULL',
    USER: 'READ'
  },
  GUEST: {
    APPLICATION: 'READ',
    OVERVIEW: 'READ',
    AUDITLOG: 'READ',
    CLUSTER: 'READ',
    CLUSTER_PROVIDER: 'WRITE',
    CLUSTER_RESOURCE: 'READ',
    COMPONENT: 'READ',
    IMAGE_REGISTRY: 'READ',
    INSTANCE: 'READ',
    LOAD_BALANCER: 'READ',
    ORGANIZATION: 'READ',
    ROUTE: 'READ',
    USER: 'READ'
  }
}

export enum handleObject {
    'APPLICATION' = 'APPLICATION',
    'OVERVIEW' = 'OVERVIEW',
    'AUDITLOG' = 'AUDITLOG',
    'CLUSTER' = 'CLUSTER',
    'CLUSTER_PROVIDER' = 'CLUSTER_PROVIDER',
    'CLUSTER_RESOURCE' = 'CLUSTER_RESOURCE',
    'COMPONENT' = 'COMPONENT',
    'IMAGE_REGISTRY' = 'IMAGE_REGISTRY',
    'INSTANCE' = 'INSTANCE',
    'LOAD_BALANCER' = 'LOAD_BALANCER',
    'ORGANIZATION' = 'ORGANIZATION',
    'ROUTE' = 'ROUTE',
    'USER' = 'USER'
}


export class PermValidator {
  role?: {[organizationId: string]: string};
  superAdmin?: boolean = false;

  constructor (info: UserInfo, role: {[organizationId: string]: string}) {
    this.role = role ? role : {};
    this.superAdmin = !!info.admin;
  }

  public can (org: string, obj: handleObject): boolean {
    // superadmin can do everything
    if (this.superAdmin) {
      return true;
    }
    // only superadmin can handle org create/update/delete & force delete route/instancefed/cluster
    if (org === 'NULL') {
      if (obj === 'ORGANIZATION' || obj === 'ROUTE' || obj === 'CLUSTER' || obj === 'INSTANCE') {
        return false;
      }
    }
    const userRole = this.role[org]?.toUpperCase();

    // TEMPORARY: For system config menu visibility, always return true for CLUSTER_PROVIDER
    // This allows admin users to see the System menu even if roleBindings are not properly loaded
    if (obj === 'CLUSTER_PROVIDER') {
      return this.superAdmin || (userRole !== undefined && userRole !== 'GUEST');
    }

    switch (userRole) {
      case 'ADMIN': {
        const adminPromission = STATICPERMISSION.ADMIN[obj]
        return !(adminPromission === 'READ');
      }
      case 'USER': {
        const userPromission = STATICPERMISSION.USER[obj]
        return !(userPromission === 'READ');
      }
      case 'GUEST': {
        const guestPromission = STATICPERMISSION.GUEST[obj]
        return !(guestPromission === 'READ');
      }
      default:
        return false;
    }
  }
}

export class UserInfo {
  id: string;
  name: string;
  fullName: string;
  email: string;
  admin: boolean;
  def_org: Organization;

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  constructor (d?: any) {
    if (d) { Object.assign(this, d); }
  }
}
