export class StandardPackageRepository {
  id: string;
  name: string;
  description?: string;
  image: string;
  default: boolean;
  createdAt?: string;
  lastModified?: string;
}

export class StandardPackageRepositories {
  totalCount: number;
  edge: {
    nodes: StandardPackageRepository[];
  }
}

export class CreateStandardPackageRepository {
  name: string;
  description?: string;
  image: string;
  default?: boolean;
  dryrun?: boolean;
}

export class UpdateStandardPackageRepository {
  id: string;
  name?: string;
  description?: string;
  image?: string;
  default?: boolean;
  dryrun?: boolean;
}

export class SetDefaultStandardPackageRepository {
  id: string;
}

export class DeleteStandardPackageRepository {
  id: string;
}
