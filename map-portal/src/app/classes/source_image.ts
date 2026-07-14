import { Cluster } from './cluster';

export class SourceImageSpecSourceGit {
  revision: string;
  url: string;
}

export class SourceImageSpecSourceBlob {
  url: string;
}

export class SourceImageSpecSourceRegistry {
  image: string;
}

export class SourceImageSpecSource {
  git: SourceImageSpecSourceGit;
  blob: SourceImageSpecSourceBlob;
  registry: SourceImageSpecSourceRegistry;
  subPath: string;
}

export class SourceImageSpec {
  source: SourceImageSpecSource;
  env: JSON;
  distImage: string;
}

export class SourceImageStatus {
  ready: boolean;
}

export class SourceImage {
  name: string;
  annotations: JSON;
  spec: SourceImageSpec;
  status: SourceImageStatus;
  cluster: Cluster;
  createdAt: string;
  lastModified: string;
}

export class CreateSourceImageSpecSourceGit {
  revision?: string;
  url?: string;
}

export class CreateSourceImageSpecSourceBlob {
  url?: string;
}

export class CreateSourceImageSpecSourceRegistry {
  image?: string;
}

export class CreateSourceImageSpecSource {
  git?: CreateSourceImageSpecSourceGit;
  blob?: CreateSourceImageSpecSourceBlob;
  registry?: CreateSourceImageSpecSourceRegistry;
  subPath?: string;
}

export class CreateSourceImageSpec {
  source: CreateSourceImageSpecSource;
  env: JSON;
  distImage: string;
}

export class CreateSourceImage {
  name: string;
  annotations: JSON;
  spec: CreateSourceImageSpec;
  cluster: string;
}

export class UpdateSourceImageSpecSourceGit {
  revision?: string;
  url?: string;
}

export class UpdateSourceImageSpecSourceBlob {
  url?: string;
}

export class UpdateSourceImageSpecSourceRegistry {
  image?: string;
}

export class UpdateSourceImageSpecSource {
  git?: UpdateSourceImageSpecSourceGit;
  blob?: UpdateSourceImageSpecSourceBlob;
  registry?: UpdateSourceImageSpecSourceRegistry;
  subPath?: string;
}

export class UpdateSourceImageSpec {
  source: UpdateSourceImageSpecSource;
  env: JSON;
  distImage: string;
}

export class UpdateSourceImage {
  name: string;
  annotations: JSON;
  spec: UpdateSourceImageSpec;
  cluster: string;
}

export class DeleteSourceImage {
  name: string;
  cluster: string;
}
