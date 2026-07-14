import { Organization } from './organization';
import {X509Certificate} from "./init";

export enum LoadBalancerType {
  UNKNOWN = 'UNKNOWN',
  AVI = 'AVI',
  EXTERNAL = 'EXTERNAL'
}
export class AviLoadBalancerSpec {
  licenseTier: string;
  preferredNetwork: string;
  seGroup: string;
}

export type LoadBalancerSpec = AviLoadBalancerSpec;

export class AviCloud {
  _loadBalancer: string;
  id: string;
  name: string;
  vtype: string;
}

export class AviSubnetIPRange {
  begin: string;
  end: string;
}

export class AviSubnet {
  cidr: string;
  staticIPRanges: AviSubnetIPRange[];
}

export class AviNetwork {
  _loadBalancer: string;
  id: string;
  name: string;
  subnets: AviSubnet[];
}

export class AviVIP {
  ip: string;
  subnet: string;
}

export class AviVsVIP {
  _loadBalancer: string;
  id: string;
  name: string;
  vips: AviVIP[];
}

export class AviNetworkProfile {
  _loadBalancer: string;
  id: string;
  name: string;
}

export class AviLoadBalancerStatus {
  _id: string;
  clouds: AviCloud[];
  networkProfiles: AviNetworkProfile[];
  networks: AviNetwork[];
  vsVips: AviVsVIP[];
}

//union LoadBalancerStatus = AviLoadBalancerStatus

export class LoadBalancer {
  id: string;
  name: string;
  description: string;
  type: LoadBalancerType;
  endpoint: string;
  caCert: string;
  insecureSkipTLSVerify: boolean;
  tlsCertChain?: X509Certificate[];
  authUser?: string;
  spec: LoadBalancerSpec;
  status: any;  //use LoadBalancerStatus
  organization: Organization;
  createdAt: string;
  lastModified: string;
}

export class LoadBalancers {
  totalCount: number;
  edge: {
    nodes: LoadBalancer[];
  }
}

export class CreateAVILoadBalancerSpec {
  preferredNetwork?: string;
  seGroup?: string;
}

export class CreateLoadBalancerSpec {
  avi?: CreateAVILoadBalancerSpec;
}

export class CreateLoadBalancer {
  name: string;
  description?: string;
  type?: LoadBalancerType;
  endpoint: string;
  caCert?: string;
  insecureSkipTLSVerify?: boolean;
  authUser?: string;
  authPassword?: string;
  spec?: CreateLoadBalancerSpec;
  organization: string;
  validate?: boolean;
  dryrun?: boolean;
}

export class UpdateAVILoadBalancerSpec {
  preferredNetwork?: string;
}

export class UpdateLoadBalancerSpec {
  avi?: UpdateAVILoadBalancerSpec;
}

export class UpdateLoadBalancer {
  id: string;
  name?: string;
  description?: string;
  endpoint?: string;
  caCert?: string;
  insecureSkipTLSVerify?: boolean;
  authUser?: string;
  authPassword?: string;
  spec?: UpdateLoadBalancerSpec;
  validate?: boolean;
  dryrun?: boolean;
}

export class DeleteLoadBalancer {
  id: string;
}
