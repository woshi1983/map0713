import { Cluster } from './cluster';
import { Instance } from './instance';
import { CustomComponent } from './component';
import { Application } from './application';

export class ConfigMap {
  name?: string;
  namespace?: string;
  labels?: JSON;
  data?: JSON;
  instance?: Instance;
  component?: CustomComponent;
  application?: Application;
  cluster?: Cluster;
  createdAt?: string;
  lastModified?: string;
}

export class CreateConfigMap {
  name: string;
  namespace: string;
  data: JSON;
  cluster: string;
}

export class UpdateConfigMap {
  name: string;
  namespace: string;
  data: JSON;
  cluster: string;
}

export class DeleteConfigMap {
  name: string;
  namespace: string;
  cluster: string;
}
