import { User } from './user';
import { Organization } from './organization';

export class AuditLog {
  id: string;
  action: string;
  objectType: string;
  objectName: string;
  objectID: string;
  userName: string;
  user: User;
  organization: Organization;
  createdAt: string;
  lastModified: string;
}

export class AuditLogNodes {
  nodes: AuditLog[];
}

export class AuditLogs {
  // page: PageInfo!
  // filters: [FilterInfo!]
  totalCount: number;
  edge: AuditLogNodes;
}
