import { RoleBindings } from './role_binding';

export class User {
  id?: string;
  name?: string;
  email?: string;
  displayName?: string;
  remark: string;
  admin: boolean;
  managed: boolean;
  roleBindings?: RoleBindings;
  createdAt?: string;
  lastModified?: string;
}

// used for random password
export class CreatedUser {
  id?: string;
  name?: string;
  email?: string;
  displayName?: string;
  remark: string;
  admin?: boolean;
  managed: boolean;
  password?: string;
  roleBindings: RoleBindings;
  createdAt?: string;
  lastModified?: string;
}

export class Users {
  totalCount?: number;
  edge?: {
    nodes?: User[];
  };
}

// default user domain tanzu.local
export class CreateUser {
  name: string;
  displayName: string;
  admin?: boolean;
  password?: string;
}

export class UpdateUser {
  id: string;
  name?: string;
  displayName?: string;
  admin?: boolean;
  oldPassword?: string;
  password?: string;
}

export class DeleteUser {
  id: string;
}
