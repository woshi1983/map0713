import { Cluster, Clusters } from './cluster';
import { Organization } from './organization';

export enum ClusterExtensionName {
  TBS = 'TBS',
  OCTANT = 'OCTANT',
  CONTOUR = 'CONTOUR',
  ELASTICSEARCH = 'ELASTICSEARCH',
  FLUENTBIT = 'FLUENTBIT',
  GRAFANA = 'GRAFANA',
  PROMETHEUS = 'PROMETHEUS',
  OTHER = 'OTHER',
  WEB_TERMINAL = 'WEB_TERMINAL',
  CSI_DRIVER_NFS = 'CSI_DRIVER_NFS'
}

export enum ClusterExtensionState {
  UNKNOWN = 'UNKNOWN',
  ENABLING = 'ENABLING',
  ENABLED = 'ENABLED',
  DISABLING = 'DISABLING'
}

export enum ClusterExtensionStatus {
  PENDING = 'PENDING',
  PROVISIONING = 'PROVISIONING',
  PROVISIONED = 'PROVISIONED',
  FAILED = 'FAILED'
}

export class ClusterExtension {
  id: string;
  name: ClusterExtensionName;
  endpoints: string[];
  spec?: JSON;
  state: ClusterExtensionState;
  status?: ClusterExtensionStatus;
  related?: Clusters;
  cluster?: Cluster;
  organization?: Organization;
  createdAt?: string;
  lastModified?: string;
}

export class ClusterExtensions {
  totalCount: number;
  edge: {
    nodes: ClusterExtension[];
  }
}

export class EnableClusterExtension {
  name: string;
  spec?: JSON;
  related?: string[];
  cluster: string;
}

export class ModifyClusterExtension {
  id: string;
  spec?: JSON;
  addRelated?: string[];
  removeRelated?: string[];
}

export class DisableClusterExtension {
  id: string;
}
