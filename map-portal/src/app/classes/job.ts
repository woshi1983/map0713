import { Cluster } from './cluster';
import { Application } from './application';
import { CustomComponent } from './component';
import { PodTemplate } from './pod';
import { Instance } from './instance';

export enum JobStatus {
  UNKNOWN = 'UNKNOWN',
  PENDING = 'PENDING',
  RUNNING = 'RUNNING',
  SUCCESS = 'SUCCESS',
  FAILURE = 'FAILURE',
}

export class JobSpec {
  activeDeadlineSeconds: number;
  backoffLimit: number;
  completions: number;
  parallelism: number;
  template: PodTemplate;
}

export class Job {
  name: string;
  namespace: string;
  labels: JSON;
  spec: JobSpec;
  status: JobStatus;
  instance: Instance;
  component: CustomComponent;
  application: Application;
  cluster: Cluster;
  createdAt: string;
  lastModified: string;
}
