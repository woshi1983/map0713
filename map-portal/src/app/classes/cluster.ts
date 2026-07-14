import { ImageRegistry } from './repository';
import { LoadBalancer } from './load_balancer';
import { Namespace } from './namespace';
import { ClusterProvider } from './cluster_provider';
import { Organization } from './organization';
import { ClusterExtensions } from './cluster_extension';
import {X509Certificate} from "./init";

export enum ClusterStatus {
  PENDING = 'PENDING',
  INITIALIZING = 'INITIALIZING',
  HEALTHY = 'HEALTHY',
  UNHEALTHY = 'UNHEALTHY',
  MAINTENANCE = 'MAINTENANCE'
}

export enum ClusterState {
  UNKNOWN = 'UNKNOWN',
  CREATING = 'CREATING',
  RUNNING = 'RUNNING',
  UPDATING = 'UPDATING',
  FAILED = 'FAILED',
  DELETING = 'DELETING'
}

export enum ClusterType {
  ATTACHED = 'ATTACHED',
  PROVISIONED = 'PROVISIONED'
}

export enum ClusterClass {
  'UNKNOWN' = 'UNKNOWN',
  'TKGS' = 'TKGS',
  'TKGM' = 'TKGM',
  'TKG2M' = 'TKG2M'
}

export class ClusterImageRegistry {
  _id: string;
  imageRegistry: ImageRegistry;
  default: boolean;
}

export class ClusterImageRegistryNodes {
  nodes: ClusterImageRegistry[];
}

export class ClusterImageRegistries {
  _cluster: string;
  totalCount: number;
  edge: ClusterImageRegistryNodes;
}

// union ClusterSpec = TKGSClusterSpec | TKGMClusterSpec
export class Cluster {
  id?: string;
  name?: string;
  description?: string;
  type?: string;
  class?: ClusterClass;
  endpoint?: string;
  caCert?: string;
  insecureSkipTLSVerify?: boolean;
  tlsCertChain?: X509Certificate[];
  version?: string;
  status?: ClusterStatus;
  state?: ClusterState;
  spec?: any;
  extensions?: ClusterExtensions;
  kubeConfig?: string;
  sshPrivateKey?: string;
  namespaces?: Namespace[];
  imageRegistries?: ClusterImageRegistries;
  loadBalancer?: LoadBalancer;
  provider?: ClusterProvider;
  organization?: Organization;
  createdAt?: string;
  lastModified?: string;
}

export class TKGSClusterSpec {
  namespace: string;
  cni?: string;
  podCIDR?: string;
  svcCIDR?: string;
  svcDomain?: string;
  defStorage?: string;
  ctlPlaneClass?: string;
  ctlPlaneCount?: number;
  ctlPlaneStorage?: string;
  ctlPlaneStorageCapGi?: number;
  dataPlaneClass?: string;
  dataPlaneCount?: number;
  dataPlaneStorage?: string;
  dataPlaneStorageCapGi?: number;
}

export class TKGMClusterSpec {
  namespace: string;
  cni: string;
  podCIDR: string;
  svcCIDR: string;
  ctlPlaneCount: number;
  ctlPlaneCPUs: number;
  ctlPlaneMemMi: number;
  ctlPlaneStorageCapGi: number;
  dataPlaneCount: number;
  dataPlaneCPUs: number;
  dataPlaneMemMi: number;
  dataPlaneStorageCapGi: number;
  /*
  """env contains a full set of environment variables that were used for
  provisioning a TKGm workload cluster.
  """
  */
  env: any;
}

export class TKG2MClusterSpec {
  namespace: string;
  osName: string;
  cni: string;
  clusterCIDR: string;
  serviceCIDR: string;
  ctlPlaneReplicas: number;
  ctlPlaneNumCpus: number;
  ctlPlaneMemMiB: number;
  ctlPlaneDiskGiB: number;
  workerNodePool0Replicas: number;
  workerNodePool0Zone: string;
  workerNodePool1Replicas: number;
  workerNodePool1Zone: string;
  workerNodePool2Replicas: number;
  workerNodePool2Zone: string;
  workerNumCpus: number;
  workerMemMiB: number;
  workerDiskGiB: number;
  vsphereNetwork: string;
  vsphereResourcePool: string;
  vsphereFolder: string;
  vsphereDatastore: string;
  aviServiceEngineGroup: string;
  aviDataNetwork: string;
  aviDataNetworkCidr: string;
}

export class Clusters {
  totalCount: number;
  edge: {
    nodes: Cluster[];
  }
}

export class AddClusterExtension {
  name: string;
  spec?: JSON;
  related?: string[];
}

export class CreateTKGSClusterSpec {
  namespace?: string;
  cni?: string;
  podCIDR?: string;
  svcCIDR?: string;
  svcDomain?: string;
  defStorage?: string;
  ctlPlaneClass?: string;
  ctlPlaneCount?: number;
  ctlPlaneStorage?: string;
  ctlPlaneStorageCapGi?: number;
  dataPlaneClass?: string;
  dataPlaneCount?: number;
  dataPlaneStorage?: string;
  dataPlaneStorageCapGi?: number;
}

export class CreateTKGMClusterSpec {
  namespace?: string;
  cni?: string;
  podCIDR?: string;
  svcCIDR?: string;
  ctlPlaneCount?: number;
  ctlPlaneCPUs?: number;
  ctlPlaneMemMi?: number;
  ctlPlaneStorageCapGi?: number;
  dataPlaneCount?: number;
  dataPlaneCPUs?: number;
  dataPlaneMemMi?: number;
  dataPlaneStorageCapGi?: number;
  env?: JSON;
}

export class CreateTKG2MClusterSpec {
  namespace?: string;
  osName?: string;
  cni?: string;
  clusterCIDR?: string;
  serviceCIDR?: string;
  ctlPlaneReplicas?: number;
  ctlPlaneNumCpus?: number;
  ctlPlaneMemMiB?: number;
  ctlPlaneDiskGiB?: number;
  workerNodePool0Replicas?: number;
  workerNodePool0Zone?: string;
  workerNodePool1Replicas?: number;
  workerNodePool1Zone?: string;
  workerNodePool2Replicas?: number;
  workerNodePool2Zone?: string;
  workerNumCpus?: number;
  workerMemMiB?: number;
  workerDiskGiB?: number;
  vsphereNetwork?: string;
  vsphereResourcePool?: string;
  vsphereFolder?: string;
  vsphereDatastore?: string;
  aviServiceEngineGroup?: string;
  aviDataNetwork?: string;
  aviDataNetworkCidr?: string;
}

export class CreateClusterSpec {
  tkgs?: CreateTKGSClusterSpec;
  tkgm?: CreateTKGMClusterSpec;
  tkg2m?: CreateTKG2MClusterSpec;
}

export class CreateCluster {
  name?: string;
  description?: string;
  version?: string;
  spec?: CreateClusterSpec;
  extensions?: AddClusterExtension[];
  provider?: string;
}

export class AddClusterImageRegistry {
  id: string;
  default?: boolean;
}

export class AttachClusterSpec {
  rootCAs: string[];
}

export class AttachCluster {
  name: string;
  description?: string;
  endpoint: string;
  caCert?: string;
  insecureSkipTLSVerify?: boolean;
  authToken: string;
  spec?: AttachClusterSpec;
  extensions?: AddClusterExtension[];
  imageRegistries: AddClusterImageRegistry[];
  loadBalancer?: string;
  dryrun?: boolean;
  validate?: boolean;
}

export class UpdateTKGSClusterSpec {
  ctlPlaneClass?: string;
  ctlPlaneCount?: number;
  dataPlaneClass?: string;
  dataPlaneCount?: number;
}

export class UpdateTKGMClusterSpec {
  ctlPlaneCount?: number;
  ctlPlaneCPUs?: number;
  ctlPlaneMemMi?: number;
  ctlPlaneStorageCapGi?: number;
  dataPlaneCount?: number;
  dataPlaneCPUs?: number;
  dataPlaneMemMi?: number;
  dataPlaneStorageCapGi?: number;
}

export class UpdateTKG2MClusterSpec {
  ctlPlaneReplicas?: number;
  ctlPlaneNumCpus?: number;
  ctlPlaneMemMiB?: number;
  ctlPlaneDiskGiB?: number;
  workerNodePool0Replicas?: number;
  workerNodePool0Zone?: string;
  workerNodePool1Replicas?: number;
  workerNodePool1Zone?: string;
  workerNodePool2Replicas?: number;
  workerNodePool2Zone?: string;
  workerNumCpus?: number;
  workerMemMiB?: number;
  workerDiskGiB?: number;
}

export class UpdateClusterSpec {
  tkgs?: UpdateTKGSClusterSpec;
  tkgm?: UpdateTKGMClusterSpec;
  tkg2m?: UpdateTKG2MClusterSpec;
}

export class UpdateCluster {
  id?: string;
  description?: string;
  // Only allowed on PROVISIONED cluster
  version?: string;
  // Only allowed on PROVISIONED cluster
  spec?: UpdateClusterSpec;
  // TODO: 该参数是否还需要？
  maintenance?: boolean;
  // reinit cluster
  recover?: boolean;
}

export class DeleteCluster {
  id: string;
  force?: boolean;
}
