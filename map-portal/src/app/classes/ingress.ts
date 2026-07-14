import { Instance } from './instance';
import { CustomComponent } from './component';
import { Application } from './application';
import { Cluster } from './cluster';

export class IngressClass {
  name: string;
  labels: JSON;
  instance: Instance;
  component: CustomComponent;
  application: Application;
  cluster: Cluster;
  createdAt: string;
  lastModified: string;
}

export class IngressTLS {
  hosts: string[];
  secret: string;
}

export class ServiceBackendPort {
  name: string;
  number: number;
}
export class IngressServiceBackend {
  name: string;
  port: ServiceBackendPort;
}

export class IngressBackend {
  service: IngressServiceBackend;
}

export class HTTPIngressPath {
  path: string;
  pathType: string;
  backend: IngressBackend;
}

export class HTTPIngressRuleValue {
  paths: HTTPIngressPath[];
}

export class IngressRule {
  host: string;
  http: HTTPIngressRuleValue;
}

export class IngressSpec {
  ingressClass: string;
  tls: IngressTLS[];
  rules: IngressRule[];
}
export class IngressStatusIngress {
  ip: string;
  hostname: string;
}

export class IngressStatus {
  ingress: IngressStatusIngress[];
}

export class Ingress {
  name: string;
  namespace: string;
  annotations: JSON;
  labels: JSON;
  spec: IngressSpec;
  status: IngressStatus;
  instance: Instance;
  component: CustomComponent;
  application: Application;
  cluster: Cluster;
  createdAt: string;
  lastModified: string;
}

export class IngressTLSInput {
  hosts?: string[];
  secret?: string;
}

export class ServiceBackendPortInput {
  number?: number;
}

export class IngressServiceBackendInput {
  name?: string;
  port?: ServiceBackendPortInput;
}

export class IngressBackendInput {
  service?: IngressServiceBackendInput;
}

export class HTTPIngressPathInput {
  path?: string;
  pathType?: string;
  backend: IngressBackendInput;
}

export class HTTPIngressRuleValueInput {
  paths: HTTPIngressPathInput[];
}

export class IngressRuleInput {
  host?: string;
  http?: HTTPIngressRuleValueInput;
}

export class IngressSpecInput {
  ingressClass?: string;
  tls?: IngressTLSInput[];
  rules?: IngressRuleInput[];
}

export class CreateIngress {
  name: string;
  namespace: string;
  annotations: JSON;
  labels: JSON;
  spec: IngressSpecInput;
  cluster: string;
}

export class UpdateIngress {
  name: string;
  namespace: string;
  annotations: JSON;
  labels: JSON;
  spec: IngressSpecInput;
  cluster: string;
}

export class DeleteIngress {
  name: string;
  namespace: string;
  cluster: string;
}
