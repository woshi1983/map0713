import { Cluster } from './cluster';
import { Instance } from './instance';
import { CustomComponent } from './component';
import { Application } from './application';

export class StorageClass {
  name: string;
  labels: JSON;
  annotations: JSON;
  provisioner: string;
  parameters: JSON;
  reclaimPolicy: string;
  mountOptions: string[];
  allowVolumeExpansion: boolean;
  volumeBindingMode: string;
  instance: Instance;
  component: CustomComponent;
  application: Application;
  cluster: Cluster;
  createdAt: string;
  lastModified: string;
}

export class CreateStorageClass {
  name: string;
  labels?: JSON;
  annotations?: JSON;
  provisioner?: string;
  parameters?: JSON;
  reclaimPolicy?: string;
  mountOptions?: string[];
  allowVolumeExpansion?: boolean
  volumeBindingMode?: string;
  cluster: string;
}

export class UpdateStorageClass {
  name: string;
  labels?: JSON;
  annotations?: JSON;
  mountOptions?: string[];
  allowVolumeExpansion?: boolean;
  cluster: string;
}

export class DeleteStorageClass {
  name: string;
  cluster: string;
}
