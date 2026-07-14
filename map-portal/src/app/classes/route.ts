import { LoadBalancer, LoadBalancerType } from './load_balancer';
import { Application } from './application';
import { CustomComponent } from './component';
import { Organization } from './organization';
import { Instance } from './instance';

export enum RouteStatus {
  PROVISIONING = 'PROVISIONING',
  PROVISIONED = 'PROVISIONED',
  DELETING = 'DELETING'
}

export enum RouteProtocol {
  TCP = 'TCP',
  UDP = 'UDP'
}

export class AviRouteSpec {
  networkProfile: string;
  lbAlgorithm: string;
  lbAlgorithmHash: string;
  lbAlgorithmConsistentHashHeader: string;
  manualAllocatedVsVIP: boolean;
  persistentVsVIP: boolean;
}

export type RouteSpec = AviRouteSpec;

export class RouteBackend {
  id: string;
  name: string;
  namespace: string;
  type: string;
  ip: string;
  port: number;
  ratio: number;
  instance: Instance;
}

export class Route {
  id: string;
  name: string;
  description: string;
  type: LoadBalancerType;
  protocol: RouteProtocol;
  ip?: string;
  port: number;
  backends: RouteBackend[];
  spec: RouteSpec;
  status: RouteStatus;
  component: CustomComponent;
  application: Application;
  loadBalancer: LoadBalancer;
  organization: Organization;
  createdAt: string;
  lastModified: string;
}

export class Routes {
  totalCount: number;
  edge: {
    nodes: Route[];
  }
}

export enum LBAlgorithm {
  LB_ALGORITHM_LEAST_CONNECTIONS = 'LB_ALGORITHM_LEAST_CONNECTIONS',
  LB_ALGORITHM_ROUND_ROBIN = 'LB_ALGORITHM_ROUND_ROBIN',
  LB_ALGORITHM_FASTEST_RESPONSE = 'LB_ALGORITHM_FASTEST_RESPONSE',
  LB_ALGORITHM_CONSISTENT_HASH = 'LB_ALGORITHM_CONSISTENT_HASH',
  LB_ALGORITHM_LEAST_LOAD = 'LB_ALGORITHM_LEAST_LOAD',
  LB_ALGORITHM_FEWEST_SERVERS = 'LB_ALGORITHM_FEWEST_SERVERS',
  LB_ALGORITHM_RANDOM = 'LB_ALGORITHM_RANDOM',
  LB_ALGORITHM_FEWEST_TASKS = 'LB_ALGORITHM_FEWEST_TASKS',
  LB_ALGORITHM_NEAREST_SERVER = 'LB_ALGORITHM_NEAREST_SERVER',
  LB_ALGORITHM_CORE_AFFINITY = 'LB_ALGORITHM_CORE_AFFINITY',
  LB_ALGORITHM_TOPOLOGY = 'LB_ALGORITHM_TOPOLOGY'
}

/*
* LBAlgorithm in avi
 */
export const LBAlgorithmShortName = [
  {
    value: LBAlgorithm.LB_ALGORITHM_LEAST_CONNECTIONS,
    shortName: 'Lease Connections'
  },
  {
    value: LBAlgorithm.LB_ALGORITHM_ROUND_ROBIN,
    shortName: 'Round Robin'
  },
  {
    value: LBAlgorithm.LB_ALGORITHM_FASTEST_RESPONSE,
    shortName: 'Fastest Response'
  },
  {
    value: LBAlgorithm.LB_ALGORITHM_CONSISTENT_HASH,
    shortName: 'Consistent Hash'
  },
  {
    value: LBAlgorithm.LB_ALGORITHM_LEAST_LOAD,
    shortName: 'Lease Load'
  },
  {
    value: LBAlgorithm.LB_ALGORITHM_FEWEST_SERVERS,
    shortName: 'Fastest Servers'
  },
  {
    value: LBAlgorithm.LB_ALGORITHM_RANDOM,
    shortName: 'Random'
  },
  {
    value: LBAlgorithm.LB_ALGORITHM_FEWEST_TASKS,
    shortName: 'Fastest Tasks'
  },
  {
    value: LBAlgorithm.LB_ALGORITHM_NEAREST_SERVER,
    shortName: 'Nearest Server'
  },
  {
    value: LBAlgorithm.LB_ALGORITHM_CORE_AFFINITY,
    shortName: 'Core Affinity'
  },
  {
    value: LBAlgorithm.LB_ALGORITHM_TOPOLOGY,
    shortName: 'Topology'
  }
]

export enum AlgorithmHash {
  LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS = 'LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS',
  LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS_AND_PORT = 'LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS_AND_PORT',
  LB_ALGORITHM_CONSISTENT_HASH_URI = 'LB_ALGORITHM_CONSISTENT_HASH_URI',
  LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER = 'LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER',
  LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_STRING = 'LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_STRING',
  LB_ALGORITHM_CONSISTENT_HASH_CALLID = 'LB_ALGORITHM_CONSISTENT_HASH_CALLID'
}

export const AlgorithmHashShortName = [
  {
    value: AlgorithmHash.LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS,
    shortName: 'Source IP Address'
  },
  {
    value: AlgorithmHash.LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS_AND_PORT,
    shortName: 'Source IP Address And Port'
  },
  {
    value: AlgorithmHash.LB_ALGORITHM_CONSISTENT_HASH_URI,
    shortName: 'URI'
  },
  {
    value: AlgorithmHash.LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER,
    shortName: 'Custom Header'
  },
  {
    value: AlgorithmHash.LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_STRING,
    shortName: 'Custom String'
  },
  {
    value: AlgorithmHash.LB_ALGORITHM_CONSISTENT_HASH_CALLID,
    shortName: 'Callid'
  }
]

export class CreateAviRouteSpec {
  networkProfile?: string;
  lbAlgorithm?: LBAlgorithm;
  lbAlgorithmHash?: AlgorithmHash;
  // only used for LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER
  lbAlgorithmConsistentHashHeader?: string;
  vsVip?: string;
  persistentVsVIP?: boolean;
}

export class CreateRouteSpec {
  avi?: CreateAviRouteSpec
}

export class AddRouteBackend {
  name: string;
  type?: string = 'Service';
  ip?: string;
  port?: number;
  ratio: number;
  instance: string;
}

export class CreateRoute {
  name: string;
  description?: string;
  protocol?: RouteProtocol;
  port: number;
  backends?: AddRouteBackend[];
  spec?: CreateRouteSpec;
  component: string;
  loadBalancer: string;
}

export class UpdateAviRouteSpec {
  lbAlgorithm?: LBAlgorithm;
  lbAlgorithmHash?: AlgorithmHash;
  lbAlgorithmConsistentHashHeader?: string;
  persistentVsVIP?: boolean;
}

export class UpdateRouteSpec {
  avi?: UpdateAviRouteSpec;
}

export class RemoveRouteBackend {
  id: string;
}

export class UpdateRoute {
  id: string;
  description?: string;
  port: number;
  addBackends?: AddRouteBackend[];
  removeBackends?: RemoveRouteBackend[];
  spec?: UpdateRouteSpec;
}

export class DeleteRoute {
  id: string;
  force?: boolean;
}
