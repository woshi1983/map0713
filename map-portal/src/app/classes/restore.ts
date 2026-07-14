import { BackupStorageLocation } from './backup_storage_location';
import { Cluster } from './cluster';
import { Backup } from './backup';

export enum RestoreStatus {
  UNKNOWN = 'UNKNOWN',
  PENDING = 'PENDING',
  RUNNING = 'RUNNING',
  SUCCESS = 'SUCCESS',
  FAILURE = 'FAILURE'
}

export class RestoreSpec {
  includedNamespaces: string[];
  excludedNamespaces: string[];
  includedLabels?: JSON;
  namespaceMapping?: JSON;
}

export class Restore {
  name?: string;
  labels?: JSON;
  spec?: RestoreSpec;
  status?: RestoreStatus;
  backup?: Backup;
  location?: BackupStorageLocation;
  cluster?: Cluster;
  createdAt?: string;
  lastModified?: string;
}

export class CreateRestoreSpec {
  includedNamespaces: string[];
  excludedNamespaces: string[];
  includedLabels?: JSON;
  namespaceMapping?: JSON;
}

export class CreateRestore {
  name: string;
  spec?: CreateRestoreSpec;
  backup: string;
  cluster: string;
}
