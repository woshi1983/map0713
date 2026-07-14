export class OrganizationQuotaUsage {
  cpus: number;
  memoryMi: number;
  storageGi: number;
}

export class OrganizationQuota {
  _id: string;
  used: OrganizationQuotaUsage;
  total: OrganizationQuotaUsage;
}

export class Organization {
  id: string;
  name: string;
  description?: string;
  avatar?: string;
  link?: string;
  quota?: OrganizationQuota;
  createdAt?: string;
  lastModified?: string;
}

export class Organizations {
  // page: PageInfo!
  // filters: [FilterInfo!]
  totalCount: number;
  edge: {
    nodes: Organization[]
  }
}

export class QuotaResult {
  cpu?: {
   used?: number;
   total?: number;
   free?: number;
   percent?: number;
  };

  memory?: {
   used?: number;
   total?: number;
   free?: number;
   percent?: number;
  };

  storage?: {
   used?: number;
   total?: number;
   free?: number;
   percent?: number;
  }
}

export class OrganizationQuotaInput {
  // zero or negative value means unlimited (default).
  cpus?: number;
  memoryMi?: number;
  storageGi?: number;
}

export class CreateOrganization {
  name: string;
  description: string;
  avatar?: string;
  link?: string;
  quota?: OrganizationQuotaInput;
}

export class UpdateOrganization {
  id: string;
  name: string;
  description: string;
  avatar?: string;
  link?: string;
  quota?: OrganizationQuotaInput;
}

export class DeleteOrganization {
  id: string;
}
