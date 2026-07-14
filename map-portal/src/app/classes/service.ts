import { Application } from './application';
import { Cluster } from './cluster';
import { Instance } from './instance';
import { CustomComponents } from './component';

export class ServiceSpecPort {
  name: string;
  nodePort: number;
  port: number;
  protocol: string;
  targetPort: string|number;
}

export class ServiceSpec {
  selector: JSON;
  clusterIP: string;
  externalTrafficPolicy: string;
  ports: ServiceSpecPort[];
  sessionAffinity: string;
  type: string;
}

export class ServiceStatusIngress {
  ip: string;
}

export class ServiceStatus {
  ingress: ServiceStatusIngress[];
}

export class Service {
  name?: string;
  namespace?: string;
  labels?: JSON;
  spec?: ServiceSpec;
  status?: ServiceStatus;
  instance: Instance;
  component: CustomComponents;
  application?: Application;
  cluster?: Cluster;
  createdAt?: string;
  lastModified?: string;
}

export class AddServiceSpecPort {
  name?: string;
  port: number;
  protocol?: string;
  targetPort: number | string;
}

export class CreateServiceSpec {
  ports: AddServiceSpecPort[];
  type: string;
}
