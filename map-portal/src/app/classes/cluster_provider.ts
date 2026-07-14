import { ImageRegistry } from './repository';
import { LoadBalancer } from './load_balancer';
import { Organization } from './organization';
import { ClusterClass } from './cluster';
import { X509Certificate } from './init';

// union ClusterProviderSpec = TKGSClusterProviderSpec | TKGMClusterProviderSpec

export class ClusterProviderImageRegistry {
  _id: string;
  imageRegistry: ImageRegistry;
  default: boolean;
}

export class ClusterProviderImageRegistryNodes {
  nodes: ClusterProviderImageRegistry[];
}

export class ClusterProviderImageRegistries {
  _provider: string;
  totalCount: number;
  edge: ClusterProviderImageRegistryNodes;
}

export class ClusterProvider {
  id: string;
  name: string;
  description: string;
  class: ClusterClass;
  spec: any;
  tlsCertChain?: X509Certificate[];
  imageRegistries: ClusterProviderImageRegistries;
  loadBalancer: LoadBalancer;
  organization: Organization;
  createdAt: string;
  lastModified: string;
}

export class ClusterProviders {
  totalCount: number;
  edge: {
    nodes: ClusterProvider[]
  }
}

export class ClusterProviderSpecTrustedCA {
  name: string;
  cert: string;
}

export class TKGSClusterProviderSpecNamespace {
  _clusterProvider: string;
  name: string;
  labels: JSON;
  storages?: string[];
  createdAt: string;
  lastModified: string;
}

export class VMClass {
  name?: string;
  spec?: {
    hardware?: {
      cpus?: number;
      memory?: string;
    }
  };
}

export class TKGSClusterProviderSpec {
  _id: string;
  endpoint: string;
  caCert: string;
  insecureSkipTLSVerify: boolean;
  authUser: string;
  defaultNamespace: string;
  vmClasses: VMClass[];
  clusterReleases: string[];
  namespaces: TKGSClusterProviderSpecNamespace[];
  trustedCAs: ClusterProviderSpecTrustedCA[];
}

export class TKGMClusterProviderSpecNamespace {
  _clusterProvider: string;
  name: string;
  labels: JSON;
  createdAt: string;
  lastModified: string;
}

export class TKGMClusterProviderSpec {
  _id: string;
  endpoint: string;
  defaultNamespace: string;
  defaultEnv: JSON;
  clusterReleases: string[];
  namespaces: TKGMClusterProviderSpecNamespace[];
  trustedCAs: ClusterProviderSpecTrustedCA[];
}

export class AddClusterProviderTrustedCA {
  name: string;
  cert: string;
}

export class CreateTKGSClusterProviderSpec {
  endpoint: string;
  caCert?: string;
  insecureSkipTLSVerify?: boolean;
  authUser: string;
  authPassword: string;
  defaultNamespace?: string;
  trustedCAs?: AddClusterProviderTrustedCA[];
}

export class CreateTKGMClusterProviderSpec {
  kubeConfig: string;
  defaultNamespace?: string;
  defaultEnv?: JSON;
  /*
  """certificates of root CA that will be only used for configuring TBS.

  These certificates will not be injected to the certificate pool on the hosts.

  To inject those CA certificates into the provisioned hosts, you MUST modify
  the template file that WILL BE used for provisioning the management cluster.
  """ */
  trustedCAs?: AddClusterProviderTrustedCA[];
}

export class AddClusterProviderImageRegistry {
  id: string;
  default?: boolean;
}

export class CreateTKGSClusterProvider {
  name: string;
  description?: string;
  spec: CreateTKGSClusterProviderSpec;
  imageRegistries: AddClusterProviderImageRegistry[];
  loadBalancer?: string;
  dryrun?: boolean;
  validate?: boolean;
}

export class CreateTKGMClusterProvider {
  name: string;
  description?: string;
  spec: CreateTKGMClusterProviderSpec;
  imageRegistries: AddClusterProviderImageRegistry[];
  loadBalancer?: string;
  dryrun?: boolean;
  validate?: boolean;
}

export class CreateTKG2MClusterProviderSpec {
  kubeConfig: string;
  defaultNamespace?: string;
  trustedCAs?: AddClusterProviderTrustedCA[];
  vsphereNetwork?: string;
  vsphereResourcePool?: string;
  vsphereFolder?: string;
  vsphereDatastore?: string;
  aviServiceEngineGroup?: string;
  aviDataNetwork?: string;
  aviDataNetworkCidr?: string;
}

export class CreateTKG2MClusterProvider {
  name: string;
  description?: string;
  spec: CreateTKG2MClusterProviderSpec;
  imageRegistries: AddClusterProviderImageRegistry[];
  loadBalancer?: string;
  dryrun?: boolean;
  validate?: boolean;
}

export class RemoveClusterProviderTrustedCA {
  name: string;
}

export class UpdateTKGSClusterProviderSpec {
  endpoint?: string;
  caCert?: string;
  insecureSkipTLSVerify?: boolean;
  authUser?: string;
  authPassword?: string;
  defaultNamespace?: string;
  addTrustedCAs?: AddClusterProviderTrustedCA[];
  removeTrustedCAs?: RemoveClusterProviderTrustedCA[];
}

export class UpdateTKGMClusterProviderSpec {
  kubeConfig?: string;
  defaultNamespace?: string;
  defaultEnv?: JSON;
  addTrustedCAs?: AddClusterProviderTrustedCA[];
  removeTrustedCAs?: RemoveClusterProviderTrustedCA[];
}

export class RemoveClusterProviderImageRegistry {
  id: string;
}

export class UpdateTKGSClusterProvider {
  id: string;
  description?: string;
  spec?: UpdateTKGSClusterProviderSpec;
  addImageRegistries?: AddClusterProviderImageRegistry[];
  removeImageRegistries?: RemoveClusterProviderImageRegistry[];
  loadBalancer?: string;
  dryrun?: boolean;
  validate?: boolean;
}

export class UpdateTKGMClusterProvider {
  id: string;
  description?: string;
  spec: UpdateTKGMClusterProviderSpec;
  addImageRegistries?: AddClusterProviderImageRegistry[];
  removeImageRegistries?: RemoveClusterProviderImageRegistry[];
  loadBalancer?: string;
  dryrun?: boolean;
  validate?: boolean;
}

export class UpdateTKG2MClusterProviderSpec {
  kubeConfig?: string;
  defaultNamespace?: string;
  addTrustedCAs?: AddClusterProviderTrustedCA[];
  removeTrustedCAs?: RemoveClusterProviderTrustedCA[];
  vsphereNetwork?: string;
  vsphereResourcePool?: string;
  vsphereFolder?: string;
  vsphereDatastore?: string;
  aviServiceEngineGroup?: string;
  aviDataNetwork?: string;
  aviDataNetworkCidr?: string;
}

export class UpdateTKG2MClusterProvider {
  id: string;
  description?: string;
  spec: UpdateTKG2MClusterProviderSpec;
  addImageRegistries?: AddClusterProviderImageRegistry[];
  removeImageRegistries?: RemoveClusterProviderImageRegistry[];
  loadBalancer?: string;
  dryrun?: boolean;
  validate?: boolean;
}

export class DeleteClusterProvider {
  id: string;
}
