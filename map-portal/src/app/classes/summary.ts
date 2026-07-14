export class SummaryBasic {
  orgCount: number;
  activeUsers: number;
  customComponentCount: number;
  instanceCount: number;
  imageBuildsCount: number;
  ingressCount: number;
  routeCount: number;
}

export class ClusterStatusMetrics {
  name?: string;
  organization?: string;
  status?: string;
}

export class Extension {
  cluster: string;
  organization: string;
  state: string;
}

export class SummaryExtensions {
  contour: {
    extensions: Extension[];
    readyCount: number;
  };

  octant: {
    extensions: Extension[];
    readyCount: number;
  };

  grafana: {
    extensions: Extension[];
    readyCount: number;
  };

  prometheus: {
    extensions: Extension[];
    readyCount: number;
  };

  elasticsearch: {
    extensions: Extension[];
    readyCount: number;
  };

  fluentbit: {
    extensions: Extension[];
    readyCount: number;
  }
}
export class OrgMetrics {
  name: string[];
  cpuUsed?: number[];
  cpuFree?: number[];
  memoryUsed?: number[];
  memoryFree?: number[];
  storageUsed?: number[];
  storageFree?: number[];
}

export class OrgResourceMetrics {
  totalResource: {
    cpu: {
      used?: number;
      total: number;
    };
    memory: {
      used?: number;
      total: number;
    };
    storage: {
      used?: number;
      total: number;
    };
  };

  orgMetrics: OrgMetrics;
}

export class Metrics {
  activeUsers: number;
  localUsersDisabled?: boolean;
}
