import { Instance } from './instance';
import { Route } from './route';
import { CustomComponent } from './component';
import { Application } from './application';
import { Organization } from './organization';

export enum InstanceFederationStatus {
  PENDING = 'PENDING',
  PROVISIONING = 'PROVISIONING',
  PROVISIONED = 'PROVISIONED',
  DELETING = 'DELETING'
}

export class InstanceFederation {
  id: string;
  name: string;
  description: string;
  status: InstanceFederationStatus;
  instances: Instance[];
  routes: Route[];
  component: CustomComponent;
  application: Application;
  organization: Organization;
  createdAt: string;
  lastModified: string;
}

export class InstanceFederations {
  totalCount: number;
  edge: {
    nodes: InstanceFederation[];
  }
}

export class AddInstance {
  autoRestart?: boolean;
  release: string;
  namespace?: string;
  cluster: string;
}

export class ModifyInstance {
  autoRestart?: boolean;
  id: string;
  release?: string;
}

export class RemoveInstance {
  id: string;
  force?: boolean;
}

export class UpdateInstanceFederation {
  id: string;
  description?: string;
  addInstances?: AddInstance[];
  modifyInstances?: ModifyInstance[];
  removeInstances?: RemoveInstance[];
}
