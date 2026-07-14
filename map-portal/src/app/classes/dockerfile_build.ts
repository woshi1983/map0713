import { Cluster } from './cluster';
import { Job } from './job';

export class DockerfileBuildSpec {
  source: string;
  distImage: string;
  branch: string;
  subDir: string;
  fileName: string;
  url: string;
}

export class DockerfileBuildStatus {
  num: number;
  status: string;
  image: string;
  job: Job;
}

export class DockerfileBuild {
  name: string;
  spec: DockerfileBuildSpec;
  status: DockerfileBuildStatus;
  cluster: Cluster;
  createdAt: string;
  lastModified: string;
}

export class CreateDockerfileBuild {
  image: string;
  cluster: string;
}

export class DeleteDockerfileBuild {
  name: string;
  cluster: string;
}
