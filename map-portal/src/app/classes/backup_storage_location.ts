import { Cluster } from './cluster';
import { Backup } from './backup';
import { Restore } from './restore';

export enum BackupStorageLocationAccessMode {
  READ_ONLY = 'READ_ONLY',
  READ_WRITE = 'READ_WRITE'
}

export enum BackupStorageLocationStatus {
  UNKNOWN = 'UNKNOWN',
  AVAILABLE = 'AVAILABLE',
  UNAVAILABLE = 'UNAVAILABLE'
}

export class BackupStorageLocationSpecConfig {
  region: string;
  s3ForcePathStyle: boolean;
  s3Url: string;
}

export class BackupStorageLocationSpecObjectStorage {
  bucket?: string;
  prefix?: string;
  caCert?: string;
}

export class BackupStorageLocationSpec {
  provider?: string;
  config?: BackupStorageLocationSpecConfig;
  objectStorage?: BackupStorageLocationSpecObjectStorage;
  accessMode?: BackupStorageLocationAccessMode;
}

export class BackupStorageLocation {
  name?: string;
  labels?: JSON;
  spec?: BackupStorageLocationSpec;
  status?: BackupStorageLocationStatus;
  backups?: Backup[];
  restores?: Restore[];
  cluster?: Cluster;
  createdAt?: string;
  lastModified?: string;
}

export class CreateBackupStorageLocationSpecConfig {
  region?: string;
  s3ForcePathStyle?: boolean;
  s3Url?: string;
}

export class CreateBackupStorageLocationSpecObjectStorage {
  bucket?: string;
  prefix?: string;
  caCert?: string;
}

export class CreateBackupStorageLocationSpecProfile {
  awsAccessKeyID?: string;
  awsSecretAccessKey?: string;
}

export class CreateBackupStorageLocationSpec {
  provider?: string;
  config?: CreateBackupStorageLocationSpecConfig;
  objectStorage?: CreateBackupStorageLocationSpecObjectStorage;
  accessMode?: BackupStorageLocationAccessMode;
  profile?: CreateBackupStorageLocationSpecProfile;
}

export class CreateBackupStorageLocation {
  name?: string;
  spec?: CreateBackupStorageLocationSpec;
  cluster?: string;
  dryrun?: boolean;
  validate?: boolean;
}

export class DeleteBackupStorageLocation {
  name: string;
  cluster: string;
}
