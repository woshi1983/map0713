import { Cluster } from './cluster';
import { Instance } from './instance';
import { CustomComponent } from './component';
import { Application } from './application';

export class PersistentVolumeClaimSpecResourcesLimits {
  storage: string;
}

export class PersistentVolumeClaimSpecResourcesRequests {
  storage: string;
}

export class PersistentVolumeClaimSpecResources {
  limits: PersistentVolumeClaimSpecResourcesLimits;
  requests: PersistentVolumeClaimSpecResourcesRequests;
}

export class PersistentVolumeClaimSpec {
  accessModes: string[];
  resources: PersistentVolumeClaimSpecResources;
  storageClass: string;
  volumeMode: string;
}

export class PersistentVolumeClaim {
  name?: string;
  namespace?: string;
  labels?: JSON;
  spec?: PersistentVolumeClaimSpec;
  status?: string;
  instance?: Instance;
  component?: CustomComponent;
  application?: Application;
  cluster?: Cluster;
  createdAt?: string;
  lastModified?: string;
}

export class CreatePersistentVolumeClaimSpecResourcesLimits {
  storage: string;
}

export class CreatePersistentVolumeClaimSpecResourcesRequests {
  storage: string;
}

export class CreatePersistentVolumeClaimSpecResources {
  limits?: CreatePersistentVolumeClaimSpecResourcesLimits;
  requests: CreatePersistentVolumeClaimSpecResourcesRequests;
}

export class CreatePersistentVolumeClaimSpec {
  accessModes: string[];
  resources: CreatePersistentVolumeClaimSpecResources;
  storageClass?: string;
  volumeMode?: string;
}

export class CreatePersistentVolumeClaim {
  name: string;
  namespace: string;
  spec: CreatePersistentVolumeClaimSpec;
  cluster: string;
}

export class UpdatePersistentVolumeClaimSpecResourcesLimits {
  storage?: string;
}

export class UpdatePersistentVolumeClaimSpecResourcesRequests {
  storage?: string;
}

export class UpdatePersistentVolumeClaimSpecResources {
  limits?: UpdatePersistentVolumeClaimSpecResourcesLimits;
  requests?: UpdatePersistentVolumeClaimSpecResourcesRequests;
}

export class UpdatePersistentVolumeClaimSpec {
  accessModes?: string[];
  resources?: UpdatePersistentVolumeClaimSpecResources;
  storageClass?: string;
  volumeMode?: string;
}

export class UpdatePersistentVolumeClaim {
  name: string;
  namespace: string;
  spec: UpdatePersistentVolumeClaimSpec;
  cluster: string;
}

export class DeletePersistentVolumeClaim {
  name: string;
  namespace: string;
  cluster: string;
}
