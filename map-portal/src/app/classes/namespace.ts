import { Cluster } from './cluster';

export class Namespace {
  name?: string;
  labels?: JSON;
  cluster?: Cluster;
  createdAt?: string;
  lastModified?: string;
}

export class CreateNamespace {
  name: string;
  labels?: JSON;
  cluster: string;
}

export class UpdateNamespace {
  name: string;
  labels?: JSON;
  cluster: string;
}

export class DeleteNamespace {
  name: string;
  cluster: string;
}
