import { Cluster } from './cluster';
import { JobStatus } from './job';

export class SourceBuildSpecSourceGit {
  revision: string;
  url: string;
}

export class SourceBuildSpecSourceBlob {
  url: string;
}

export class SourceBuildSpecSourceRegistry {
  image: string;
}

export class SourceBuildSpecSource {
  git: SourceBuildSpecSourceGit;
  blob: SourceBuildSpecSourceBlob;
  registry: SourceBuildSpecSourceRegistry;
  subPath: string;
}

export class SourceBuildSpec {
  source: SourceBuildSpecSource;
  env: JSON;
  distImage: string;
}

export class SourceBuildStatus {
  num: number;
  status: JobStatus;
  image: string;
}

export class SourceBuild {
  name: string;
  spec: SourceBuildSpec;
  status: SourceBuildStatus;
  cluster: Cluster;
  createdAt: string;
  lastModified: string;
}

export class CreateSourceBuild {
  image: string;
  cluster: string;
}

export class DeleteSourceBuild {
  name: string;
  cluster: string;
}
