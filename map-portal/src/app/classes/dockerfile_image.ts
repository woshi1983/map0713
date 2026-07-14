import { Cluster } from './cluster';

export class DockerfileImageSpec {
  source: string;
  distImage: string;
  branch: string;
  subDir: string;
  fileName: string;
  url: string;
}

export class DockerfileImage {
  name: string;
  labels: JSON;
  annotations: JSON;
  spec: DockerfileImageSpec;
  cluster: Cluster;
  createdAt: string;
  lastModified: string;
}

export class CreateDockerfileImageSpec {
  source?: string;
  distImage: string;
  branch?: string;
  subDir?: string;
  fileName?: string;
  url: string;
}

export class CreateDockerfileImage {
  name: string;
  annotations: JSON;
  spec: CreateDockerfileImageSpec;
  cluster: string;
}

export class UpdateDockerfileImageSpec {
  source: string;
  distImage: string;
  branch: string;
  subDir: string;
  fileName: string;
  url: string;
}

export class UpdateDockerfileImage {
  name: string;
  annotations: JSON;
  spec: UpdateDockerfileImageSpec;
  cluster: string;
}

export class DeleteDockerfileImage {
  name: string;
  cluster: string;
}
