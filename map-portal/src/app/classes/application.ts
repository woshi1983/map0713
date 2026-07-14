import { Organization } from './organization';

export class Application {
  id: string;
  name: string;
  description: string;
  avatar: string;
  organization: Organization;
  createdAt: string;
  lastModified: string;
}

export class Applications {
  totalCount: number;
  edge: {
    nodes: Application[];
  }
}

export class CreateApplication {
  name: string;
  description?: string;
  avatar?: string;
  organization: string;
}

export class UpdateApplication {
  id: string;
  description?: string;
  avatar?: string;
}

export class DeleteApplication {
  id: string;
}
