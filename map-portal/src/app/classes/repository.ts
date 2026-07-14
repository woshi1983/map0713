import { Organization } from './organization';
import { X509Certificate } from './init';

export enum ImageRegistryType {
  UNKNOWN = 'UNKNOWN',
  HARBOR = 'HARBOR'
}

export class ImageRegistry {
  id: string;
  name: string;
  description?: string;
  endpoint?: string;
  caCert?: string;
  tlsCertChain?: X509Certificate[];
  authUser?: string;
  type?: ImageRegistryType;
  projects?: string[];
  organization?: Organization;
  createdAt?: string;
  lastModified?: string;
}

export class ImageRegistries {
  totalCount: number;
  edge: {
    nodes: ImageRegistry[];
  }
}

export class ImageRepo {
  name: string;
  project: string;
  tags: string[];
  imageRegistry: ImageRegistry;
}

export class ChartDetailsDependency {
  name: string;
  repository: string;
  version: string;
}

export class ChartDetails {
  name: string;
  description: string;
  version: string;
  avatar: string;
  sources: string[];
  homepage: string;
  download: string;
  keywords: string[];
  dependencies: ChartDetailsDependency[];
  files: JSON;
  values: any;
}

export class ChartVersion {
  _repo: string;
  _project: string;
  _imageRegistry: string;
  version: string;
  details: ChartDetails;
}

export class ChartRepo {
  name: string;
  latestVersion: string;
  totalVersions: number;
  avatar: string;
  deprecated: boolean;
  homepage: string;
  versions: ChartVersion[];
  project: string;
  imageRegistry: ImageRegistry;
}

export class CreateImageRegistry {
  name: string;
  description?: string;
  endpoint: string;
  authUser?: string;
  authPassword?: string;
  caCert?: string;
  type?: string = 'HARBOR';
  organization: string;
  dryrun?: boolean;
  validate?: boolean;
}

export class UpdateImageRegistry {
  id: string;
  name?: string;
  description?: string;
  endpoint?: string;
  authUser?: string;
  authPassword?: string;
  caCert?: string;
  dryrun?: boolean;
  validate?: boolean;
}

export class DeleteImageRegistry {
  id: string;
}
