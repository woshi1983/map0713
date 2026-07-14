import { BackupStorageLocation } from './backup_storage_location';
import { Cluster } from './cluster';
import { Restore } from './restore'

export enum BackupStatus {
  UNKNOWN = 'UNKNOWN',
  PENDING = 'PENDING',
  RUNNING = 'RUNNING',
  SUCCESS = 'SUCCESS',
  FAILURE = 'FAILURE'
}

export class BackupSpec {
  useRestic?: boolean;
  includedNamespaces?: string[];
  excludedNamespaces?: string[];
  includedResources?: string[];
  excludedResources?: string[];
  ttl?: string;
  includeClusterResources?: boolean;
}

export class BackupInstance {
  name: string;
  namespace: string;
}

export class Backup {
  name?: string;
  labels?: JSON;
  spec?: BackupSpec;
  status?: BackupStatus;
  instances: BackupInstance[];
  restores?: Restore;
  location?: BackupStorageLocation;
  cluster?: Cluster;
  createdAt?: string;
  lastModified?: string;
}

export class CreateBackupSpec {
  useRestic?: boolean;
  includedNamespaces?: string[];
  excludedNamespaces?: string[];
  includedResources?: string[];
  excludedResources?: string[];
  components?: string[];
  ttl?: string;
  includeClusterResources?: boolean;
}

export class CreateBackup {
  name?: string;
  spec?: CreateBackupSpec;
  cluster?: string;
  location?: string;
}

export class DeleteBackup {
  name: string;
  cluster: string;
}
