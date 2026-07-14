import { Cluster } from './cluster';
import { Job } from './job';

export class BinaryBuildSpec {
  baseImage: string;
  distImage: string;
  binaryDir: string;
  url: string;
  processUID: number;
}

export class BinaryBuildStatus {
  num: number;
  status: string;
  image: string;
  job: Job;
}

export class BinaryBuild {
  name: string;
  spec: BinaryBuildSpec;
  status: BinaryBuildStatus;
  cluster: Cluster;
  createdAt: string;
  lastModified: string;
}

export class CreateBinaryBuild {
  image: string;
  cluster: string;
}

export class DeleteBinaryBuild {
  name: string;
  cluster: string;
}
