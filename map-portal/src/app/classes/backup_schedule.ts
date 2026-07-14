import { Cluster } from './cluster';
import { BackupSpec, CreateBackupSpec } from './backup';
import { Organization } from './organization';
import { CustomComponents } from './component';

export class ClusterBackupSchedule {
  id: string;
  name: string;
  spec: BackupSpec;
  location: string;
  schedule: string;
  lastFire: string;
  nextFire: string;
  components: CustomComponents;
  cluster: Cluster;
  organization: Organization;
  createdAt: string;
  lastModified: string;
}

export class ClusterBackupSchedules {
  edge: {
    nodes: ClusterBackupSchedule[];
  }
}

export class CreateClusterBackupSchedule {
  name: string;
  spec?: CreateBackupSpec;
  location: string;
  schedule: string;
  cluster: string;
}

export class UpdateClusterBackupSchedule {
  id: string;
  schedule?: string;
}

export class DeleteClusterBackupSchedule {
  id: string;
}
