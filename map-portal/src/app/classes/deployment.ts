import { Cluster } from './cluster';
import { Application } from './application';
import { Instance } from './instance';
import { CustomComponents } from './component';
import { PodTemplate, CreatePodTemplate } from './pod';

export class RollingUpdateDeployment {
  maxUnavailable: number | string;
  maxSurge: number | string;
}

export class DeploymentStrategy {
  type: string;
  rollingUpdate: RollingUpdateDeployment;
}

export class DeploymentSpec {
  progressDeadlineSeconds?: number;
  replicas?: number;
  revisionHistoryLimit?: number;
  template: PodTemplate;
  strategy?: DeploymentStrategy;
}

export class DeploymentStatus {
  availableReplicas?: number;
  readyReplicas?: number;
  replicas?: number;
  unavailableReplicas?: number;
  updatedReplicas?: number;
}

export class Deployment {
  name?: string;
  namespace?: string;
  labels?: JSON;
  spec?: DeploymentSpec;
  status?: DeploymentStatus;
  instance: Instance;
  component: CustomComponents;
  application?: Application;
  cluster?: Cluster;
  createdAt?: string;
  lastModified?: string;
}

export class CreateRollingUpdateDeployment {
  maxUnavailable?: number | string;
  maxSurge?: number | string;
}

export class CreateDeploymentStrategy {
  type: string;
  rollingUpdate?: CreateRollingUpdateDeployment;
}

export class CreateDeploymentSpec {
  progressDeadlineSeconds?: number;
  replicas?: number;
  revisionHistoryLimit?: number;
  template: CreatePodTemplate;
  strategy?: CreateDeploymentStrategy;
}
