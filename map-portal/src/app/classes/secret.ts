import { Cluster } from './cluster';
import { Instance } from './instance';
import { CustomComponent } from './component';
import { Application } from './application';

export class Secret {
  name?: string;
  namespace?: string;
  labels?: JSON;
  data?: JSON;
  type?: string;
  instance?: Instance;
  component?: CustomComponent;
  application?: Application;
  cluster?: Cluster;
  createdAt?: string;
  lastModified?: string;
}

export class CreateSecret {
  name: string;
  namespace: string;
  data?: JSON;
  type?: string;
  cluster: string;
}

export class UpdateSecret {
  name: string;
  namespace: string;
  data?: JSON;
  cluster: string;
}

export class DeleteSecret {
  name: string;
  namespace: string;
  cluster: string;
}
