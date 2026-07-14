import { Service } from './service';
import { Cluster } from './cluster';
import { Release } from './release';
import { InstanceFederation } from './instance_federation';
import { CustomComponents } from './component';
import { Application } from './application';
import { Organization } from './organization';

export enum InstanceState {
  PENDING = 'PENDING',
  PROVISIONING = 'PROVISIONING',
  SUCCESS = 'SUCCESS',
  FAILURE = 'FAILURE',
  DELETING = 'DELETING'
}

export class InstanceStatusTask {
  name: string;
  namespace: string;
  kind: string;
  state: InstanceState;
  message: string;
}

export class InstanceStatus {
  state: InstanceState;
  tasks: InstanceStatusTask[];
}

export class InstanceWorkloadHealth {
  name: string;
  namespace: string;
  kind: string;
  readyReplicas: number;
  replicas: number;
}

export class InstanceHealth {
  available: number;
  ready: number;
  workloads: InstanceWorkloadHealth[];
}

export class Instance {
  id: string;
  name: string;
  namespace: string;
  autoRestart: boolean;
  manifest: any[];
  status: InstanceStatus;
  health: InstanceHealth;
  services: Service[];
  cluster: Cluster;
  release: Release;
  federation: InstanceFederation;
  component: CustomComponents;
  application: Application;
  organization: Organization;
  createdAt: string;
  lastModified: string;
}

export class Instances {
  totalCount: number;
  edge: {
    nodes: Instance[];
  }
}
