import { Cluster } from './cluster';

export class NodeTaint {
  key?: string;
  value?: string;
  effect?: string;
}

export class NodeSpec {
  unschedulable?: boolean;
  taints?: NodeTaint[];
}

export class NodeAllocatedResources {
  limits: JSON;
  requests: JSON;
}

export class NodeStatus {
  _name?: string
  _cluster?: string;
  kernelVersion?: string;
  os?: string;
  osImage?: string;
  usage?: JSON;
  allocated?: NodeAllocatedResources;
  allocatable?: JSON;
  capacity?: JSON;
  ready?: boolean;
  internalIP?: string[];
  externalIP?: string[];
}

export class Node {
  name?: string;
  labels?: JSON;
  spec?: NodeSpec;
  status?: NodeStatus;
  capacity?: {
    cpu: number;
    memory: number;
  };

  cluster?: Cluster;
  createdAt?: string;
  lastModified?: string;
}

export class NodeResource {
  name?: string;
  cpu?: {
    used?: number;
    total?: number;
    request?: number;
    limit?: number;
    allocatable?: number;
  };

  memory?: {
    used?: number;
    total?: number;
    request?: number;
    limit?: number;
    allocatable?: number;
  };
}

export class DrainNodeEvent {
  evictedPods: number;
  totalPods: number;
  stdout: string;
  stderr: string;
  timestamp: string;
}

export class AddNodeTaint {
  // internal taint like `node.kubernetes.io/unschedulable` is not allowed to be removed.
  key: string;
  value?: string;
  effect: string;
}

export class RemoveNodeTaint {
  key: string;
}

export class UpdateNodeSpec {
  unschedulable?: boolean;
  addTaints?: AddNodeTaint[];
  removeTaints?: RemoveNodeTaint[];
}

export class UpdateNode {
  name: string;
  labels?: JSON;
  spec?: UpdateNodeSpec;
  cluster: string;
}

export class DrainNode {
  name: string;
  timeout: string;
  cluster: string;
}
