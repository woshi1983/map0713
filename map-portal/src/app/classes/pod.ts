import { HostPathInput, EmptyDirInput, ConfigMapVolumeInput, SecretVolumeInput, PersistentVolumeClaimVolumeInput } from './volume';
import { Application } from './application';
import { Cluster } from './cluster';
import { CustomComponent } from './component';
import { Instance } from './instance';

export class podLogsSpec {
  cluster: string;
  namespace: string;
  controller: string;
  issuer: string;
}

export class PodSpecContainerVolumeMount {
  name: string;
  readOnly: boolean;
  mountPath: string;
  subPath: string;
  subPathExpr: string;
  mountPropagation: string;
}

export class PodSpecContainerPort {
  name: string;
  hostPort: number;
  containerPort: number;
  protocol: string;
  hostIP: string;
}

export class PodSpecContainerResourcesLimits {
  cpu: string;
  memory: string;
}

export class PodSpecContainerResourcesRequests {
  cpu: string;
  memory: string;
}

export class PodSpecContainerResources {
  limits: PodSpecContainerResourcesLimits;
  requests: PodSpecContainerResourcesRequests;
}

export class PodSpecContainerExecAction {
  command: string[];
}

export class PodSpecContainerHandlerHTTPGetActionHTTPHeader {
  name: string;
  value: string;
}

export class PodSpecContainerHTTPGetAction {
  path: string;
  port: number | string;
  host: string;
  scheme: string;
  httpHeaders: PodSpecContainerHandlerHTTPGetActionHTTPHeader[];
}

export class PodSpecContainerTCPSocketAction {
  port: number | string;
  host: string;
}

export class PodSpecContainerHandler {
  exec: PodSpecContainerExecAction;
  httpGet: PodSpecContainerHTTPGetAction;
  tcpSocket: PodSpecContainerTCPSocketAction;
}

export class PodSpecContainerProbe {
  exec: PodSpecContainerExecAction;
  httpGet: PodSpecContainerHTTPGetAction;
  tcpSocket: PodSpecContainerTCPSocketAction;
  initialDelaySeconds: number;
  timeoutSeconds: number;
  periodSeconds: number;
  successThreshold: number;
  failureThreshold: number;
  terminationGracePeriodSeconds: number;
}

export class PodSpecContainerLifecycle {
  postStart: PodSpecContainerHandler;
  preStop: PodSpecContainerHandler;
}

export class PodSpecContainer {
  name: string;
  image: string;
  command: string[];
  args: string[];
  workingDir: string;
  ports: PodSpecContainerPort[];
  env: JSON;
  resources: PodSpecContainerResources;
  volumeMounts: PodSpecContainerVolumeMount[];
  livenessProbe: PodSpecContainerProbe;
  readinessProbe: PodSpecContainerProbe;
  startupProbe: PodSpecContainerProbe;
  lifecycle: PodSpecContainerLifecycle;
  imagePullPolicy: string;
}

export class PodSpecVolume {
  name: string;
  volume: any;
}

export class PodSpec {
  initContainers: PodSpecContainer[];
  containers: PodSpecContainer[];
  dnsPolicy: string;
  restartPolicy: string;
  schedulerName: string;
  serviceAccount: string;
  terminationGracePeriodSeconds: number;
  volumes: PodSpecVolume[];
  imagePullSecrets: string[];
}

export class PodStatusContainerStatus {
  containerID: string;
  image: string;
  imageID: string;
  name: string;
  ready: boolean;
  restartCount: number;
  started: boolean;
}

export class PodStatus {
  phase: string;
  message: string;
  reason: string;
  hostIP: string;
  podIPs: string[];
  initContainerStatuses: PodStatusContainerStatus[];
  containerStatuses: PodStatusContainerStatus[];
  qosClass: string;
}

export class Pod {
  name: string;
  namespace: string;
  labels: JSON;
  spec: PodSpec;
  status: PodStatus;
  instance: Instance;
  component: CustomComponent;
  application: Application;
  cluster: Cluster;
  createdAt: string;
  lastModified: string;
}

export class PodTemplate {
  labels: JSON;
  spec: PodSpec;
}

export class AddPodSpecVolume {
  name: string;
  hostPath?: HostPathInput;
  emptyDir: EmptyDirInput;
  configMap: ConfigMapVolumeInput;
  secret: SecretVolumeInput;
  persistentVolumeClaim: PersistentVolumeClaimVolumeInput
}

export class AddPodSpecContainerResourcesLimits {
  cpu?: string;
  memory?: string;
}

export class AddPodSpecContainerResourcesRequests {
  cpu?: string;
  memory?: string;
}

export class AddPodSpecContainerResources {
  limits?: AddPodSpecContainerResourcesLimits;
  requests?: AddPodSpecContainerResourcesRequests;
}

export class AddPodSpecContainerVolumeMount {
  name: string;
  readOnly?: boolean;
  mountPath: string;
  subPath: string;
  subPathExpr?: string;
  mountPropagation?: string;
}

export class InputPodSpecContainerExecAction {
  command?: string[];
}

export class InputPodSpecContainerHandlerHTTPGetActionHTTPHeader {
  name: string;
  value: string;
}

export class InputPodSpecContainerHTTPGetAction {
  path?: string;
  port?: number | string;
  host?: string;
  scheme?: string;
  httpHeaders?: InputPodSpecContainerHandlerHTTPGetActionHTTPHeader[];
}

export class InputPodSpecContainerTCPSocketAction {
  port: number | string;
  host?: string;
}

export class InputPodSpecContainerHandler {
  exec?: InputPodSpecContainerExecAction;
  httpGet?: InputPodSpecContainerHTTPGetAction;
  tcpSocket?: InputPodSpecContainerTCPSocketAction;
}

export class InputPodSpecContainerLifecycle {
  postStart?: InputPodSpecContainerHandler;
  preStop?: InputPodSpecContainerHandler;
}

export class InputPodSpecContainerProbe {
  exec?: InputPodSpecContainerExecAction;
  httpGet?: InputPodSpecContainerHTTPGetAction;
  tcpSocket?: InputPodSpecContainerTCPSocketAction;
  initialDelaySeconds?: number;
  timeoutSeconds?: number;
  periodSeconds?: number;
  successThreshold?: number;
  failureThreshold?: number;
  terminationGracePeriodSeconds?: number;
}

export class AddPodSpecContainerPort {
  name: string;
  protocol: string;
  containerPort: number;
}

export class AddPodSpecContainer {
  name: string;
  image: string;
  command?: string[];
  args?: string[];
  workingDir?: string;
  ports?: AddPodSpecContainerPort[];
  env?: JSON;
  resources?: AddPodSpecContainerResources;
  volumeMounts?: AddPodSpecContainerVolumeMount[];
  livenessProbe?: InputPodSpecContainerProbe;
  readinessProbe?: InputPodSpecContainerProbe;
  startupProbe?: InputPodSpecContainerProbe;
  lifecycle?: InputPodSpecContainerLifecycle;
  imagePullPolicy?: string;
}

export class CreatePodSpec {
  initContainers?: AddPodSpecContainer[];
  containers?: AddPodSpecContainer[];
  dnsPolicy?: string;
  restartPolicy?: string;
  schedulerName?: string;
  serviceAccount?: string;
  terminationGracePeriodSeconds?: number;
  volumes?: AddPodSpecVolume[];
  imagePullSecrets?: string[];
}

export class CreatePodTemplate {
  spec: CreatePodSpec;
}

export class RemovePodSpecVolume {
  name: string;
}

export class RemovePodSpecContainerVolumeMount {
  name: string;
}

export class ModifyPodSpecContainerResourcesLimits {
  cpu: string;
  memory: string;
}

export class ModifyPodSpecContainerResourcesRequests {
  cpu: string;
  memory: string;
}

export class ModifyPodSpecContainerResources {
  limits: ModifyPodSpecContainerResourcesLimits;
  requests: ModifyPodSpecContainerResourcesRequests;
}

export class RemovePodSpecContainerPort {
  name: string;
}

export class RemovePodSpecContainer {
  name: string;
}

export class ModifyPodSpecContainer {
  name: string;
  image: string;
  command: string[];
  args: string[];
  workingDir: string;
  addPorts: AddPodSpecContainerPort[];
  removePorts: RemovePodSpecContainerPort[];
  env: JSON;
  resources: ModifyPodSpecContainerResources;
  addVolumeMounts: AddPodSpecContainerVolumeMount[];
  removeVolumeMounts: RemovePodSpecContainerVolumeMount[];
  livenessProbe: InputPodSpecContainerProbe;
  readinessProbe: InputPodSpecContainerProbe;
  startupProbe: InputPodSpecContainerProbe;
  lifecycle: InputPodSpecContainerLifecycle;
  imagePullPolicy: string;
}

export class UpdatePodSpec {
  addInitContainers: AddPodSpecContainer[];
  modifyInitContainers: ModifyPodSpecContainer[];
  removeInitContainers: RemovePodSpecContainer[];
  addContainers: AddPodSpecContainer[];
  modifyContainers: ModifyPodSpecContainer[];
  removeContainers: RemovePodSpecContainer[];
  dnsPolicy: string;
  restartPolicy: string;
  schedulerName: string;
  serviceAccount: string;
  terminationGracePeriodSeconds: number;
  addVolumes: AddPodSpecVolume[];
  removeVolumes: RemovePodSpecVolume[];
  imagePullSecrets: string[];
}

export class UpdatePodTemplate {
  spec: UpdatePodSpec;
}

export class WritePodExec {
  id: string;
  content: string;
}

export class PodLog {
  message: string;
  timestamp: string;
}

export class PodExec {
  id: string;
  message: string;
  timestamp: string;
}
