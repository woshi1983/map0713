import { DeploymentSpec, CreateDeploymentSpec } from './deployment';
import { ServiceSpec, CreateServiceSpec } from './service';
import { Application } from './application';
import { Organization } from './organization';
import { ChartDetails } from './repository';

import { PersistentVolumeClaimSpec, CreatePersistentVolumeClaimSpec } from './persistent_volume_claim';

// union ComponentTemplate = BasicComponentTemplate | HelmComponentTemplate | RAWComponentTemplate

export class Storage {
  name: string;
  storageType: string;
  secretType?: string;
  configMap?: {
    data: JSON;
  };

  secret?: {
    data?: JSON;
  };

  persistentVolumeClaim?: {
    spec: CreatePersistentVolumeClaimSpec;
  };
}

export enum ComponentType {
  BASIC = 'BASIC',
  HELM = 'HELM',
  RAW = 'RAW'
}

export class CustomComponent {
  id: string;
  name: string;
  description: string;
  type: ComponentType;
  avatar: string;
  template: any;
  application: Application;
  organization: Organization;
  createdAt: string;
  lastModified: string;
}

export class CustomComponents {
  totalCount: number;
  edge: {
    nodes: CustomComponent[];
  }
}

export class HorizontalPodAutoscalerSpec {
  maxReplicas: number;
  minReplicas: number;
  targetCPUUtilizationPercentage: number;
  targetMemoryUtilizationPercentage: number;
}

export class BasicComponentTemplateDeployment {
  name: string;
  labels: JSON;
  spec: DeploymentSpec;
  hpa: HorizontalPodAutoscalerSpec;
}

export class BasicComponentTemplateService {
  name: string;
  labels: JSON;
  spec: ServiceSpec;
}
export class BasicComponentTemplateGenericSecret {
  name: string;
  labels: JSON;
  data: JSON;
  type: string;
}

export class BasicComponentTemplateConfigMap {
  name: string;
  labels: JSON;
  data: JSON;
}

export class BasicComponentTemplatePersistentVolumeClaim {
  name: string;
  labels: JSON;
  spec: PersistentVolumeClaimSpec;
}

export class BasicComponentTemplate {
  deployments: BasicComponentTemplateDeployment[];
  services: BasicComponentTemplateService[];
  secrets: BasicComponentTemplateGenericSecret;
  configMaps: BasicComponentTemplateConfigMap[];
  persistentVolumeClaims: BasicComponentTemplatePersistentVolumeClaim[];
}

export class HelmComponentTemplate {
  chart: ChartDetails
  url: string;
  values: any;
  extraInfo: JSON;
}

export class RAWComponentTemplate {
  raw: any[];
}

export class CreateHorizontalPodAutoscalerSpec {
  maxReplicas: number;
  minReplicas: number;
  targetCPUUtilizationPercentage?: number;
  targetMemoryUtilizationPercentage?: number;
}

export class AddBasicComponentTemplateDeployment {
  name: string
  labels?: JSON;
  spec: CreateDeploymentSpec;
  hpa?: CreateHorizontalPodAutoscalerSpec;
}

export class AddBasicComponentTemplateService {
  name: string;
  labels?: JSON
  spec: CreateServiceSpec;
}

export class AddBasicComponentTemplateGenericSecret {
  data: JSON;
  type: string;
}

export class AddBasicComponentTemplateSecret {
  name: string;
  generic?: AddBasicComponentTemplateGenericSecret;
}

export class AddBasicComponentTemplateConfigMap {
  name: string;
  data: JSON;
}

export class AddBasicComponentTemplatePersistentVolumeClaim {
  name: string;
  spec: CreatePersistentVolumeClaimSpec;
}

export class CreateBasicComponentTemplate {
  deployments?: AddBasicComponentTemplateDeployment[];
  services?: AddBasicComponentTemplateService[];
  secrets: AddBasicComponentTemplateSecret[];
  configMaps: AddBasicComponentTemplateConfigMap[];
  persistentVolumeClaims: AddBasicComponentTemplatePersistentVolumeClaim[];
}

export class CreateHelmComponentTemplate {
  url: string;
  values: any;
  extraInfo?: JSON;
}

export class CreateComponentTemplate {
  basic?: CreateBasicComponentTemplate;
  helm?: CreateHelmComponentTemplate;
  raw?: string;
}

export class CreateComponent {
  name: string;
  description?: string;
  avatar?: string;
  template: CreateComponentTemplate;
  application: string;
}

export class RemoveBasicComponentTemplateDeployment {
  name: string;
}

export class RemoveBasicComponentTemplateService {
  name: string;
}

export class RemoveBasicComponentTemplateSecret {
  name: string;
}

export class RemoveBasicComponentTemplateConfigMap {
  name: string;
}

export class RemoveBasicComponentTemplatePersistentVolumeClaim {
  name: string;
}

export class UpdateBasicComponentTemplate {
  addDeployments?: AddBasicComponentTemplateDeployment[];
  removeDeployments?: RemoveBasicComponentTemplateDeployment[];
  addServices?: AddBasicComponentTemplateService[];
  removeServices?: RemoveBasicComponentTemplateService[];
  addSecrets: AddBasicComponentTemplateSecret[];
  removeSecrets: RemoveBasicComponentTemplateSecret[];
  addConfigMaps: AddBasicComponentTemplateConfigMap[];
  removeConfigMaps: RemoveBasicComponentTemplateConfigMap[];
  addPersistentVolumeClaims: AddBasicComponentTemplatePersistentVolumeClaim[];
  removePersistentVolumeClaims: RemoveBasicComponentTemplatePersistentVolumeClaim[];
}

export class UpdateHelmComponentTemplate {
  url: string;
  values: any;
  extraInfo?: JSON;
}

export class UpdateComponentTemplate {
  basic?: UpdateBasicComponentTemplate;
  helm?: UpdateHelmComponentTemplate;
  raw?: string;
}

export class UpdateComponent {
  id: string;
  description?: string;
  avatar?: string;
  template?: UpdateComponentTemplate;
}

export class DeleteComponent {
  id: string;
}
