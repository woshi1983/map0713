import { Organization } from './organization';
import { User } from './user';

export enum Role {
  GUEST = 'GUEST',
  USER = 'USER',
  ADMIN ='ADMIN'
}

export class RoleBinding {
  id: string;
  user: User;
  role: Role;
  organization: Organization;
}

export class RoleBindings {
  totalCount: number;
  edge: {
    nodes: RoleBinding[];
  }
}

export class CreateRoleBinding {
  user: string;
  role: string;
  organization: string;
}

export class UpdateRoleBinding {
  id: string;
  role: string;
}

export class DeleteRoleBinding {
  id: string;
}
