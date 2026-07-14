import { Cluster } from './cluster';

export class BinaryImageSpec {
  baseImage: string;
  distImage: string;
  binaryDir: string;
  url: string;
  processUID: number;
}

export class BinaryImage {
  name: string;
  labels: JSON;
  annotations: JSON;
  spec: BinaryImageSpec;
  cluster: Cluster;
  createdAt: string;
  lastModified: string;
}

export class CreateBinaryImageSpec {
  baseImage: string;
  distImage: string;
  binaryDir: string;
  url: string;
  processUID?: number;
}

export class CreateBinaryImage {
  name: string;
  annotations: JSON;
  spec: CreateBinaryImageSpec;
  cluster: string;
}

export class UpdateBinaryImageSpec {
  baseImage: string;
  distImage: string;
  binaryDir: string;
  url: string;
  processUID?: number;
}

export class UpdateBinaryImage {
  name: string;
  annotations: JSON;
  spec: UpdateBinaryImageSpec;
  cluster: string;
}

export class DeleteBinaryImage {
  name: string;
  cluster: string;
}
