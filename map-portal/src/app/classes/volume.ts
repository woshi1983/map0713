import { ConfigMap } from './configmap';
import { Secret } from './secret';
import { PersistentVolumeClaim } from './persistent_volume_claim';

export class TemplateSpecVolumeKeyToPath {
  key: string;
  path: string;
  mode: number;
}

// union Volume = HostPath | EmptyDir | ConfigMapVolume | SecretVolume | PersistentVolumeClaimVolume
// add hostpath id not supported current
export class HostPath {
  __typename: string;
  path: string;
  type: string;
}

export class EmptyDir {
  __typename: string;
  medium: string;
  size: string;
}

export class ConfigMapVolume {
  __typename: string;
  name: string;
  ref: ConfigMap;
  items: TemplateSpecVolumeKeyToPath[];
  defaultMode: number;
}

export class SecretVolume {
  __typename: string;
  name: string;
  ref: Secret;
  items: TemplateSpecVolumeKeyToPath[];
  defaultMode: number;
}

export class PersistentVolumeClaimVolume {
  __typename: string;
  name: string;
  ref: PersistentVolumeClaim;
  readOnly: boolean;
}

export class HostPathInput {
  path: string;
  type: string;
}

export class EmptyDirInput {
  medium: string;
  size: string;
}

export class TemplateSpecVolumeKeyToPathInput {
  key?: string;
  path?: string;
  mode?: number;
}

export class ConfigMapVolumeInput {
  name: string;
  items?: TemplateSpecVolumeKeyToPathInput[];
  defaultMode?: number;
}

export class SecretVolumeInput {
  name: string;
  items?: TemplateSpecVolumeKeyToPathInput[];
  defaultMode?: number;
}

export class PersistentVolumeClaimVolumeInput {
  name: string;
  readOnly?: boolean;
}
