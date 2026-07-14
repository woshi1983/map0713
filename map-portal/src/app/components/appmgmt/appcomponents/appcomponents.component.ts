import { Component, ElementRef, OnInit, ViewChild } from '@angular/core';
import { AuthService } from '../../../services/auth.service';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from '@ngx-translate/core';
import {
  AddBasicComponentTemplateConfigMap,
  AddBasicComponentTemplateDeployment,
  AddBasicComponentTemplatePersistentVolumeClaim,
  AddBasicComponentTemplateSecret,
  AddBasicComponentTemplateService,
  AddPodSpecContainer,
  AddPodSpecContainerPort,
  AddPodSpecContainerResourcesLimits,
  AddPodSpecContainerResourcesRequests,
  AddPodSpecContainerVolumeMount,
  AddPodSpecVolume,
  AddServiceSpecPort,
  Application,
  BasicComponentTemplateConfigMap,
  BasicComponentTemplatePersistentVolumeClaim,
  ChartVersion,
  CreateApplication,
  CreateComponent,
  CreateHorizontalPodAutoscalerSpec,
  CustomComponent,
  handleObject,
  InputPodSpecContainerHandler,
  InputPodSpecContainerHandlerHTTPGetActionHTTPHeader,
  InputPodSpecContainerLifecycle,
  InputPodSpecContainerProbe,
  RemoveBasicComponentTemplateConfigMap,
  RemoveBasicComponentTemplateDeployment,
  RemoveBasicComponentTemplatePersistentVolumeClaim,
  RemoveBasicComponentTemplateSecret,
  RemoveBasicComponentTemplateService,
  Secret,
  Storage,
  TemplateSpecVolumeKeyToPathInput,
  UpdateComponent
} from '../../../classes';
import { ClrDatagridPagination, ClrDatagridStateInterface, ClrWizard } from '@clr/angular';
import { ComponentService } from '../../../services/component.service';
import {
  CAFormatValidator,
  checkOneYamlFormat,
  checkYamlFormat,
  CustomValidatorDirective,
  KEYFormatValidator,
  nameValidator,
  portOverlapValidator
} from '../../../utils/validators/custom-validator.directive';
import { FormArray, FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';
import YAML from 'yaml';
import { RepositorysService } from '../../../services/repositorys.service';
import { OrganizationService } from '../../../services/organization.service';
import { SELECTED_APPLICATION } from '../../../utils/static-entries/shared.const';
import { ApplicationsService } from '../../../services/applications.service';
import { deepCopy } from '../../../utils/operator';

@Component({
  selector: 'app-components',
  templateUrl: './appcomponents.component.html',
  styleUrls: ['./appcomponents.component.css']
})
export class AppcomponentsComponent implements OnInit {
  // admin and user role can handle component
  get canWriteComponent (): boolean {
    return this.authService.can(this.orgId, handleObject.COMPONENT);
  }

  // admin and user role can handle application
  get canWriteApplication (): boolean {
    return this.authService.can(this.orgId, handleObject.APPLICATION);
  }

  constructor (
    public authService: AuthService,
    private toast: ToastrService,
    private translate: TranslateService,
    private componentService: ComponentService,
    private repositoryService: RepositorysService,
    private orgService: OrganizationService,
    private applicationService: ApplicationsService,
    private fb: FormBuilder
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  orgId: string;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  selectedApplication: string;
  allApplications: Application[] = [];
  // used to reset DOM
  deleteApplicationList: Application[] = [];
  applicationLoading = false;
  showCreateAppModel = false;
  showDeleteAppModel = false;
  basicAppForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(32), Validators.pattern(this.restricNamePattern), nameValidator([])])
  });

  deleteAppForm = new FormGroup({
    id: new FormControl('', [Validators.required])
  });

  allComponents: CustomComponent[] = [];
  selectedComponent: CustomComponent;

  loading = false;
  showSearch = true;
  searchName: string;
  total = 0;

  // add storage for component
  allStorages: Storage[] = [];
  configMapStorageExist = false;
  secretStorageExist = false;
  pvcStorageExist = false;
  allConfigMapStorages: Storage[];
  allSecretStorages: Storage[];
  allPVCStorages: Storage[];
  showAddConfgiMap = false;
  showAddSecret = false;
  showAddPVC = false;
  showAddCMStorageFile = false;
  showAddSecretStorageFile = false;
  showAddEditConfigmapStorage = false;
  showAddEditPVCStorage = false;
  showAddEditSecretStorage = false;
  showAddEditCMStorageFile = false;
  showAddEditSecretStorageFile = false;

  @ViewChild('fileCMUpload', { static: false }) fileCMUpload: ElementRef;
  cmStorageFilename = '';
  cmStorageDataList = new FormArray([]);
  cmStorageDataForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    cmData: this.cmStorageDataList
  });

  @ViewChild('fileSecretUpload', { static: false }) fileSecretUpload: ElementRef;
  @ViewChild('wizardsecret') wizardsecret: ClrWizard;
  basicSecretForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern), nameValidator(['vmware-system-map-registry'])]),
    secretType: new FormControl('Opaque', Validators.required)
  });

  secrteStorageFilename = '';
  genericSecretDataList = new FormArray([]);
  genericSecretDataForm = new FormGroup({
    data: this.genericSecretDataList
  });

  readonly packagePattern = CustomValidatorDirective.packagePattern;
  dockerConfigJSONSecretForm = new FormGroup({
    endpoint: new FormControl('', [Validators.required, Validators.pattern(this.packagePattern)]),
    authUser: new FormControl('', Validators.required),
    authPassword: new FormControl('', Validators.required),
    email: new FormControl('', Validators.email)
  });

  dockerConfigSecretForm = new FormGroup({
    config: new FormControl('', Validators.required)
  });

  basicAuthSecretForm = new FormGroup({
    authUser: new FormControl('', Validators.required),
    authPassword: new FormControl('', Validators.required)
  });

  sshAuthSecretForm = new FormGroup({
    privateKey: new FormControl('', [Validators.required, KEYFormatValidator()])
  });

  TLSSecretForm = new FormGroup({
    key: new FormControl('', [Validators.required, KEYFormatValidator()]),
    cert: new FormControl('', [Validators.required, CAFormatValidator()])
  });

  basicPVCForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    storageClass: new FormControl(''),
    accessMode: new FormControl('', Validators.required),
    request: new FormControl('', Validators.required),
    limit: new FormControl(''),
    volumeMode: new FormControl('Filesystem')
  });

  addStorageEditIndex: number;
  @ViewChild('fileEditCMUpload', { static: false }) fileEditCMUpload: ElementRef;
  cmEditStorageFilename = '';
  cmEditStorageDataList = new FormArray([]);
  cmEditStorageDataForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    cmData: this.cmEditStorageDataList
  });

  @ViewChild('fileEditSecretUpload', { static: false }) fileEditSecretUpload: ElementRef;
  @ViewChild('wizardeditsecret') wizardeditsecret: ClrWizard;
  editSecretForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern), nameValidator(['vmware-system-map-registry'])]),
    secretType: new FormControl('', Validators.required)
  });

  secrteEditStorageFilename = '';
  editGenericSecretDataList = new FormArray([]);
  editGenericSecretDataForm = new FormGroup({
    data: this.editGenericSecretDataList
  });

  editDockerConfigJSONSecretForm = new FormGroup({
    endpoint: new FormControl('', [Validators.required, Validators.pattern(this.packagePattern)]),
    authUser: new FormControl('', Validators.required),
    authPassword: new FormControl('', Validators.required),
    email: new FormControl('', Validators.email)
  });

  editDockerConfigSecretForm = new FormGroup({
    config: new FormControl('', Validators.required)
  });

  editBasicAuthSecretForm = new FormGroup({
    authUser: new FormControl('', Validators.required),
    authPassword: new FormControl('', Validators.required)
  });

  editSshAuthSecretForm = new FormGroup({
    privateKey: new FormControl('', [Validators.required, KEYFormatValidator()])
  });

  editTLSSecretForm = new FormGroup({
    key: new FormControl('', [Validators.required, KEYFormatValidator()]),
    cert: new FormControl('', [Validators.required, CAFormatValidator()])
  });

  editPVCForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    storageClass: new FormControl(''),
    accessMode: new FormControl('', Validators.required),
    request: new FormControl('', Validators.required),
    limit: new FormControl(''),
    volumeMode: new FormControl('')
  });

  // add/edit volume in workload
  showAddEmptyDirVolume = false;
  showAddComnfigMapVolume = false;
  showAddSecretVolume = false;
  showAddPVCVolume = false;
  cmKeysList: string[] = [];
  secretKeysList: string[] = [];
  allVolumes: AddPodSpecVolume[] = [];

  AddEmptyVolumeForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern)]),
    medium: new FormControl('default'),
    size: new FormControl(1)
  });

  cmItemsList = new FormArray([]);
  AddConfigMapVolumeForm = new FormGroup({
    cmName: new FormControl('', Validators.required),
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern)]),
    cmItems: this.cmItemsList,
    defaultMode: new FormControl(420, [Validators.min(0), Validators.max(511)]),
    enableItems: new FormControl(false)
  });

  secretItemsList = new FormArray([]);
  AddSecretVolumeForm = new FormGroup({
    secretName: new FormControl('', Validators.required),
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern)]),
    secretItems: this.secretItemsList,
    defaultMode: new FormControl(420, [Validators.min(0), Validators.max(511)]),
    enableItems: new FormControl(false)
  });

  AddPVCVolumeForm = new FormGroup({
    pvcName: new FormControl('', Validators.required),
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern)]),
    readonly: new FormControl(false)
  });

  showEmptyVolumeDetails = false;
  showConfigmapVolumeDetails = false;
  showSecretVolumeDetails = false;
  showPVCVolumeDetails = false;
  addVolumeEditIndex: number;
  editEmptyVolumeForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern)]),
    medium: new FormControl('default'),
    size: new FormControl(1)
  });

  editAddCMItemsList = new FormArray([]);
  editConfigMapVolumeForm = new FormGroup({
    cmName: new FormControl('', Validators.required),
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern)]),
    cmItems: this.editAddCMItemsList,
    defaultMode: new FormControl(420, [Validators.min(0), Validators.max(511)]),
    enableItems: new FormControl(false)
  });

  editAddSecretItemsList = new FormArray([]);
  editSecretVolumeForm = new FormGroup({
    secretName: new FormControl('', Validators.required),
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern)]),
    secretItems: this.editAddSecretItemsList,
    defaultMode: new FormControl(420, [Validators.min(0), Validators.max(511)]),
    enableItems: new FormControl(false)
  });

  editPVCVolumeForm = new FormGroup({
    pvcName: new FormControl('', Validators.required),
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern)]),
    readonly: new FormControl(false)
  });

  // add/edit container in workload
  showAddContaienr = false;
  portOverlap = false;
  allInitContainers: AddPodSpecContainer[] = [];
  addInitContainer = false;
  allContainers: AddPodSpecContainer[] = [];
  containerCommandList = new FormArray([]);
  containerArgsList = new FormArray([]);
  envs = new FormArray([]);
  ports = new FormArray([]);
  mounts = new FormArray([]);
  AddContainerForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(12), Validators.pattern(this.restricNamePattern)]),
    image: new FormControl('', Validators.required),
    ports: this.ports,
    advancedParams: new FormControl(false),
    enableResource: new FormControl(false),
    enableEnv: new FormControl(false),
    enableHealthCheck: new FormControl(false),
    enableLifecycleHook: new FormControl(false),
    imagePullPolicy: new FormControl('IfNotPresent'),
    commands: this.containerCommandList,
    args: this.containerArgsList,
    workingDir: new FormControl(''),
    cpuRequest: new FormControl(''),
    cpuLimit: new FormControl(''),
    memRequest: new FormControl(''),
    memLimit: new FormControl(''),
    envs: this.envs,
    mounts: this.mounts
  });

  // define health check for container
  showAddHealthCheck = false;
  addHealthCheckType = '';
  allHealthChecks: {'checkType': string, 'params': InputPodSpecContainerProbe}[] = [];
  headersList = new FormArray([]);
  commandList = new FormArray([]);
  AddHealthCheckForm = new FormGroup({
    handlerType: new FormControl('HTTPGET', Validators.required),
    commands: this.commandList,
    https: new FormControl(false),
    headers: this.headersList,
    path: new FormControl('/'),
    port: new FormControl(8080, Validators.required),
    tcpPort: new FormControl(8080, Validators.required),
    failureThreshold: new FormControl(3),
    successThreshold: new FormControl(1),
    initDelay: new FormControl(0),
    period: new FormControl(10),
    timeout: new FormControl(1)
  })

  showAddHealthCheckDetails = false;
  addHealthCheckEditIndex: number;
  editHeadersList = new FormArray([]);
  editCommandList = new FormArray([]);
  editAddHealthCheckForm = new FormGroup({
    handlerType: new FormControl('HTTPGET', Validators.required),
    commands: this.editCommandList,
    https: new FormControl(false),
    headers: this.editHeadersList,
    path: new FormControl('/'),
    port: new FormControl(8080, Validators.required),
    tcpPort: new FormControl(8080, Validators.required),
    failureThreshold: new FormControl(3),
    successThreshold: new FormControl(1),
    initDelay: new FormControl(0),
    period: new FormControl(10),
    timeout: new FormControl(1)
  })

  // define lifecycle hook for container
  showAddLifecycleHook = false;
  addLifecycleHook = '';
  allLifecycleHook: {'hookType': string, 'params': InputPodSpecContainerHandler}[] = [];
  hookHeadersList = new FormArray([]);
  hookCommandList = new FormArray([]);
  AddLiftcycleHookForm = new FormGroup({
    handlerType: new FormControl('HTTPGET', Validators.required),
    commands: this.hookCommandList,
    https: new FormControl(false),
    headers: this.hookHeadersList,
    path: new FormControl('/'),
    port: new FormControl(8080, Validators.required),
    tcpPort: new FormControl(8080, Validators.required)
  })

  showAddLifecycleHookDetails = false;
  addLifecycleHookEditIndex: number;
  editHookHeadersList = new FormArray([]);
  editHookCommandList = new FormArray([]);
  editAddLiftcycleHookForm = new FormGroup({
    handlerType: new FormControl('HTTPGET', Validators.required),
    commands: this.editHookCommandList,
    https: new FormControl(false),
    headers: this.editHookHeadersList,
    path: new FormControl('/'),
    port: new FormControl(8080, Validators.required),
    tcpPort: new FormControl(8080, Validators.required)
  })

  // edit containers
  showAddContaienrDetails = false;
  addContainerEditIndex: number;
  editContainerCommandList = new FormArray([]);
  editContainerArgsList = new FormArray([]);
  editAddEnvs = new FormArray([]);
  editAddPorts = new FormArray([]);
  editAddMounts = new FormArray([]);
  editAddContainerForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(12), Validators.pattern(this.restricNamePattern)]),
    image: new FormControl('', Validators.required),
    ports: this.editAddPorts,
    advancedParams: new FormControl(false),
    enableResource: new FormControl(false),
    enableEnv: new FormControl(false),
    enableHealthCheck: new FormControl(false),
    enableLifecycleHook: new FormControl(false),
    imagePullPolicy: new FormControl('IfNotPresent'),
    commands: this.editContainerCommandList,
    args: this.editContainerArgsList,
    workingDir: new FormControl(''),
    cpuRequest: new FormControl(''),
    cpuLimit: new FormControl(''),
    memRequest: new FormControl(''),
    memLimit: new FormControl(''),
    envs: this.editAddEnvs,
    mounts: this.editAddMounts
  });

  // add/edit workload
  dockerConfigSecretList: Secret[];
  workloadParamsForm = new FormGroup({
    workloadType: new FormControl('Deployment', Validators.required),
    replicas: new FormControl(1, Validators.required),
    advancedParams: new FormControl(false),
    imagePullSecrets: new FormControl(['vmware-system-map-registry']),
    serviceAccount: new FormControl('default'),
    restartPolicy: new FormControl('Always'),
    dnsPolicy: new FormControl('ClusterFirst'),
    schedulerName: new FormControl('default-scheduler'),
    terminationGracePeriodSeconds: new FormControl(30),
    progressDeadlineSeconds: new FormControl(600),
    revisionHistoryLimit: new FormControl(10),
    strategyType: new FormControl('RollingUpdate'),
    maxUnavailable: new FormControl('25%'),
    maxSurge: new FormControl('25%')
  })

  editWorkloadParamsForm = new FormGroup({
    workloadType: new FormControl('Deployment', Validators.required),
    replicas: new FormControl(1, Validators.required),
    advancedParams: new FormControl(false),
    imagePullSecrets: new FormControl(['vmware-system-map-registry']),
    serviceAccount: new FormControl('default'),
    restartPolicy: new FormControl('Always'),
    dnsPolicy: new FormControl('ClusterFirst'),
    schedulerName: new FormControl('default-scheduler'),
    terminationGracePeriodSeconds: new FormControl(30),
    progressDeadlineSeconds: new FormControl(600),
    revisionHistoryLimit: new FormControl(10),
    strategyType: new FormControl('RollingUpdate'),
    maxUnavailable: new FormControl('25%'),
    maxSurge: new FormControl('25%')
  })

  // add/edit service
  serviceList = new FormArray([]);
  allPortList = [];
  addServiceForm = new FormGroup({
    services: this.serviceList
  });

  workloadPortsList = [];

  editServiceForm = new FormGroup({
    services: this.serviceList
  });

  // add/edit other resource
  labelsList = new FormArray([]);
  hpaReplicasInvaild = false;
  addOtherResourceForm = new FormGroup({
    labels: this.labelsList,
    enableHpa: new FormControl(false),
    hpaMetrics: new FormControl('CPU'),
    percentage: new FormControl(80, [Validators.min(1), Validators.max(99)]),
    minReplicas: new FormControl(1, [Validators.min(0)]),
    maxReplicas: new FormControl(10, [Validators.min(1)])
  })

  editOtherResourceForm = new FormGroup({
    labels: this.labelsList,
    enableHpa: new FormControl(false),
    hpaMetrics: new FormControl('CPU'),
    percentage: new FormControl(80, [Validators.min(1), Validators.max(99)]),
    minReplicas: new FormControl(1, [Validators.min(0)]),
    maxReplicas: new FormControl(10, [Validators.min(1)])
  })

  // component
  createComponentOpen = false;
  updateComponentOpen = false;
  cloneComponentOpen = false;
  removeDialog = false;

  // RAW params
  @ViewChild('fileUpload', { static: false }) fileUpload: ElementRef;
  @ViewChild('editFileUpload', { static: false }) editFileUpload: ElementRef;
  editorOptions = { theme: 'vs', language: 'yaml', automaticLayout: true };
  manifests: string;
  showAddFile = false;
  filename = '';
  showEditAddFile = false;

  // HELM params
  chartVersionForm = new FormGroup({
    registry: new FormControl(''),
    project: new FormControl(''),
    name: new FormControl(''),
    version: new FormControl(''),
    url: new FormControl('')
  });

  chartVersionListSorted: ChartVersion[];
  valuesEditorOptions = { theme: 'vs', language: 'yaml', automaticLayout: true };
  originValues: string;
  newValues: string;
  checkFormateOk = false;
  parseError = '';
  showCheckedAlert = false;
  noChartVersionAlert = false;
  valueload = false;

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardcpt') wizardnewcomponent: ClrWizard;
  // used to reset DOM
  allComponentApplications: Application[] = [];
  showNoApplicationAlert = false;
  basicComponentForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    application: new FormControl('', [Validators.required]),
    type: new FormControl('', Validators.required)
  });

  readonly disAllowLabels = CustomValidatorDirective.disAllowLabels;
  @ViewChild('wizardeditcpt') wizardeditcomponent: ClrWizard;
  showUpdateComponentDescDialog = false;
  updateComponentDescForm = new FormGroup({
    description: new FormControl('')
  });
  //clone component
  @ViewChild('wizardclonecpt') wizardclonecomponent: ClrWizard;
  cloneComponentForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    application: new FormControl('', [Validators.required]),
    type: new FormControl('', Validators.required)
  });
  removeConfigMapStorage: RemoveBasicComponentTemplateConfigMap[] = [];
  removeSecretStorage: RemoveBasicComponentTemplateSecret[] = [];
  removePVCStprage: RemoveBasicComponentTemplatePersistentVolumeClaim[] = [];
  removeDeployment: RemoveBasicComponentTemplateDeployment[] = [];
  removeService: RemoveBasicComponentTemplateService[] = [];
  deleteComponentForm = new FormGroup({
    confirm: new FormControl('', [Validators.required])
  });

  ngOnInit (): void {
    this.getAllApplication();
  }

  refresh (state?: ClrDatagridStateInterface): void {
    this.loading = true;
    let filters;
    if (this.selectedApplication === 'ALL') {
      filters = [{ key: 'organization', value: [this.orgId] }];
    } else {
      filters = [{ key: 'application', value: [this.selectedApplication] }];
    }
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.componentService.getAllComponents(state?.page?.current ? state.page.current - 1 : 0, state?.page?.size ? state.page.size : Number(this.pagination.pageSize), filters).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.GET-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.allComponents = data.data?.Components?.edge?.nodes ? data.data.Components.edge.nodes : [];
      this.selectedComponent = null;
      this.total = data?.data?.Components?.totalCount ? data.data.Components.totalCount : 0;
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.GET-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => this.loading = false);
  }

  // load application in init page
  getAllApplication (): void {
    this.applicationLoading = true;
    this.applicationService.getAllApplicationsByOrg(this.orgId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.APPLICATION.GET-APP-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.allApplications = data.data?.Applications?.edge?.nodes ? data.data.Applications.edge.nodes : [];
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.APPLICATION.GET-APP-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      const cachedApplication = localStorage.getItem(SELECTED_APPLICATION)
      if (cachedApplication == null) {
        this.selectedApplication = 'ALL';
        localStorage.setItem(SELECTED_APPLICATION, 'ALL');
      } else {
        if (!this.allApplications.find(x => x.id === cachedApplication)) {
          this.selectedApplication = 'ALL';
          localStorage.setItem(SELECTED_APPLICATION, 'ALL');
        } else {
          this.selectedApplication = cachedApplication;
        }
      }
      this.applicationLoading = false
    });
  }

  refreshApplication (): void {
    this.applicationService.getAllApplicationsByOrg(this.orgId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.APPLICATION.GET-APP-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.allApplications = data.data?.Applications?.edge?.nodes ? data.data.Applications.edge.nodes : [];
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.APPLICATION.GET-APP-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
    });
  }

  changeApplication (appId: string): void {
    this.selectedApplication = appId;
    localStorage.setItem(SELECTED_APPLICATION, appId);
    this.refresh();
  }

  matchApplication (appId: string): boolean {
    return this.selectedApplication === appId;
  }

  // fix "ng-reflect-model shows correct value but not reflecting in input", see https://stackoverflow.com/questions/40663198/ng-reflect-model-shows-correct-value-but-not-reflecting-in-input
  applicationTrackBy (index: number, obj: any): any {
    return index;
  }

  showCreateApp (): void {
    this.showCreateAppModel = true;
  }

  doCreateApplication (): void {
    this.showCreateAppModel = false;
    const input: CreateApplication = {
      name: this.basicAppForm.value.name,
      organization: this.orgId
    };
    this.applicationService.createApplication(input).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.CREATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.APPLICATION.CREATE-APP-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.CreateApplication) {
        this.toast.show(`${data.data.CreateApplication.name} ` + this.translate.instant('MESSAGE.APPLICATION.CREATE-APP-MSG'), this.translate.instant('MESSAGE.APPLICATION.CREATE-APP-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
      }
      this.refreshApplication();
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.APPLICATION.CREATE-APP-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.basicAppForm.reset();
    });
  }

  showDeleteApp (): void {
    this.deleteApplicationList = deepCopy(this.allApplications);
    this.showDeleteAppModel = true;
  }

  deleteApplicationById (): void {
    this.showDeleteAppModel = false;
    this.applicationService.deleteApplication(this.deleteAppForm.value.id).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            if (each.message === 'object is busy') {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ' 请先清除应用下所有组件后再执行此操作！', this.translate.instant('MESSAGE.APPLICATION.DELETE-APP-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            } else {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.APPLICATION.DELETE-APP-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            }
          }
        }
      }
      if (data.data?.DeleteApplication) {
        this.toast.show(`${data.data.DeleteApplication.name} ` + this.translate.instant('MESSAGE.APPLICATION.DELETE-APP-MSG'), this.translate.instant('MESSAGE.APPLICATION.DELETE-APP-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
      }
      // delete current application selected
      if (this.selectedApplication === this.deleteAppForm.value.id) {
        localStorage.removeItem(SELECTED_APPLICATION);
        this.getAllApplication();
      } else {
        this.refreshApplication();
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.APPLICATION.DELETE-APP-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.deleteApplicationList = null;
      this.deleteAppForm.reset();
    });
  }

  // add/edit storage(configmap, secret and pvc)
  addNewConfigMap (): void {
    this.showAddConfgiMap = true;
  }

  addNewSecret (): void {
    this.showAddSecret = true;
  }

  addNewPVC (): void {
    this.showAddPVC = true;
  }

  deleteStorage (index: number, update?: boolean): void {
    const storageName = this.allStorages[index].name;
    const storageType = this.allStorages[index].storageType;
    // remove dockerconfig secret
    if (storageType === 'secret' && (this.allStorages[index].secretType === 'kubernetes.io/dockerconfigjson' || this.allStorages[index].secretType === 'kubernetes.io/dockercfg')) {
      const pullSecretList = [];
      if (update) {
        for (const pullSecret of this.editWorkloadParamsForm.value.imagePullSecrets) {
          if (pullSecret !== storageName) {
            pullSecretList.push(pullSecret)
          }
        }
        this.editWorkloadParamsForm.patchValue({
          imagePullSecrets: pullSecretList
        })
      } else {
        for (const pullSecret of this.workloadParamsForm.value.imagePullSecrets) {
          if (pullSecret !== storageName) {
            pullSecretList.push(pullSecret)
          }
        }
        this.workloadParamsForm.patchValue({
          imagePullSecrets: pullSecretList
        })
      }
    } else {
      // remove volume in workload
      if (this.allVolumes.length > 0) {
        for (let i = 0; i < this.allVolumes.length; i++) {
          if (!!this.allVolumes[i][storageType] && storageName === this.allVolumes[i][storageType].name) {
            // delete volumeMount in container
            if (this.allContainers.length > 0) {
              for (const container of this.allContainers) {
                if (container.volumeMounts) {
                  container.volumeMounts.forEach((value, index) => {
                    if (value.name === this.allVolumes[i].name) {
                      container.volumeMounts.splice(index, 1);
                    }
                  })
                }
              }
            }
            if (this.allInitContainers.length > 0) {
              for (const initContainer of this.allInitContainers) {
                if (initContainer.volumeMounts) {
                  initContainer.volumeMounts.forEach((value, index) => {
                    if (value.name === this.allVolumes[i].name) {
                      initContainer.volumeMounts.splice(index, 1);
                    }
                  })
                }
              }
            }
            this.allVolumes.splice(i, 1)
            break;
          }
        }
      }
    }
    this.allStorages.splice(index, 1)
  }

  deleteCMStorageData (index: number): void {
    this.cmStorageDataList.removeAt(index);
  }

  addCMStorageKeyValueData (): void {
    this.cmStorageDataList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  showAddCMStorageFileData (): void {
    this.showAddCMStorageFile = true;
  }

  addCMStorageFileToData (): void {
    const fileUpload = this.fileCMUpload.nativeElement;
    const file = fileUpload.files[0];
    if (file) {
      const reader = new FileReader();
      reader.readAsText(file, 'UTF-8');
      reader.onload = () => {
        const fileContext = reader.result;
        this.cmStorageDataList.push(new FormGroup({
          key: new FormControl(file.name, Validators.required),
          value: new FormControl(fileContext, Validators.required)
        }));
      };
    }
    this.fileCMUpload.nativeElement.value = '';
    this.cmStorageFilename = null;
    this.showAddCMStorageFile = false;
  }

  cancelAddCMStorageFile (): void {
    this.fileCMUpload.nativeElement.value = null;
    this.cmStorageFilename = null;
    this.showAddCMStorageFile = false;
  }

  getCMStorageFileName (): void {
    const fileUpload = this.fileCMUpload.nativeElement;
    const file = fileUpload.files[0];
    this.cmStorageFilename = file ? file.name : '';
  }

  deleteGenericSecretData (index: number): void {
    this.genericSecretDataList.removeAt(index);
  }

  addGenericSecretData (): void {
    this.genericSecretDataList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required),
      nonFile: new FormControl(true)
    }));
  }

  showAddSecretStorageFileData (): void {
    this.showAddSecretStorageFile = true;
  }

  addSecretStorageFileToData (): void {
    const fileUpload = this.fileSecretUpload.nativeElement;
    const file = fileUpload.files[0];
    if (file) {
      const reader = new FileReader();
      reader.readAsText(file, 'UTF-8');
      reader.onload = () => {
        const fileContext = reader.result;
        this.genericSecretDataList.push(new FormGroup({
          key: new FormControl(file.name, Validators.required),
          value: new FormControl(fileContext, Validators.required),
          nonFile: new FormControl(false)
        }));
      };
    }
    this.fileSecretUpload.nativeElement.value = '';
    this.secrteStorageFilename = null;
    this.showAddSecretStorageFile = false;
  }

  cancelAddSecretStorageFile (): void {
    this.fileSecretUpload.nativeElement.value = null;
    this.secrteStorageFilename = null;
    this.showAddSecretStorageFile = false;
  }

  getSecretStorageFileName (): void {
    const fileUpload = this.fileSecretUpload.nativeElement;
    const file = fileUpload.files[0];
    this.secrteStorageFilename = file ? file.name : '';
  }

  doAddStorageFinish (storageType: string): void {
    switch (storageType) {
      case 'configMap': {
        const dataMap: JSON = JSON.parse(JSON.stringify({}));
        if (this.cmStorageDataForm.value.cmData) {
          for (const each of this.cmStorageDataForm.value.cmData) {
            dataMap[each.key] = each.value;
          }
        }
        this.allStorages.push({
          name: this.cmStorageDataForm.value.name,
          storageType: 'configMap',
          configMap: {
            data: dataMap
          }
        })
        break;
      }
      case 'secret': {
        const secretType = this.basicSecretForm.value.secretType;
        switch (secretType) {
          case 'Opaque': {
            const dataMap: JSON = JSON.parse(JSON.stringify({}));
            if (this.genericSecretDataForm.value.data) {
              for (const each of this.genericSecretDataForm.value.data) {
                dataMap[each.key] = each.value;
              }
            }
            this.allStorages.push({
              name: this.basicSecretForm.value.name,
              storageType: 'secret',
              secretType: 'Opaque',
              secret: {
                data: dataMap
              }
            })
            break;
          }
          case 'kubernetes.io/dockerconfigjson': {
            const auth = btoa(this.dockerConfigJSONSecretForm.value.authUser + ':' + this.dockerConfigJSONSecretForm.value.authPassword)
            const dockerConfigJson = '{"auths":{"' + this.dockerConfigJSONSecretForm.value.endpoint + '":{"username":"' + this.dockerConfigJSONSecretForm.value.authUser + '","password":"' + this.dockerConfigJSONSecretForm.value.authPassword + '","email":"' + (this.dockerConfigJSONSecretForm.value.email ? this.dockerConfigJSONSecretForm.value.email : '') + '","auth":"' + auth + '"}}}'
            const dockerConfigJsonDataMap: JSON = JSON.parse(JSON.stringify({}));
            dockerConfigJsonDataMap['.dockerconfigjson'] = dockerConfigJson;
            this.allStorages.push({
              name: this.basicSecretForm.value.name,
              storageType: 'secret',
              secretType: 'kubernetes.io/dockerconfigjson',
              secret: {
                data: dockerConfigJsonDataMap
              }
            })
            break;
          }
          case 'kubernetes.io/dockercfg': {
            const dockerConfigDataMap: JSON = JSON.parse(JSON.stringify({}));
            dockerConfigDataMap['.dockercfg'] = this.dockerConfigSecretForm.value.config;
            this.allStorages.push({
              name: this.basicSecretForm.value.name,
              storageType: 'secret',
              secretType: 'kubernetes.io/dockercfg',
              secret: {
                data: dockerConfigDataMap
              }
            })
            break;
          }
          case 'kubernetes.io/basic-auth': {
            this.allStorages.push({
              name: this.basicSecretForm.value.name,
              storageType: 'secret',
              secretType: 'kubernetes.io/basic-auth',
              secret: {
                data: JSON.parse(JSON.stringify({ username: this.basicAuthSecretForm.value.authUser, password: this.basicAuthSecretForm.value.authPassword }))
              }
            })
            break;
          }
          case 'kubernetes.io/ssh-auth': {
            const SSHAuthDataMap: JSON = JSON.parse(JSON.stringify({}));
            SSHAuthDataMap['ssh-privatekey'] = this.sshAuthSecretForm.value.privateKey;
            this.allStorages.push({
              name: this.basicSecretForm.value.name,
              storageType: 'secret',
              secretType: 'kubernetes.io/ssh-auth',
              secret: {
                data: SSHAuthDataMap
              }
            })
            break;
          }
          case 'kubernetes.io/tls': {
            this.allStorages.push({
              name: this.basicSecretForm.value.name,
              storageType: 'secret',
              secretType: 'kubernetes.io/tls',
              secret: {
                data: JSON.parse(JSON.stringify({ 'tls.crt': this.TLSSecretForm.value.cert, 'tls.key': this.TLSSecretForm.value.key }))
              }
            })
            break;
          }
          default:
            console.log('not support secret type')
            break;
        }
        break;
      }
      case 'persistentVolumeClaim': {
        this.allStorages.push({
          name: this.basicPVCForm.value.name,
          storageType: 'persistentVolumeClaim',
          persistentVolumeClaim: {
            spec: {
              accessModes: [this.basicPVCForm.value.accessMode],
              resources: {
                requests: {
                  storage: this.basicPVCForm.value.request + 'Gi'
                },
                limits: {
                  storage: this.basicPVCForm.value.limit ? this.basicPVCForm.value.limit + 'Gi' : null
                }
              },
              storageClass: this.basicPVCForm.value.storageClass ? this.basicPVCForm.value.storageClass : null,
              volumeMode: this.basicPVCForm.value.volumeMode
            }
          }
        })
        break;
      }
      default:
        console.log('not support')
    }
    this.doAddStorageReset();
  }

  doAddStorageCancel (): void {
    this.doAddStorageReset();
  }

  doAddStorageReset (): void {
    this.showAddConfgiMap = false;
    this.showAddSecret = false;
    this.showAddPVC = false;
    this.wizardsecret.reset();
    this.basicSecretForm.reset();
    this.basicSecretForm.patchValue({
      secretType: 'Opaque'
    })
    this.cmStorageFilename = '';
    this.cmStorageDataForm.reset();
    while (this.cmStorageDataList.length !== 0) {
      this.cmStorageDataList.removeAt(0);
    }
    this.secrteStorageFilename = '';
    this.genericSecretDataForm.reset();
    while (this.genericSecretDataList.length !== 0) {
      this.genericSecretDataList.removeAt(0);
    }
    this.dockerConfigJSONSecretForm.reset();
    this.basicAuthSecretForm.reset();
    this.sshAuthSecretForm.reset();
    this.TLSSecretForm.reset();
    this.basicPVCForm.reset();
    this.basicPVCForm.patchValue({
      volumeMode: 'Filesystem'
    })
  }

  showStorageDetails (storage: Storage, index: number): void {
    this.addStorageEditIndex = index;
    const storageType = storage.storageType;
    switch (storageType) {
      case 'configMap': {
        if (storage.configMap.data) {
          const datas = storage.configMap.data;
          for (const key of Object.keys(datas)) {
            const value = datas[key];
            this.cmEditStorageDataList.push(new FormGroup({
              key: new FormControl(key),
              value: new FormControl(value)
            }));
          }
        }
        this.cmEditStorageDataForm.patchValue({
          name: storage.name
        })
        this.showAddEditConfigmapStorage = true;
        break;
      }
      case 'secret': {
        this.editSecretForm.setValue({
          name: storage.name,
          secretType: storage.secretType
        })
        switch (storage.secretType) {
          case 'Opaque': {
            if (storage.secret.data) {
              const datas = storage.secret.data;
              for (const key of Object.keys(datas)) {
                const value = datas[key];
                const match = /[\r\n]/.exec(value);
                if (!match) {
                  this.editGenericSecretDataList.push(new FormGroup({
                    key: new FormControl(key),
                    value: new FormControl(value),
                    nonFile: new FormControl(true)
                  }));
                } else {
                  this.editGenericSecretDataList.push(new FormGroup({
                    key: new FormControl(key),
                    value: new FormControl(value),
                    nonFile: new FormControl(false)
                  }));
                }
              }
            }
            break;
          }
          case 'kubernetes.io/dockerconfigjson': {
            const dockerConfigJsonData = JSON.parse(storage.secret.data['.dockerconfigjson']);
            const ep = Object.keys(dockerConfigJsonData.auths)[0];
            let email = '';
            // eslint-disable-next-line no-prototype-builtins
            if (dockerConfigJsonData.auths[ep].hasOwnProperty('email')) {
              email = dockerConfigJsonData.auths[ep].email;
            }
            this.editDockerConfigJSONSecretForm.setValue({
              endpoint: ep,
              authUser: dockerConfigJsonData.auths[ep].username,
              authPassword: dockerConfigJsonData.auths[ep].password,
              email: email
            });
            break;
          }
          case 'kubernetes.io/dockercfg': {
            this.editDockerConfigSecretForm.setValue({
              config: storage.secret.data['.dockercfg']
            });
            break;
          }
          case 'kubernetes.io/basic-auth': {
            this.editBasicAuthSecretForm.setValue({
              authUser: storage.secret.data['username'],
              authPassword: storage.secret.data['password']
            });
            break;
          }
          case 'kubernetes.io/ssh-auth': {
            this.editSshAuthSecretForm.setValue({
              privateKey: storage.secret.data['ssh-privatekey']
            });
            break;
          }
          case 'kubernetes.io/tls': {
            this.editTLSSecretForm.setValue({
              cert: storage.secret.data['tls.crt'],
              key: storage.secret.data['tls.key']
            });
            break;
          }
          default:
            console.log('not support')
            break;
        }
        this.showAddEditSecretStorage = true;
        break;
      }
      case 'persistentVolumeClaim': {
        const requestUnit = String(storage.persistentVolumeClaim.spec.resources.requests.storage);
        const limitUnit = String(storage.persistentVolumeClaim.spec.resources.limits?.storage);
        this.editPVCForm.setValue({
          name: storage.name,
          storageClass: storage.persistentVolumeClaim.spec.storageClass ? storage.persistentVolumeClaim.spec.storageClass : '',
          accessMode: storage.persistentVolumeClaim.spec.accessModes[0],
          request: requestUnit.split('Gi', 2)[0],
          limit: storage.persistentVolumeClaim.spec.resources.limits ? limitUnit.split('Gi', 2)[0] : null,
          volumeMode: storage.persistentVolumeClaim.spec.volumeMode
        })
        this.showAddEditPVCStorage = true;
        break;
      }
      default:
        console.log('not support')
    }
  }

  deleteEditCMStorageData (index: number): void {
    this.cmEditStorageDataList.removeAt(index);
  }

  addEditCMStorageKeyValueData (): void {
    this.cmEditStorageDataList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  showAddEditCMStorageFileData (): void {
    this.showAddEditCMStorageFile = true;
  }

  addEditCMStorageFileToData (): void {
    const fileUpload = this.fileEditCMUpload.nativeElement;
    const file = fileUpload.files[0];
    if (file) {
      const reader = new FileReader();
      reader.readAsText(file, 'UTF-8');
      reader.onload = () => {
        const fileContext = reader.result;
        this.cmEditStorageDataList.push(new FormGroup({
          key: new FormControl(file.name, Validators.required),
          value: new FormControl(fileContext, Validators.required)
        }));
      };
    }
    this.fileEditCMUpload.nativeElement.value = '';
    this.cmEditStorageFilename = null;
    this.showAddEditCMStorageFile = false;
  }

  cancelAddEditCMStorageFile (): void {
    this.fileEditCMUpload.nativeElement.value = null;
    this.cmEditStorageFilename = null;
    this.showAddEditCMStorageFile = false;
  }

  getEditCMStorageFileName (): void {
    const fileUpload = this.fileEditCMUpload.nativeElement;
    const file = fileUpload.files[0];
    this.cmEditStorageFilename = file ? file.name : '';
  }

  deleteEditGenericSecretData (index: number): void {
    this.editGenericSecretDataList.removeAt(index);
  }

  addEditGenericSecretData (): void {
    this.editGenericSecretDataList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required),
      nonFile: new FormControl(true)
    }));
  }

  showAddEditFileData (): void {
    this.showAddEditSecretStorageFile = true;
  }

  addEditSecretStorageFileToData (): void {
    const fileUpload = this.fileEditSecretUpload.nativeElement;
    const file = fileUpload.files[0];
    if (file) {
      const reader = new FileReader();
      reader.readAsText(file, 'UTF-8');
      reader.onload = () => {
        const fileContext = reader.result;
        this.editGenericSecretDataList.push(new FormGroup({
          key: new FormControl(file.name, Validators.required),
          value: new FormControl(fileContext, Validators.required),
          nonFile: new FormControl(false)
        }));
      };
    }
    this.fileEditSecretUpload.nativeElement.value = '';
    this.secrteEditStorageFilename = null;
    this.showAddEditSecretStorageFile = false;
  }

  cancelAddEditSecretStorageFile (): void {
    this.fileEditSecretUpload.nativeElement.value = null;
    this.secrteEditStorageFilename = null;
    this.showAddEditSecretStorageFile = false;
  }

  getEditSecretStorageFileName (): void {
    const fileUpload = this.fileEditSecretUpload.nativeElement;
    const file = fileUpload.files[0];
    this.secrteEditStorageFilename = file ? file.name : '';
  }

  doAddEditStorageFinish (storageType: string): void {
    switch (storageType) {
      case 'configMap': {
        const dataMap: JSON = JSON.parse(JSON.stringify({}));
        if (this.cmEditStorageDataForm.value.cmData) {
          for (const each of this.cmEditStorageDataForm.value.cmData) {
            dataMap[each.key] = each.value;
          }
        }
        this.allStorages[this.addStorageEditIndex] = {
          name: this.cmEditStorageDataForm.value.name,
          storageType: 'configMap',
          configMap: {
            data: dataMap
          }
        }
        break;
      }
      case 'secret': {
        const secretType = this.editSecretForm.value.secretType;
        switch (secretType) {
          case 'Opaque': {
            const dataMap: JSON = JSON.parse(JSON.stringify({}));
            if (this.editGenericSecretDataForm.value.data) {
              for (const each of this.editGenericSecretDataForm.value.data) {
                dataMap[each.key] = each.value;
              }
            }
            this.allStorages[this.addStorageEditIndex] = {
              name: this.editSecretForm.value.name,
              storageType: 'secret',
              secretType: 'Opaque',
              secret: {
                data: dataMap
              }
            }
            break;
          }
          case 'kubernetes.io/dockerconfigjson': {
            const auth = btoa(this.editDockerConfigJSONSecretForm.value.authUser + ':' + this.editDockerConfigJSONSecretForm.value.authPassword)
            const dockerConfigJson = '{"auths":{"' + this.editDockerConfigJSONSecretForm.value.endpoint + '":{"username":"' + this.editDockerConfigJSONSecretForm.value.authUser + '","password":"' + this.editDockerConfigJSONSecretForm.value.authPassword + '","email":"' + (this.editDockerConfigJSONSecretForm.value.email ? this.editDockerConfigJSONSecretForm.value.email : '') + '","auth":"' + auth + '"}}}'
            const dockerConfigJsonDataMap: JSON = JSON.parse(JSON.stringify({}));
            dockerConfigJsonDataMap['.dockerconfigjson'] = dockerConfigJson;
            this.allStorages[this.addStorageEditIndex] = {
              name: this.editSecretForm.value.name,
              storageType: 'secret',
              secretType: 'kubernetes.io/dockerconfigjson',
              secret: {
                data: dockerConfigJsonDataMap
              }
            }
            break;
          }
          case 'kubernetes.io/dockercfg': {
            const dockerConfigDataMap: JSON = JSON.parse(JSON.stringify({}));
            dockerConfigDataMap['.dockercfg'] = this.editDockerConfigSecretForm.value.config;
            this.allStorages[this.addStorageEditIndex] = {
              name: this.editSecretForm.value.name,
              storageType: 'secret',
              secretType: 'kubernetes.io/dockercfg',
              secret: {
                data: dockerConfigDataMap
              }
            }
            break;
          }
          case 'kubernetes.io/basic-auth': {
            const basicAuthDataMap: JSON = JSON.parse(JSON.stringify({}));
            basicAuthDataMap['username'] = this.editBasicAuthSecretForm.value.authUser;
            basicAuthDataMap['password'] = this.editBasicAuthSecretForm.value.authPassword;
            this.allStorages[this.addStorageEditIndex] = {
              name: this.editSecretForm.value.name,
              storageType: 'secret',
              secretType: 'kubernetes.io/basic-auth',
              secret: {
                data: basicAuthDataMap
              }
            }
            break;
          }
          case 'kubernetes.io/ssh-auth': {
            const SSHAuthDataMap: JSON = JSON.parse(JSON.stringify({}));
            SSHAuthDataMap['ssh-privatekey'] = this.editSshAuthSecretForm.value.privateKey;
            this.allStorages[this.addStorageEditIndex] = {
              name: this.editSecretForm.value.name,
              storageType: 'secret',
              secretType: 'kubernetes.io/ssh-auth',
              secret: {
                data: SSHAuthDataMap
              }
            }
            break;
          }
          case 'kubernetes.io/tls': {
            const TLSDataMap: JSON = JSON.parse(JSON.stringify({}));
            TLSDataMap['tls.crt'] = this.editTLSSecretForm.value.cert;
            TLSDataMap['tls.key'] = this.editTLSSecretForm.value.key;
            this.allStorages[this.addStorageEditIndex] = {
              name: this.editSecretForm.value.name,
              storageType: 'secret',
              secretType: 'kubernetes.io/tls',
              secret: {
                data: TLSDataMap
              }
            }
            break;
          }
          default:
            console.log('not support secret type')
        }
        break;
      }
      case 'persistentVolumeClaim': {
        this.allStorages[this.addStorageEditIndex] = {
          name: this.editPVCForm.value.name,
          storageType: 'persistentVolumeClaim',
          persistentVolumeClaim: {
            spec: {
              accessModes: [this.editPVCForm.value.accessMode],
              resources: {
                requests: {
                  storage: this.editPVCForm.value.request + 'Gi'
                },
                limits: {
                  storage: this.editPVCForm.value.limit ? this.editPVCForm.value.limit + 'Gi' : null
                }
              },
              storageClass: this.editPVCForm.value.storageClass ? this.editPVCForm.value.storageClass : null,
              volumeMode: this.editPVCForm.value.volumeMode
            }
          }
        }
        break;
      }
      default:
        console.log('not support')
    }
    this.doAddEditStorageReset();
  }

  doAddEditStorageCancel (): void {
    this.doAddEditStorageReset();
  }

  doAddEditStorageReset (): void {
    this.showAddEditConfigmapStorage = false;
    this.showAddEditPVCStorage = false;
    this.showAddEditSecretStorage = false;
    this.wizardeditsecret.reset();
    this.editSecretForm.reset();
    this.cmEditStorageFilename = '';
    this.cmEditStorageDataForm.reset();
    while (this.cmEditStorageDataList.length !== 0) {
      this.cmEditStorageDataList.removeAt(0);
    }
    this.secrteEditStorageFilename = '';
    this.editGenericSecretDataForm.reset();
    while (this.editGenericSecretDataList.length !== 0) {
      this.editGenericSecretDataList.removeAt(0);
    }
    this.editDockerConfigJSONSecretForm.reset();
    this.editDockerConfigSecretForm.reset();
    this.editBasicAuthSecretForm.reset();
    this.editSshAuthSecretForm.reset();
    this.editTLSSecretForm.reset();
    this.editPVCForm.reset();
  }

  // add/edit volume in workload
  doStoragePageCustomNext (buttonType: string): void {
    this.allConfigMapStorages = [];
    this.allSecretStorages = [];
    this.allPVCStorages = [];
    this.dockerConfigSecretList = null;
    if (buttonType === 'custom-next') {
      this.configMapStorageExist = !!this.allStorages.find(x => x.storageType === 'configMap');
      this.secretStorageExist = !!this.allStorages.find(x => x.storageType === 'secret');
      this.pvcStorageExist = !!this.allStorages.find(x => x.storageType === 'persistentVolumeClaim');
    }
    this.getDockerConfigSecret();
    this.wizardnewcomponent.next();
  }

  doEditStoragePageCustomNext (buttonType: string): void {
    this.allConfigMapStorages = [];
    this.allSecretStorages = [];
    this.allPVCStorages = [];
    this.dockerConfigSecretList = null;
    if (buttonType === 'custom-next') {
      this.configMapStorageExist = !!this.allStorages.find(x => x.storageType === 'configMap');
      this.secretStorageExist = !!this.allStorages.find(x => x.storageType === 'secret');
      this.pvcStorageExist = !!this.allStorages.find(x => x.storageType === 'persistentVolumeClaim');
    }
    this.getDockerConfigSecret();
    this.wizardeditcomponent.next();
  }

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  getAllKeysOfCM ($event: any): void {
    this.cmKeysList = null;
    const cmName = $event.target.value;
    for (const cm of this.allConfigMapStorages) {
      if (cm.name === cmName) {
        this.cmKeysList = cm.configMap.data ? Object.keys(cm.configMap.data) : [];
      }
    }
    this.AddConfigMapVolumeForm.patchValue({
      name: cmName
    })
  }

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  getAllKeysOfSecret ($event: any): void {
    this.secretKeysList = null;
    const secretName = $event.target.value;
    for (const secret of this.allSecretStorages) {
      if (secret.name === secretName) {
        this.secretKeysList = secret.secret.data ? Object.keys(secret.secret.data) : [];
      }
    }
    this.AddSecretVolumeForm.patchValue({
      name: secretName
    })
  }

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  setVolumeName ($event: any): void {
    this.AddPVCVolumeForm.patchValue({
      name: $event.target.value
    })
  }

  addEmptyDirVolume (): void {
    this.showAddEmptyDirVolume = true;
  }

  addConfigMapVolume (): void {
    this.allConfigMapStorages = null;
    const tempList = [];
    for (const storage of this.allStorages) {
      if (storage.storageType === 'configMap') {
        tempList.push(storage)
      }
    }
    this.allConfigMapStorages = tempList;
    this.showAddComnfigMapVolume = true;
  }

  addSecretVolume (): void {
    // reset dom
    this.allSecretStorages = null;
    const tempList = [];
    for (const storage of this.allStorages) {
      if (storage.storageType === 'secret') {
        tempList.push(storage)
      }
    }
    this.allSecretStorages = tempList;
    this.showAddSecretVolume = true;
  }

  addPVCVolume (): void {
    this.allPVCStorages = null;
    const tempList = [];
    for (const storage of this.allStorages) {
      if (storage.storageType === 'persistentVolumeClaim') {
        tempList.push(storage)
      }
    }
    this.allPVCStorages = tempList;
    this.showAddPVCVolume = true;
  }

  deleteVolume (volumeIndex: number): void {
    const volumeName = this.allVolumes[volumeIndex].name;
    for (const container of this.allContainers) {
      if (container.volumeMounts) {
        container.volumeMounts.forEach((value, index) => {
          if (value.name === volumeName) {
            container.volumeMounts.splice(index, 1);
          }
        })
      }
    }
    for (const initContainer of this.allInitContainers) {
      if (initContainer.volumeMounts) {
        initContainer.volumeMounts.forEach((value, index) => {
          if (value.name === volumeName) {
            initContainer.volumeMounts.splice(index, 1);
          }
        })
      }
    }
    this.allVolumes.splice(volumeIndex, 1);
  }

  addNewConfigmapItems (): void {
    this.cmItemsList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      path: new FormControl(''),
      mode: new FormControl('')
    }));
  }

  deleteConfigmapItems (index: number): void {
    this.cmItemsList.removeAt(index);
  }

  removeCMItemsList (enabled: boolean): void {
    if (!enabled) {
      while (this.cmItemsList.length > 0) {
        this.cmItemsList.removeAt(0)
      }
    }
  }

  addNewSecretItems (): void {
    this.secretItemsList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      path: new FormControl(''),
      mode: new FormControl('')
    }));
  }

  deleteSecretItems (index: number): void {
    this.secretItemsList.removeAt(index);
  }

  removeSecretItemsList (enabled: boolean): void {
    if (!enabled) {
      while (this.secretItemsList.length > 0) {
        this.secretItemsList.removeAt(0)
      }
    }
  }

  doAddNewVolumes (volumeType: string): void {
    switch (volumeType) {
      case 'emptyDir': {
        this.allVolumes.push({
          name: this.AddEmptyVolumeForm.value.name,
          emptyDir: {
            medium: this.AddEmptyVolumeForm.value.empty_medium === 'default' ? '' : 'Memory',
            size: this.AddEmptyVolumeForm.value.empty_size ? this.AddEmptyVolumeForm.value.empty_size + 'Gi' : '1Gi'
          },
          configMap: null,
          secret: null,
          persistentVolumeClaim: null
        });
        break;
      }
      case 'configMap': {
        const cmMountList: TemplateSpecVolumeKeyToPathInput[] = [];
        if (this.AddConfigMapVolumeForm.value.enableItems) {
          for (const each of this.AddConfigMapVolumeForm.value.cmItems) {
            cmMountList.push({
              key: each.key,
              path: each.path ? each.path : each.key,
              mode: each.mode ? each.mode : null
            })
          }
        }
        this.allVolumes.push({
          name: this.AddConfigMapVolumeForm.value.name,
          emptyDir: null,
          configMap: {
            name: this.AddConfigMapVolumeForm.value.cmName,
            defaultMode: this.AddConfigMapVolumeForm.value.defaultMode ? this.AddConfigMapVolumeForm.value.defaultMode : 420,
            items: cmMountList
          },
          secret: null,
          persistentVolumeClaim: null
        });
        break;
      }
      case 'secret': {
        const secretMountList: TemplateSpecVolumeKeyToPathInput[] = [];
        if (this.AddSecretVolumeForm.value.enableItems) {
          for (const each of this.AddSecretVolumeForm.value.secretItems) {
            secretMountList.push({
              key: each.key,
              path: each.path ? each.path : each.key,
              mode: each.mode ? each.mode : null
            })
          }
        }
        this.allVolumes.push({
          name: this.AddSecretVolumeForm.value.name,
          emptyDir: null,
          configMap: null,
          secret: {
            name: this.AddSecretVolumeForm.value.secretName,
            defaultMode: this.AddSecretVolumeForm.value.defaultMode ? this.AddSecretVolumeForm.value.defaultMode : 420,
            items: secretMountList
          },
          persistentVolumeClaim: null
        });
        break;
      }
      case 'persistentVolumeClaim': {
        this.allVolumes.push({
          name: this.AddPVCVolumeForm.value.name,
          emptyDir: null,
          configMap: null,
          secret: null,
          persistentVolumeClaim: {
            name: this.AddPVCVolumeForm.value.pvcName,
            readOnly: this.AddPVCVolumeForm.value.readonly
          }
        });
        break;
      }
    }
    this.doAddNewVolumesReset();
  }

  doAddNewVolumeCancel (): void {
    this.doAddNewVolumesReset();
  }

  doAddNewVolumesReset (): void {
    this.AddEmptyVolumeForm.reset();
    this.AddEmptyVolumeForm.patchValue({
      medium: 'default',
      size: 1
    });
    this.AddConfigMapVolumeForm.reset();
    this.AddConfigMapVolumeForm.patchValue({
      defaultMode: 420
    })
    while (this.cmItemsList.length !== 0) {
      this.cmItemsList.removeAt(0);
    }
    this.AddSecretVolumeForm.reset();
    this.AddSecretVolumeForm.patchValue({
      defaultMode: 420
    })
    while (this.secretItemsList.length !== 0) {
      this.secretItemsList.removeAt(0);
    }
    this.AddPVCVolumeForm.reset();
    this.AddPVCVolumeForm.patchValue({
      readonly: false
    })
    this.cmKeysList = null;
    this.secretKeysList = null;
    this.showAddEmptyDirVolume = false;
    this.showAddComnfigMapVolume = false;
    this.showAddSecretVolume = false;
    this.showAddPVCVolume = false;
  }

  showVolumeDetails (volume: AddPodSpecVolume, index: number): void {
    this.addVolumeEditIndex = index;
    const volumeType = volume.configMap ? 'configMap' : volume.emptyDir ? 'emptyDir' : volume.secret ? 'secret' : volume.persistentVolumeClaim ? 'persistentVolumeClaim' : '-';
    switch (volumeType) {
      case 'emptyDir': {
        this.editEmptyVolumeForm.patchValue({
          name: volume.name,
          medium: !!volume.emptyDir.medium ? volume.emptyDir.medium : 'default',
          size: volume.emptyDir.size.split('Gi', 2)[0]
        });
        this.showEmptyVolumeDetails = true;
        break;
      }
      case 'configMap': {
        this.allConfigMapStorages = null;
        const cmtempList = [];
        for (const storage of this.allStorages) {
          if (storage.storageType === 'configMap') {
            cmtempList.push(storage)
          }
        }
        this.allConfigMapStorages = cmtempList;
        const cmItemsList = volume.configMap.items ? volume.configMap.items : [];
        for (const each of cmItemsList) {
          this.editAddCMItemsList.push(new FormGroup({
            key: new FormControl(each.key),
            path: new FormControl(each.path),
            mode: new FormControl(each.mode)
          }))
        }
        this.cmKeysList = null;
        for (const cm of this.allConfigMapStorages) {
          if (cm.name === volume.configMap.name) {
            this.cmKeysList = cm.configMap.data ? Object.keys(cm.configMap.data) : [];
          }
        }
        this.editConfigMapVolumeForm.patchValue({
          cmName: volume.configMap.name,
          name: volume.name,
          defaultMode: volume.configMap.defaultMode,
          enableItems: this.editAddCMItemsList.length > 0
        });
        this.showConfigmapVolumeDetails = true;
        break;
      }
      case 'secret': {
        this.allSecretStorages = null;
        const secrettempList = [];
        for (const storage of this.allStorages) {
          if (storage.storageType === 'secret') {
            secrettempList.push(storage)
          }
        }
        this.allSecretStorages = secrettempList;
        const secretItemsList = volume.secret.items ? volume.secret.items : [];
        for (const each of secretItemsList) {
          this.editAddSecretItemsList.push(new FormGroup({
            key: new FormControl(each.key),
            path: new FormControl(each.path),
            mode: new FormControl(each.mode)
          }))
        }
        this.secretKeysList = null;
        for (const secret of this.allSecretStorages) {
          if (secret.name === volume.secret.name) {
            this.secretKeysList = secret.secret.data ? Object.keys(secret.secret.data) : [];
          }
        }
        this.editSecretVolumeForm.patchValue({
          secretName: volume.secret.name,
          name: volume.name,
          defaultMode: volume.secret.defaultMode,
          enableItems: this.editAddSecretItemsList.length > 0
        });
        this.showSecretVolumeDetails = true;
        break;
      }
      case 'persistentVolumeClaim': {
        this.allPVCStorages = null;
        const pvctempList = [];
        for (const storage of this.allStorages) {
          if (storage.storageType === 'persistentVolumeClaim') {
            pvctempList.push(storage)
          }
        }
        this.allPVCStorages = pvctempList;
        this.editPVCVolumeForm.patchValue({
          pvcName: volume.persistentVolumeClaim.name,
          name: volume.name,
          readonly: volume.persistentVolumeClaim.readOnly
        });
        this.showPVCVolumeDetails = true;
        break;
      }
    }
  }

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  getAllKeysOfEditCM ($event: any): void {
    this.cmKeysList = null;
    const cmName = $event.target.value;
    for (const cm of this.allConfigMapStorages) {
      if (cm.name === cmName) {
        this.cmKeysList = cm.configMap.data ? Object.keys(cm.configMap.data) : [];
      }
    }
    this.editConfigMapVolumeForm.patchValue({
      name: cmName
    })
  }

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  getAllKeysOfEditSecret ($event: any): void {
    this.secretKeysList = null;
    const secretName = $event.target.value;
    for (const secret of this.allSecretStorages) {
      if (secret.name === secretName) {
        this.secretKeysList = secret.secret.data ? Object.keys(secret.secret.data) : [];
      }
    }
    this.editSecretVolumeForm.patchValue({
      name: secretName
    })
  }

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  setEditVolumeName ($event: any): void {
    this.editPVCVolumeForm.patchValue({
      name: $event.target.value
    })
  }

  addEditConfigmapItems (): void {
    this.editAddCMItemsList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      path: new FormControl(''),
      mode: new FormControl('')
    }));
  }

  deleteEditConfigmapItems (index: number): void {
    this.editAddCMItemsList.removeAt(index);
  }

  removeEditCMItemsList (enabled: boolean): void {
    if (!enabled) {
      while (this.editAddCMItemsList.length > 0) {
        this.editAddCMItemsList.removeAt(0)
      }
    }
  }

  addEditSecretItems (): void {
    this.editAddSecretItemsList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      path: new FormControl(''),
      mode: new FormControl('')
    }));
  }

  deleteEditSecretItems (index: number): void {
    this.editAddSecretItemsList.removeAt(index);
  }

  removeEditSecretItemsList (enabled: boolean): void {
    if (!enabled) {
      while (this.editAddSecretItemsList.length > 0) {
        this.editAddSecretItemsList.removeAt(0)
      }
    }
  }

  doAddEditVolumes (volumeType: string): void {
    switch (volumeType) {
      case 'emptyDir': {
        this.allVolumes[this.addVolumeEditIndex] = {
          name: this.editEmptyVolumeForm.value.name,
          emptyDir: {
            medium: this.editEmptyVolumeForm.value.empty_medium === 'default' ? '' : 'Memory',
            size: this.editEmptyVolumeForm.value.size ? this.editEmptyVolumeForm.value.size + 'Gi' : '1Gi'
          },
          configMap: null,
          secret: null,
          persistentVolumeClaim: null
        };
        break;
      }
      case 'configMap': {
        const cmMountList: TemplateSpecVolumeKeyToPathInput[] = [];
        if (this.editConfigMapVolumeForm.value.enableItems) {
          for (const each of this.editConfigMapVolumeForm.value.cmItems) {
            cmMountList.push({
              key: each.key,
              path: each.path ? each.path : each.key,
              mode: each.mode ? each.mode : null
            })
          }
        }
        this.allVolumes[this.addVolumeEditIndex] = {
          name: this.editConfigMapVolumeForm.value.name,
          emptyDir: null,
          configMap: {
            name: this.editConfigMapVolumeForm.value.cmName,
            defaultMode: this.editConfigMapVolumeForm.value.defaultMode ? this.editConfigMapVolumeForm.value.defaultMode : 420,
            items: cmMountList
          },
          secret: null,
          persistentVolumeClaim: null
        };
        break;
      }
      case 'secret': {
        const secretMountList: TemplateSpecVolumeKeyToPathInput[] = [];
        if (this.editSecretVolumeForm.value.enableItems) {
          for (const each of this.editSecretVolumeForm.value.secretItems) {
            secretMountList.push({
              key: each.key,
              path: each.path ? each.path : each.key,
              mode: each.mode ? each.mode : null
            })
          }
        }
        this.allVolumes[this.addVolumeEditIndex] = {
          name: this.editSecretVolumeForm.value.name,
          emptyDir: null,
          configMap: null,
          secret: {
            name: this.editSecretVolumeForm.value.secretName,
            defaultMode: this.editSecretVolumeForm.value.defaultMode ? this.editSecretVolumeForm.value.defaultMode : 420,
            items: secretMountList
          },
          persistentVolumeClaim: null
        };
        break;
      }
      case 'persistentVolumeClaim': {
        this.allVolumes[this.addVolumeEditIndex] = {
          name: this.editPVCVolumeForm.value.name,
          emptyDir: null,
          configMap: null,
          secret: null,
          persistentVolumeClaim: {
            name: this.editPVCVolumeForm.value.pvcName,
            readOnly: this.editPVCVolumeForm.value.readonly
          }
        };
        break;
      }
    }
    this.doAddEditVolumesReset();
  }

  doAddEditVolumeCancel (): void {
    this.doAddEditVolumesReset();
  }

  doAddEditVolumesReset (): void {
    this.editEmptyVolumeForm.reset();
    this.editEmptyVolumeForm.patchValue({
      medium: 'default',
      size: 1
    })
    this.editConfigMapVolumeForm.reset();
    this.editConfigMapVolumeForm.patchValue({
      defaultMode: 420,
      enableItems: false
    })
    while (this.editAddCMItemsList.length !== 0) {
      this.editAddCMItemsList.removeAt(0);
    }
    this.editSecretVolumeForm.reset();
    this.editSecretVolumeForm.patchValue({
      defaultMode: 420,
      enableItems: false
    })
    while (this.editAddSecretItemsList.length !== 0) {
      this.editAddSecretItemsList.removeAt(0);
    }
    this.editPVCVolumeForm.reset();
    this.editPVCVolumeForm.patchValue({
      readonly: false
    })
    this.showEmptyVolumeDetails = false;
    this.showConfigmapVolumeDetails = false;
    this.showSecretVolumeDetails = false;
    this.showPVCVolumeDetails = false;
    this.addVolumeEditIndex = null;
    this.cmKeysList = [];
    this.secretKeysList = [];
  }

  // add/edit container in workload
  checkCreatePortOverlap (port: number|string, protocol: string): void {
    // only check containers port
    // Adding a container for the first time does not check
    this.portOverlap = portOverlapValidator(this.allContainers, Number(port), protocol);
  }

  addNewContainer (): void {
    this.addNewPorts();
    this.checkCreatePortOverlap(8080, 'TCP');
    this.showAddContaienr = true;
  }

  addNewContainerCommand (): void {
    this.containerCommandList.push(new FormGroup({
      command: new FormControl('', Validators.required)
    }))
  }

  deleteContainerCommands (index: number): void {
    this.containerCommandList.removeAt(index);
  }

  addNewContainerArgs (): void {
    this.containerArgsList.push(new FormGroup({
      args: new FormControl('', Validators.required)
    }))
  }

  deleteContainerArgs (index: number): void {
    this.containerArgsList.removeAt(index);
  }

  addNewPorts (): void {
    this.ports.push(new FormGroup({
      portName: new FormControl('', Validators.required),
      protocol: new FormControl('TCP', Validators.required),
      containerPort: new FormControl(8080, Validators.required)
    }));
    this.checkCreatePortOverlap(8080, 'TCP');
  }

  deletePort (index: number): void {
    this.ports.removeAt(index);
  }

  addNewEnv (): void {
    this.envs.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  deleteEnv (index: number): void {
    this.envs.removeAt(index);
  }

  addNewMounts (): void {
    this.mounts.push(new FormGroup({
      name: new FormControl('', Validators.required),
      mountPath: new FormControl('', Validators.required),
      subPath: new FormControl('')
    }));
  }

  deleteMount (index: number): void {
    this.mounts.removeAt(index);
  }

  // add health check in container
  addNewReadinessCheck (): void {
    this.showAddHealthCheck = true;
    this.addHealthCheckType = 'Readiness';
  }

  addNewLivenessCheck (): void {
    this.showAddHealthCheck = true;
    this.addHealthCheckType = 'Liveness';
  }

  addNewStartupCheck (): void {
    this.showAddHealthCheck = true;
    this.addHealthCheckType = 'Startup';
  }

  deleteHealthCheck (index: number): void {
    this.allHealthChecks.splice(index, 1)
  }

  deleteHeaders (index: number): void {
    this.headersList.removeAt(index);
  }

  addNewHeader (): void {
    this.headersList.push(new FormGroup({
      name: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }))
  }

  addNewCommand (): void {
    this.commandList.push(new FormGroup({
      arg: new FormControl('', Validators.required)
    }))
  }

  deleteCommands (index: number): void {
    this.commandList.removeAt(index);
  }

  validHealthCheckType (type: string): boolean {
    for (const healthcheck of this.allHealthChecks) {
      if (healthcheck.checkType === type) {
        return true;
      }
    }
    return false;
  }

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  initCheckParamsOfType ($event: any): void {
    const type = $event.target.value;
    switch (type) {
      case 'EXEC':
        this.addNewCommand();
        this.AddHealthCheckForm.patchValue({
          port: 8080,
          tcpPort: 8080,
          failureThreshold: 3,
          successThreshold: 1,
          initDelay: 0,
          period: 10,
          timeout: 1
        })
        break;
      case 'HTTPGET':
        while (this.commandList.length !== 0) {
          this.commandList.removeAt(0);
        }
        this.AddHealthCheckForm.patchValue({
          https: false,
          path: '/',
          port: 8080,
          tcpPort: 8080,
          failureThreshold: 3,
          successThreshold: 1,
          initDelay: 0,
          period: 10,
          timeout: 1
        })
        break;
      case 'TCPSOCKET':
        while (this.commandList.length !== 0) {
          this.commandList.removeAt(0);
        }
        this.AddHealthCheckForm.patchValue({
          port: 8080,
          tcpPort: 8080,
          failureThreshold: 3,
          successThreshold: 1,
          initDelay: 0,
          period: 10,
          timeout: 1
        })
        break;
      default:
        break;
    }
  }

  doAddHealthCheckReset (): void {
    this.showAddHealthCheck = false;
    this.addHealthCheckType = '';
    this.AddHealthCheckForm.reset();
    this.AddHealthCheckForm.patchValue({
      handlerType: 'HTTPGET',
      https: false,
      path: '/',
      port: 8080,
      tcpPort: 8080,
      failureThreshold: 3,
      successThreshold: 1,
      initDelay: 0,
      period: 10,
      timeout: 1
    })
    while (this.headersList.length !== 0) {
      this.headersList.removeAt(0);
    }
    while (this.commandList.length !== 0) {
      this.commandList.removeAt(0);
    }
  }

  doAddHealthCheckCancel (): void {
    this.doAddHealthCheckReset();
  }

  doAddHalthCheck (): void {
    const headerList: InputPodSpecContainerHandlerHTTPGetActionHTTPHeader[] = [];
    for (const each of this.AddHealthCheckForm.value.headers) {
      headerList.push({
        name: each.name,
        value: each.value
      })
    }
    const commandList: string[] = [];
    for (const each of this.AddHealthCheckForm.value.commands) {
      commandList.push(each.arg)
    }
    this.allHealthChecks.push({
      checkType: this.addHealthCheckType,
      params: {
        exec: this.AddHealthCheckForm.value.handlerType === 'EXEC'
          ? {
              command: commandList
            }
          : null,
        httpGet: this.AddHealthCheckForm.value.handlerType === 'HTTPGET'
          ? {
              path: this.AddHealthCheckForm.value.path ? this.AddHealthCheckForm.value.path : '/',
              port: this.AddHealthCheckForm.value.port,
              scheme: this.AddHealthCheckForm.value.https ? 'https' : 'http',
              httpHeaders: headerList
            }
          : null,
        tcpSocket: this.AddHealthCheckForm.value.handlerType === 'TCPSOCKET'
          ? {
              port: this.AddHealthCheckForm.value.tcpPort
            }
          : null,
        initialDelaySeconds: this.AddHealthCheckForm.value.initDelay ? this.AddHealthCheckForm.value.initDelay : 0,
        timeoutSeconds: this.AddHealthCheckForm.value.timeout ? this.AddHealthCheckForm.value.timeout : 1,
        periodSeconds: this.AddHealthCheckForm.value.period ? this.AddHealthCheckForm.value.period : 10,
        successThreshold: this.AddHealthCheckForm.value.successThreshold ? this.AddHealthCheckForm.value.successThreshold : 1,
        failureThreshold: this.AddHealthCheckForm.value.failureThreshold ? this.AddHealthCheckForm.value.failureThreshold : 3
      }
    })
    this.doAddHealthCheckReset();
  }

  deleteEditHeaders (index: number): void {
    this.editHeadersList.removeAt(index);
  }

  addNewEditHeader (): void {
    this.editHeadersList.push(new FormGroup({
      name: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }))
  }

  addNewEditCommand (): void {
    this.editCommandList.push(new FormGroup({
      arg: new FormControl('', Validators.required)
    }))
  }

  deleteEditCommands (index: number): void {
    this.editCommandList.removeAt(index);
  }

  showHealthCheckDetails (healthcheck: {'checkType': string, 'params': InputPodSpecContainerProbe}, index: number): void {
    this.addHealthCheckEditIndex = index;
    this.addHealthCheckType = healthcheck.checkType;
    const handleType = healthcheck.params.exec ? 'EXEC' : healthcheck.params.httpGet ? 'HTTPGET' : healthcheck.params.tcpSocket ? 'TCPSOCKET' : 'UNKNOWN';
    switch (handleType) {
      case 'EXEC':
        for (const command of healthcheck.params.exec.command) {
          this.editCommandList.push(new FormGroup({
            arg: new FormControl(command)
          })
          )
        }
        this.editAddHealthCheckForm.patchValue({
          handlerType: 'EXEC',
          failureThreshold: healthcheck.params.failureThreshold,
          successThreshold: healthcheck.params.successThreshold,
          initDelay: healthcheck.params.initialDelaySeconds,
          period: healthcheck.params.periodSeconds,
          timeout: healthcheck.params.timeoutSeconds
        })
        break;
      case 'HTTPGET':
        for (const header of healthcheck.params.httpGet.httpHeaders) {
          this.editHeadersList.push(new FormGroup({
            name: new FormControl(header.name),
            value: new FormControl(header.value)
          })
          )
        }
        this.editAddHealthCheckForm.patchValue({
          handlerType: 'HTTPGET',
          https: healthcheck.params.httpGet.scheme === 'https',
          path: healthcheck.params.httpGet.path,
          port: healthcheck.params.httpGet.port,
          failureThreshold: healthcheck.params.failureThreshold,
          successThreshold: healthcheck.params.successThreshold,
          initDelay: healthcheck.params.initialDelaySeconds,
          period: healthcheck.params.periodSeconds,
          timeout: healthcheck.params.timeoutSeconds
        })
        break;
      case 'TCPSOCKET':
        this.editAddHealthCheckForm.patchValue({
          handlerType: 'TCPSOCKET',
          tcpPort: healthcheck.params.tcpSocket.port,
          failureThreshold: healthcheck.params.failureThreshold,
          successThreshold: healthcheck.params.successThreshold,
          initDelay: healthcheck.params.initialDelaySeconds,
          period: healthcheck.params.periodSeconds,
          timeout: healthcheck.params.timeoutSeconds
        })
        break;
      default:
        break;
    }
    this.showAddHealthCheckDetails = true;
  }

  doAddEditHealthCheckCancel (): void {
    this.doAddEditHealthCheckReset();
  }

  doAddEditHealthCheckReset (): void {
    this.showAddHealthCheckDetails = false;
    this.addHealthCheckType = '';
    this.editAddHealthCheckForm.reset();
    this.editAddHealthCheckForm.patchValue({
      handlerType: 'HTTPGET',
      https: false,
      path: '/',
      port: 8080,
      tcpPort: 8080,
      failureThreshold: 3,
      successThreshold: 1,
      initDelay: 0,
      period: 10,
      timeout: 1
    })
    while (this.editHeadersList.length !== 0) {
      this.editHeadersList.removeAt(0);
    }
    while (this.editCommandList.length !== 0) {
      this.editCommandList.removeAt(0);
    }
  }

  doAddEditHalthCheck (): void {
    const headerList: InputPodSpecContainerHandlerHTTPGetActionHTTPHeader[] = [];
    for (const each of this.editAddHealthCheckForm.value.headers) {
      headerList.push({
        name: each.name,
        value: each.value
      })
    }
    const commandList: string[] = [];
    for (const each of this.editAddHealthCheckForm.value.commands) {
      commandList.push(each.arg)
    }
    this.allHealthChecks[this.addHealthCheckEditIndex] = {
      checkType: this.addHealthCheckType,
      params: {
        exec: this.editAddHealthCheckForm.value.handlerType === 'EXEC'
          ? {
              command: commandList
            }
          : null,
        httpGet: this.editAddHealthCheckForm.value.handlerType === 'HTTPGET'
          ? {
              path: this.editAddHealthCheckForm.value.path ? this.editAddHealthCheckForm.value.path : '/',
              port: this.editAddHealthCheckForm.value.port,
              scheme: this.editAddHealthCheckForm.value.https ? 'https' : 'http',
              httpHeaders: headerList
            }
          : null,
        tcpSocket: this.editAddHealthCheckForm.value.handlerType === 'TCPSOCKET'
          ? {
              port: this.editAddHealthCheckForm.value.tcpPort
            }
          : null,
        initialDelaySeconds: this.editAddHealthCheckForm.value.initDelay ? this.editAddHealthCheckForm.value.initDelay : 0,
        timeoutSeconds: this.editAddHealthCheckForm.value.timeout ? this.editAddHealthCheckForm.value.timeout : 1,
        periodSeconds: this.editAddHealthCheckForm.value.period ? this.editAddHealthCheckForm.value.period : 10,
        successThreshold: this.editAddHealthCheckForm.value.successThreshold ? this.editAddHealthCheckForm.value.successThreshold : 1,
        failureThreshold: this.editAddHealthCheckForm.value.failureThreshold ? this.editAddHealthCheckForm.value.failureThreshold : 3
      }
    }
    this.doAddEditHealthCheckReset();
  }

  // add lifecycle hook in container
  addNewPostStartHook (): void {
    this.showAddLifecycleHook = true;
    this.addLifecycleHook = 'postStart';
  }

  addNewPreStopHook (): void {
    this.showAddLifecycleHook = true;
    this.addLifecycleHook = 'preStop';
  }

  deleteLifecycleHook (index: number): void {
    this.allLifecycleHook.splice(index, 1)
  }

  deleteHookHeaders (index: number): void {
    this.hookHeadersList.removeAt(index);
  }

  addNewHookHeader (): void {
    this.hookHeadersList.push(new FormGroup({
      name: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }))
  }

  addNewHookCommand (): void {
    this.hookCommandList.push(new FormGroup({
      arg: new FormControl('', Validators.required)
    }))
  }

  deleteHookCommands (index: number): void {
    this.hookCommandList.removeAt(index);
  }

  validLifecycleHookType (type: string): boolean {
    for (const hook of this.allLifecycleHook) {
      if (hook.hookType === type) {
        return true;
      }
    }
    return false;
  }

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  initHookParamsOfType ($event: any): void {
    const type = $event.target.value;
    switch (type) {
      case 'EXEC':
        this.addNewHookCommand();
        this.AddLiftcycleHookForm.patchValue({
          port: 8080,
          tcpPort: 8080
        })
        break;
      case 'HTTPGET':
        while (this.hookCommandList.length !== 0) {
          this.hookCommandList.removeAt(0);
        }
        this.AddLiftcycleHookForm.patchValue({
          https: false,
          path: '/',
          port: 8080,
          tcpPort: 8080
        })
        break;
      case 'TCPSOCKET':
        while (this.hookCommandList.length !== 0) {
          this.hookCommandList.removeAt(0);
        }
        this.AddLiftcycleHookForm.patchValue({
          port: 8080,
          tcpPort: 8080
        })
        break;
      default:
        break;
    }
  }

  doAddLifecycleHookReset (): void {
    this.showAddLifecycleHook = false;
    this.addLifecycleHook = '';
    this.AddLiftcycleHookForm.reset();
    this.AddLiftcycleHookForm.patchValue({
      handlerType: 'HTTPGET',
      https: false,
      path: '/',
      port: 8080,
      tcpPort: 8080
    })
    while (this.hookHeadersList.length !== 0) {
      this.hookHeadersList.removeAt(0);
    }
    while (this.hookCommandList.length !== 0) {
      this.hookCommandList.removeAt(0);
    }
  }

  doAddLifecycleHookCancel (): void {
    this.doAddLifecycleHookReset();
  }

  doAddLifecycleHook (): void {
    const headerList: InputPodSpecContainerHandlerHTTPGetActionHTTPHeader[] = [];
    for (const each of this.AddLiftcycleHookForm.value.headers) {
      headerList.push({
        name: each.name,
        value: each.value
      })
    }
    const commandList: string[] = [];
    for (const each of this.AddLiftcycleHookForm.value.commands) {
      commandList.push(each.arg)
    }
    this.allLifecycleHook.push({
      hookType: this.addLifecycleHook,
      params: {
        exec: this.AddLiftcycleHookForm.value.handlerType === 'EXEC'
          ? {
              command: commandList
            }
          : null,
        httpGet: this.AddLiftcycleHookForm.value.handlerType === 'HTTPGET'
          ? {
              path: this.AddLiftcycleHookForm.value.path ? this.AddLiftcycleHookForm.value.path : '/',
              port: this.AddLiftcycleHookForm.value.port,
              scheme: this.AddLiftcycleHookForm.value.https ? 'https' : 'http',
              httpHeaders: headerList
            }
          : null,
        tcpSocket: this.AddLiftcycleHookForm.value.handlerType === 'TCPSOCKET'
          ? {
              port: this.AddLiftcycleHookForm.value.tcpPort
            }
          : null
      }
    })
    this.doAddLifecycleHookReset();
  }

  showLifecycleHookDetails (hook: {'hookType': string, 'params': InputPodSpecContainerHandler}, index: number): void {
    this.addLifecycleHookEditIndex = index;
    this.addLifecycleHook = hook.hookType;
    const handleType = hook.params.exec ? 'EXEC' : hook.params.httpGet ? 'HTTPGET' : hook.params.tcpSocket ? 'TCPSOCKET' : 'UNKNOWN';
    switch (handleType) {
      case 'EXEC':
        for (const command of hook.params.exec.command) {
          this.editHookCommandList.push(new FormGroup({
            arg: new FormControl(command)
          })
          )
        }
        this.editAddLiftcycleHookForm.patchValue({
          handlerType: 'EXEC'
        })
        break;
      case 'HTTPGET':
        for (const header of hook.params.httpGet.httpHeaders) {
          this.editHookHeadersList.push(new FormGroup({
            name: new FormControl(header.name),
            value: new FormControl(header.value)
          })
          )
        }
        this.editAddLiftcycleHookForm.patchValue({
          handlerType: 'HTTPGET',
          https: hook.params.httpGet.scheme === 'https',
          path: hook.params.httpGet.path,
          port: hook.params.httpGet.port
        })
        break;
      case 'TCPSOCKET':
        this.editAddLiftcycleHookForm.patchValue({
          handlerType: 'TCPSOCKET',
          tcpPort: hook.params.tcpSocket.port
        })
        break;
    }
    this.showAddLifecycleHookDetails = true;
  }

  deleteEditHookHeaders (index: number): void {
    this.editHookHeadersList.removeAt(index);
  }

  addNewEditHookHeader (): void {
    this.editHookHeadersList.push(new FormGroup({
      name: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }))
  }

  addNewEditHookCommand (): void {
    this.editHookCommandList.push(new FormGroup({
      arg: new FormControl('', Validators.required)
    }))
  }

  deleteEditHookCommands (index: number): void {
    this.editHookCommandList.removeAt(index);
  }

  doAddEditLifecycleHookCancel (): void {
    this.doAddEditLifecycleHookReset();
  }

  doAddEditLifecycleHookReset (): void {
    this.showAddLifecycleHookDetails = false;
    this.addLifecycleHook = '';
    this.editAddLiftcycleHookForm.reset();
    this.editAddLiftcycleHookForm.patchValue({
      handlerType: 'HTTPGET',
      https: false,
      path: '/',
      port: 8080,
      tcpPort: 8080
    })
    while (this.editHookHeadersList.length !== 0) {
      this.editHookHeadersList.removeAt(0);
    }
    while (this.editHookCommandList.length !== 0) {
      this.editHookCommandList.removeAt(0);
    }
  }

  doAddEditLifecycleHook (): void {
    const headerList: InputPodSpecContainerHandlerHTTPGetActionHTTPHeader[] = [];
    for (const each of this.editAddLiftcycleHookForm.value.headers) {
      headerList.push({
        name: each.name,
        value: each.value
      })
    }
    const commandList: string[] = [];
    for (const each of this.editAddLiftcycleHookForm.value.commands) {
      commandList.push(each.arg)
    }
    this.allLifecycleHook[this.addLifecycleHookEditIndex] = {
      hookType: this.addLifecycleHook,
      params: {
        exec: this.editAddLiftcycleHookForm.value.handlerType === 'EXEC'
          ? {
              command: commandList
            }
          : null,
        httpGet: this.editAddLiftcycleHookForm.value.handlerType === 'HTTPGET'
          ? {
              path: this.editAddLiftcycleHookForm.value.path ? this.editAddLiftcycleHookForm.value.path : '/',
              port: this.editAddLiftcycleHookForm.value.port,
              scheme: this.editAddLiftcycleHookForm.value.https ? 'https' : 'http',
              httpHeaders: headerList
            }
          : null,
        tcpSocket: this.editAddLiftcycleHookForm.value.handlerType === 'TCPSOCKET'
          ? {
              port: this.editAddLiftcycleHookForm.value.tcpPort
            }
          : null
      }
    }
    this.doAddEditLifecycleHookReset();
  }

  deleteContainer (index: number): void {
    this.allContainers.splice(index, 1)
  }

  doAddNewContainer (): void {
    const container: AddPodSpecContainer = {
      name: this.AddContainerForm.value.name,
      image: this.AddContainerForm.value.image
    }
    if (this.AddContainerForm.value.ports.length > 0) {
      const portList: AddPodSpecContainerPort[] = [];
      for (const each of this.AddContainerForm.value.ports) {
        portList.push({
          name: each.portName,
          protocol: each.protocol,
          containerPort: each.containerPort
        })
      }
      container.ports = portList;
    }
    if (this.AddContainerForm.value.advancedParams) {
      container.imagePullPolicy = this.AddContainerForm.value.imagePullPolicy ? this.AddContainerForm.value.imagePullPolicy : 'IfNotPresent';
      if (this.AddContainerForm.value.commands.length > 0) {
        const commandList: string[] = [];
        for (const each of this.AddContainerForm.value.commands) {
          commandList.push(each.command)
        }
        container.command = commandList;
      }
      if (this.AddContainerForm.value.args.length > 0) {
        const argsList: string[] = [];
        for (const each of this.AddContainerForm.value.args) {
          argsList.push(each.args)
        }
        container.args = argsList;
      }
      if (this.AddContainerForm.value.workingDir) {
        container.workingDir = this.AddContainerForm.value.workingDir;
      }
    }
    if (this.AddContainerForm.value.enableEnv && this.AddContainerForm.value.envs.length > 0) {
      const envsMap: JSON = JSON.parse(JSON.stringify({}));
      for (const each of this.AddContainerForm.value.envs) {
        envsMap[each.key] = each.value;
      }
      container.env = envsMap;
    }
    if (this.AddContainerForm.value.mounts.length > 0) {
      const mountList: AddPodSpecContainerVolumeMount[] = [];
      for (const each of this.AddContainerForm.value.mounts) {
        mountList.push({
          name: each.name,
          mountPath: each.mountPath,
          subPath: each.subPath
        })
      }
      container.volumeMounts = mountList;
    }
    if (this.AddContainerForm.value.enableHealthCheck && this.allHealthChecks.length > 0) {
      for (const healthcheck of this.allHealthChecks) {
        switch (healthcheck.checkType) {
          case 'Readiness':
            container.readinessProbe = healthcheck.params
            break;
          case 'Liveness':
            container.livenessProbe = healthcheck.params
            break;
          case 'Startup':
            container.startupProbe = healthcheck.params
            break;
        }
      }
    }
    if (this.AddContainerForm.value.enableLifecycleHook && this.allLifecycleHook.length > 0) {
      const lifecycleHooks: InputPodSpecContainerLifecycle = {
        postStart: null,
        preStop: null
      }
      for (const hook of this.allLifecycleHook) {
        switch (hook.hookType) {
          case 'postStart':
            lifecycleHooks.postStart = hook.params
            break;
          case 'preStop':
            lifecycleHooks.preStop = hook.params
            break;
        }
      }
      container.lifecycle = lifecycleHooks;
    }
    if (this.AddContainerForm.value.enableResource) {
      const resourceRequest: AddPodSpecContainerResourcesRequests = {
        cpu: this.AddContainerForm.value.cpuRequest ? this.AddContainerForm.value.cpuRequest + 'm' : null,
        memory: this.AddContainerForm.value.memRequest ? this.AddContainerForm.value.memRequest + 'Mi' : null
      };
      const resourceLimit: AddPodSpecContainerResourcesLimits = {
        cpu: this.AddContainerForm.value.cpuLimit ? this.AddContainerForm.value.cpuLimit + 'm' : null,
        memory: this.AddContainerForm.value.memLimit ? this.AddContainerForm.value.memLimit + 'Mi' : null
      };
      container.resources = {
        limits: resourceLimit,
        requests: resourceRequest
      };
    }
    this.allContainers.push(container);
    this.doAddNewContainerReset();
  }

  doAddNewContainerCancel (): void {
    this.doAddNewContainerReset();
  }

  doAddNewContainerReset (): void {
    this.AddContainerForm.reset();
    this.AddContainerForm.patchValue({
      advancedParams: false,
      enableResource: false,
      enableEnv: false,
      enableHealthCheck: false,
      enableLifecycleHook: false,
      imagePullPolicy: 'IfNotPresent'
    });
    while (this.envs.length !== 0) {
      this.envs.removeAt(0);
    }
    while (this.ports.length !== 0) {
      this.ports.removeAt(0);
    }
    while (this.mounts.length !== 0) {
      this.mounts.removeAt(0);
    }
    while (this.containerCommandList.length !== 0) {
      this.containerCommandList.removeAt(0);
    }
    while (this.containerArgsList.length !== 0) {
      this.containerArgsList.removeAt(0);
    }
    this.allHealthChecks = [];
    this.allLifecycleHook = [];
    this.showAddContaienr = false;
    this.portOverlap = false;
    this.addInitContainer = false;
  }

  addNewInitContainer (): void {
    this.showAddContaienr = true;
    this.addInitContainer = true;
  }

  deleteInitContainer (index: number): void {
    this.allInitContainers.splice(index, 1)
  }

  doAddNewInitContainer (): void {
    const container: AddPodSpecContainer = {
      name: this.AddContainerForm.value.name,
      image: this.AddContainerForm.value.image
    }
    if (this.AddContainerForm.value.advancedParams) {
      container.imagePullPolicy = this.AddContainerForm.value.imagePullPolicy ? this.AddContainerForm.value.imagePullPolicy : 'IfNotPresent';
      if (this.AddContainerForm.value.commands.length > 0) {
        const commandList: string[] = [];
        for (const each of this.AddContainerForm.value.commands) {
          commandList.push(each.command)
        }
        container.command = commandList;
      }
      if (this.AddContainerForm.value.args.length > 0) {
        const argsList: string[] = [];
        for (const each of this.AddContainerForm.value.args) {
          argsList.push(each.args)
        }
        container.args = argsList;
      }
      if (this.AddContainerForm.value.workingDir) {
        container.workingDir = this.AddContainerForm.value.workingDir;
      }
    }
    if (this.AddContainerForm.value.enableEnv && this.AddContainerForm.value.envs.length > 0) {
      const envsMap: JSON = JSON.parse(JSON.stringify({}));
      for (const each of this.AddContainerForm.value.envs) {
        envsMap[each.key] = each.value;
      }
      container.env = envsMap;
    }
    if (this.AddContainerForm.value.mounts.length > 0) {
      const mountList: AddPodSpecContainerVolumeMount[] = [];
      for (const each of this.AddContainerForm.value.mounts) {
        mountList.push({
          name: each.name,
          mountPath: each.mountPath,
          subPath: each.subPath
        })
      }
      container.volumeMounts = mountList;
    }
    if (this.AddContainerForm.value.enableResource) {
      const resourceRequest: AddPodSpecContainerResourcesRequests = {
        cpu: this.AddContainerForm.value.cpuRequest ? this.AddContainerForm.value.cpuRequest + 'm' : null,
        memory: this.AddContainerForm.value.memRequest ? this.AddContainerForm.value.memRequest + 'Mi' : null
      };
      const resourceLimit: AddPodSpecContainerResourcesLimits = {
        cpu: this.AddContainerForm.value.cpuLimit ? this.AddContainerForm.value.cpuLimit + 'm' : null,
        memory: this.AddContainerForm.value.memLimit ? this.AddContainerForm.value.memLimit + 'Mi' : null
      };
      container.resources = {
        limits: resourceLimit,
        requests: resourceRequest
      };
    }
    this.allInitContainers.push(container);
    this.doAddNewContainerReset();
  }

  showContainerDetails (container: AddPodSpecContainer, index: number): void {
    this.addContainerEditIndex = index;
    this.editAddContainerForm.patchValue({
      name: container.name,
      image: container.image
    });
    if (container.ports) {
      for (const each of container.ports) {
        this.editAddPorts.push(new FormGroup({
          portName: new FormControl(each.name),
          protocol: new FormControl(each.protocol),
          containerPort: new FormControl(each.containerPort)
        }))
      }
    }
    if ((!!container.imagePullPolicy && container.imagePullPolicy !== 'IfNotPresent') || !!container.command || !!container.args || !!container.workingDir) {
      this.editAddContainerForm.patchValue({
        advancedParams: true
      })
      if (container.command) {
        for (const command of container.command) {
          this.editContainerCommandList.push(new FormGroup({
            command: new FormControl(command)
          }))
        }
      }
      if (container.args) {
        for (const arg of container.args) {
          this.editContainerArgsList.push(new FormGroup({
            args: new FormControl(arg)
          }))
        }
      }
      if (container.workingDir) {
        this.editAddContainerForm.patchValue({
          workingDir: container.workingDir
        })
      }
      this.editAddContainerForm.patchValue({
        imagePullPolicy: container.imagePullPolicy
      })
    }
    if (container.resources) {
      if (!!container.resources.requests.cpu || !!container.resources.requests.memory || !!container.resources.limits.cpu || !!container.resources.limits.memory) {
        this.editAddContainerForm.patchValue({
          enableResource: true
        })
      }
      if (container.resources.requests) {
        if (container.resources.requests.cpu) {
          this.editAddContainerForm.patchValue({
            cpuRequest: container.resources.requests.cpu.split('m', 2)[0]
          })
        }
        if (container.resources.requests.memory) {
          this.editAddContainerForm.patchValue({
            memRequest: container.resources.requests.memory.split('Mi', 2)[0]
          })
        }
      }
      if (container.resources.limits) {
        if (container.resources.limits.cpu) {
          this.editAddContainerForm.patchValue({
            cpuLimit: container.resources.limits.cpu.split('m', 2)[0]
          })
        }
        if (container.resources.limits.memory) {
          this.editAddContainerForm.patchValue({
            memLimit: container.resources.limits.memory.split('Mi', 2)[0]
          })
        }
      }
    }
    if (container.env) {
      this.editAddContainerForm.patchValue({
        enableEnv: true
      })
      const envs = container.env;
      for (const key of Object.keys(envs)) {
        this.editAddEnvs.push(new FormGroup({
          key: new FormControl(key),
          value: new FormControl(envs[key])
        }));
      }
    }
    if (container.volumeMounts) {
      const mounts = container.volumeMounts;
      for (const each of mounts) {
        this.editAddMounts.push(new FormGroup({
          name: new FormControl(each.name),
          mountPath: new FormControl(each.mountPath),
          subPath: new FormControl(each.subPath)
        }));
      }
    }
    if (!!container.readinessProbe || !!container.livenessProbe || !!container.startupProbe) {
      this.editAddContainerForm.patchValue({
        enableHealthCheck: true
      })
      if (container.readinessProbe) {
        this.allHealthChecks.push({
          checkType: 'Readiness',
          params: {
            exec: container.readinessProbe.exec,
            httpGet: container.readinessProbe.httpGet,
            tcpSocket: container.readinessProbe.tcpSocket,
            initialDelaySeconds: container.readinessProbe.initialDelaySeconds,
            timeoutSeconds: container.readinessProbe.timeoutSeconds,
            periodSeconds: container.readinessProbe.periodSeconds,
            successThreshold: container.readinessProbe.successThreshold,
            failureThreshold: container.readinessProbe.failureThreshold
          }
        })
      }
      if (container.livenessProbe) {
        this.allHealthChecks.push({
          checkType: 'Liveness',
          params: {
            exec: container.livenessProbe.exec,
            httpGet: container.livenessProbe.httpGet,
            tcpSocket: container.livenessProbe.tcpSocket,
            initialDelaySeconds: container.livenessProbe.initialDelaySeconds,
            timeoutSeconds: container.livenessProbe.timeoutSeconds,
            periodSeconds: container.livenessProbe.periodSeconds,
            successThreshold: container.livenessProbe.successThreshold,
            failureThreshold: container.livenessProbe.failureThreshold
          }
        })
      }
      if (container.startupProbe) {
        this.allHealthChecks.push({
          checkType: 'Startup',
          params: {
            exec: container.startupProbe.exec,
            httpGet: container.startupProbe.httpGet,
            tcpSocket: container.startupProbe.tcpSocket,
            initialDelaySeconds: container.startupProbe.initialDelaySeconds,
            timeoutSeconds: container.startupProbe.timeoutSeconds,
            periodSeconds: container.startupProbe.periodSeconds,
            successThreshold: container.startupProbe.successThreshold,
            failureThreshold: container.startupProbe.failureThreshold
          }
        })
      }
    }
    if (container.lifecycle) {
      this.editAddContainerForm.patchValue({
        enableLifecycleHook: true
      })
      if (container.lifecycle.postStart) {
        this.allLifecycleHook.push({
          hookType: 'postStart',
          params: {
            exec: container.lifecycle.postStart.exec,
            httpGet: container.lifecycle.postStart.httpGet,
            tcpSocket: container.lifecycle.postStart.tcpSocket
          }
        })
      }
      if (container.lifecycle.preStop) {
        this.allLifecycleHook.push({
          hookType: 'preStop',
          params: {
            exec: container.lifecycle.preStop.exec,
            httpGet: container.lifecycle.preStop.httpGet,
            tcpSocket: container.lifecycle.preStop.tcpSocket
          }
        })
      }
    }
    this.showAddContaienrDetails = true;
  }

  doEditNewContainer (): void {
    const container: AddPodSpecContainer = {
      name: this.editAddContainerForm.value.name,
      image: this.editAddContainerForm.value.image
    }
    if (this.editAddContainerForm.value.ports.length > 0) {
      const portList: AddPodSpecContainerPort[] = [];
      for (const each of this.editAddContainerForm.value.ports) {
        portList.push({
          name: each.portName,
          protocol: each.protocol,
          containerPort: each.containerPort
        })
      }
      container.ports = portList;
    }
    if (this.editAddContainerForm.value.advancedParams) {
      container.imagePullPolicy = this.editAddContainerForm.value.imagePullPolicy ? this.editAddContainerForm.value.imagePullPolicy : 'IfNotPresent';
      if (this.editAddContainerForm.value.commands.length > 0) {
        const commandList: string[] = [];
        for (const each of this.editAddContainerForm.value.commands) {
          commandList.push(each.command)
        }
        container.command = commandList;
      }
      if (this.editAddContainerForm.value.args.length > 0) {
        const argsList: string[] = [];
        for (const each of this.editAddContainerForm.value.args) {
          argsList.push(each.args)
        }
        container.args = argsList;
      }
      if (this.editAddContainerForm.value.workingDir) {
        container.workingDir = this.editAddContainerForm.value.workingDir;
      }
    }
    if (this.editAddContainerForm.value.enableEnv && this.editAddContainerForm.value.envs.length > 0) {
      const envsMap: JSON = JSON.parse(JSON.stringify({}));
      for (const each of this.editAddContainerForm.value.envs) {
        envsMap[each.key] = each.value;
      }
      container.env = envsMap;
    }
    if (this.editAddContainerForm.value.mounts.length > 0) {
      const mountList: AddPodSpecContainerVolumeMount[] = [];
      for (const each of this.editAddContainerForm.value.mounts) {
        mountList.push({
          name: each.name,
          mountPath: each.mountPath,
          subPath: each.subPath
        })
      }
      container.volumeMounts = mountList;
    }
    if (this.editAddContainerForm.value.enableResource) {
      const resourceRequest: AddPodSpecContainerResourcesRequests = {
        cpu: this.editAddContainerForm.value.cpuRequest ? this.editAddContainerForm.value.cpuRequest + 'm' : null,
        memory: this.editAddContainerForm.value.memRequest ? this.editAddContainerForm.value.memRequest + 'Mi' : null
      };
      const resourceLimit: AddPodSpecContainerResourcesLimits = {
        cpu: this.editAddContainerForm.value.cpuLimit ? this.editAddContainerForm.value.cpuLimit + 'm' : null,
        memory: this.editAddContainerForm.value.memLimit ? this.editAddContainerForm.value.memLimit + 'Mi' : null
      };
      container.resources = {
        limits: resourceLimit,
        requests: resourceRequest
      };
    }
    if (this.editAddContainerForm.value.enableHealthCheck && this.allHealthChecks.length > 0) {
      for (const healthcheck of this.allHealthChecks) {
        switch (healthcheck.checkType) {
          case 'Readiness':
            container.readinessProbe = healthcheck.params
            break;
          case 'Liveness':
            container.livenessProbe = healthcheck.params
            break;
          case 'Startup':
            container.startupProbe = healthcheck.params
            break;
        }
      }
    }
    if (this.editAddContainerForm.value.enableLifecycleHook && this.allLifecycleHook.length > 0) {
      const lifecycleHooks: InputPodSpecContainerLifecycle = {
        postStart: null,
        preStop: null
      }
      for (const hook of this.allLifecycleHook) {
        switch (hook.hookType) {
          case 'postStart':
            lifecycleHooks.postStart = hook.params
            break;
          case 'preStop':
            lifecycleHooks.preStop = hook.params
            break;
        }
      }
      container.lifecycle = lifecycleHooks;
    }
    this.allContainers[this.addContainerEditIndex] = container;
    this.doEditNewContainerReset();
  }

  doEditNewContainerCancel (): void {
    this.doEditNewContainerReset();
  }

  doEditNewContainerReset (): void {
    this.editAddContainerForm.reset();
    this.editAddContainerForm.patchValue({
      advancedParams: false,
      enableResource: false,
      enableEnv: false,
      enableHealthCheck: false,
      enableLifecycleHook: false,
      imagePullPolicy: 'IfNotPresent'
    });
    while (this.editAddEnvs.length !== 0) {
      this.editAddEnvs.removeAt(0);
    }
    while (this.editAddPorts.length !== 0) {
      this.editAddPorts.removeAt(0);
    }
    while (this.editAddMounts.length !== 0) {
      this.editAddMounts.removeAt(0);
    }
    while (this.editContainerCommandList.length !== 0) {
      this.editContainerCommandList.removeAt(0);
    }
    while (this.editContainerArgsList.length !== 0) {
      this.editContainerArgsList.removeAt(0);
    }
    this.showAddContaienrDetails = false;
    this.addContainerEditIndex = null;
    this.portOverlap = false;
    this.allHealthChecks = [];
    this.allLifecycleHook = [];
    this.addInitContainer = false;
  }

  showInitContainerDetails (container: AddPodSpecContainer, index: number): void {
    this.addContainerEditIndex = index;
    this.editAddContainerForm.patchValue({
      name: container.name,
      image: container.image
    });
    if ((!!container.imagePullPolicy && container.imagePullPolicy !== 'IfNotPresent') || !!container.command || !!container.args || !!container.workingDir) {
      this.editAddContainerForm.patchValue({
        advancedParams: true
      })
      if (container.command) {
        for (const command of container.command) {
          this.editContainerCommandList.push(new FormGroup({
            command: new FormControl(command)
          }))
        }
      }
      if (container.args) {
        for (const arg of container.args) {
          this.editContainerArgsList.push(new FormGroup({
            args: new FormControl(arg)
          }))
        }
      }
      if (container.workingDir) {
        this.editAddContainerForm.patchValue({
          workingDir: container.workingDir
        })
      }
      this.editAddContainerForm.patchValue({
        imagePullPolicy: container.imagePullPolicy
      })
    }
    if (container.resources) {
      if (!!container.resources.requests.cpu || !!container.resources.requests.memory || !!container.resources.limits.cpu || !!container.resources.limits.memory) {
        this.editAddContainerForm.patchValue({
          enableResource: true
        })
      }
      if (container.resources.requests) {
        if (container.resources.requests.cpu) {
          this.editAddContainerForm.patchValue({
            cpuRequest: container.resources.requests.cpu.split('m', 2)[0]
          })
        }
        if (container.resources.requests.memory) {
          this.editAddContainerForm.patchValue({
            memRequest: container.resources.requests.memory.split('Mi', 2)[0]
          })
        }
      }
      if (container.resources.limits) {
        if (container.resources.limits.cpu) {
          this.editAddContainerForm.patchValue({
            cpuLimit: container.resources.limits.cpu.split('m', 2)[0]
          })
        }
        if (container.resources.limits.memory) {
          this.editAddContainerForm.patchValue({
            memLimit: container.resources.limits.memory.split('Mi', 2)[0]
          })
        }
      }
    }
    if (container.env) {
      this.editAddContainerForm.patchValue({
        enableEnv: true
      })
      const envs = container.env;
      for (const key of Object.keys(envs)) {
        this.editAddEnvs.push(new FormGroup({
          key: new FormControl(key),
          value: new FormControl(envs[key])
        }));
      }
    }
    if (container.volumeMounts) {
      const mounts = container.volumeMounts;
      for (const each of mounts) {
        this.editAddMounts.push(new FormGroup({
          name: new FormControl(each.name),
          mountPath: new FormControl(each.mountPath),
          subPath: new FormControl(each.subPath)
        }));
      }
    }
    this.showAddContaienrDetails = true;
    this.addInitContainer = true;
  }

  doEditNewInitContainer (): void {
    const container: AddPodSpecContainer = {
      name: this.editAddContainerForm.value.name,
      image: this.editAddContainerForm.value.image
    }
    if (this.editAddContainerForm.value.advancedParams) {
      container.imagePullPolicy = this.editAddContainerForm.value.imagePullPolicy ? this.editAddContainerForm.value.imagePullPolicy : 'IfNotPresent';
      if (this.editAddContainerForm.value.commands.length > 0) {
        const commandList: string[] = [];
        for (const each of this.editAddContainerForm.value.commands) {
          commandList.push(each.command)
        }
        container.command = commandList;
      }
      if (this.editAddContainerForm.value.args.length > 0) {
        const argsList: string[] = [];
        for (const each of this.editAddContainerForm.value.args) {
          argsList.push(each.args)
        }
        container.args = argsList;
      }
      if (this.editAddContainerForm.value.workingDir) {
        container.workingDir = this.editAddContainerForm.value.workingDir;
      }
    }
    if (this.editAddContainerForm.value.enableEnv && this.editAddContainerForm.value.envs.length > 0) {
      const envsMap: JSON = JSON.parse(JSON.stringify({}));
      for (const each of this.editAddContainerForm.value.envs) {
        envsMap[each.key] = each.value;
      }
      container.env = envsMap;
    }
    if (this.editAddContainerForm.value.mounts.length > 0) {
      const mountList: AddPodSpecContainerVolumeMount[] = [];
      for (const each of this.editAddContainerForm.value.mounts) {
        mountList.push({
          name: each.name,
          mountPath: each.mountPath,
          subPath: each.subPath
        })
      }
      container.volumeMounts = mountList;
    }
    if (this.editAddContainerForm.value.enableResource) {
      const resourceRequest: AddPodSpecContainerResourcesRequests = {
        cpu: this.editAddContainerForm.value.cpuRequest ? this.editAddContainerForm.value.cpuRequest + 'm' : null,
        memory: this.editAddContainerForm.value.memRequest ? this.editAddContainerForm.value.memRequest + 'Mi' : null
      };
      const resourceLimit: AddPodSpecContainerResourcesLimits = {
        cpu: this.editAddContainerForm.value.cpuLimit ? this.editAddContainerForm.value.cpuLimit + 'm' : null,
        memory: this.editAddContainerForm.value.memLimit ? this.editAddContainerForm.value.memLimit + 'Mi' : null
      };
      container.resources = {
        limits: resourceLimit,
        requests: resourceRequest
      };
    }
    this.allInitContainers[this.addContainerEditIndex] = container;
    this.doEditNewContainerReset();
  }

  addEditPorts (): void {
    this.editAddPorts.push(new FormGroup({
      portName: new FormControl('', Validators.required),
      protocol: new FormControl('TCP', Validators.required),
      containerPort: new FormControl(8080, Validators.required)
    }));
    this.checkCreatePortOverlap(8080, 'TCP');
  }

  deleteEditPort (index: number): void {
    this.editAddPorts.removeAt(index);
  }

  addEditEnv (): void {
    this.editAddEnvs.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  deleteEditEnv (index: number): void {
    this.editAddEnvs.removeAt(index);
  }

  addEditMounts (): void {
    this.editAddMounts.push(new FormGroup({
      name: new FormControl('', Validators.required),
      mountPath: new FormControl('', Validators.required),
      subPath: new FormControl('')
    }));
  }

  deleteEditMount (index: number): void {
    this.editAddMounts.removeAt(index);
  }

  addNewEditContainerCommand (): void {
    this.editContainerCommandList.push(new FormGroup({
      command: new FormControl('', Validators.required)
    }))
  }

  deleteEditContainerCommands (index: number): void {
    this.editContainerCommandList.removeAt(index);
  }

  addNewEditContainerArgs (): void {
    this.editContainerArgsList.push(new FormGroup({
      args: new FormControl('', Validators.required)
    }))
  }

  deleteEditContainerArgs (index: number): void {
    this.editContainerArgsList.removeAt(index);
  }

  getDockerConfigSecret (): void {
    this.dockerConfigSecretList = [{name: 'vmware-system-map-registry'}];
    for (const storage of this.allStorages) {
      if (storage.storageType === 'secret' && (storage.secretType === 'kubernetes.io/dockerconfigjson' || storage.secretType === 'kubernetes.io/dockercfg')) {
        this.dockerConfigSecretList.push({
          name: storage.name
        })
      }
    }
  }

  updateOtherResourceForm(): void {
    this.addOtherResourceForm.updateValueAndValidity();
  }

  updateEditOtherResourceForm(): void {
    this.editOtherResourceForm.updateValueAndValidity();
  }

  addNewWorkloadLabel (): void {
    this.labelsList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
    this.updateOtherResourceForm();
    this.updateEditOtherResourceForm();
  }

  deleteWorkloadLable (index: number): void {
    this.labelsList.removeAt(index);
    this.updateOtherResourceForm();
    this.updateEditOtherResourceForm();
  }

  // add service in component
  deleteService (serviceIndex: number): void {
    this.serviceList.removeAt(serviceIndex);
    this.allPortList.splice(serviceIndex, 1);
  }

  addService (name?: string): void {
    this.allPortList[this.serviceList.length] = new FormArray([]);
    this.serviceList.push(this.fb.group({
      name: new FormControl(name || '', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern)]),
      serviceType: new FormControl('ClusterIP', Validators.required),
      ports: this.allPortList[this.serviceList.length]
    }));
    this.addServiceNewPort(this.serviceList.length - 1);
  }

  addServiceNewPort (serviceIndex: number): void {
    this.allPortList[serviceIndex].push(this.fb.group({
      servicePort: new FormControl(80, Validators.required),
      protocol: new FormControl('TCP', Validators.required),
      workloadPortName: new FormControl('', Validators.required)
    }));
  }

  deleteServicePort (serviceIndex: number, portIndex: number): void {
    this.allPortList[serviceIndex].removeAt(portIndex);
  }

  doWorkloadPageCustomNext (buttonType: string): void {
    this.workloadPortsList = null;
    if (buttonType === 'custom-next') {
      const workloadType = this.workloadParamsForm.value.workloadType;
      switch (workloadType) {
        case 'Deployment': {
          const tempPortList = [];
          for (const container of this.allContainers) {
            if (container.ports) {
              for (const port of container.ports) {
                tempPortList.push(port)
              }
            }
          }
          this.workloadPortsList = tempPortList;
          break;
        }
        default:
          this.workloadPortsList = [];
          break;
      }
    }
    // add same name service first
    if (!!this.workloadPortsList && this.workloadPortsList.length > 0 && this.serviceList.length <= 0) {
      this.addService(this.basicComponentForm.value.name);
    }
    this.wizardnewcomponent.next();
  }

  doEditWorkloadPageCustomNext (buttonType: string): void {
    this.workloadPortsList = null;
    if (buttonType === 'custom-next') {
      const workloadType = this.editWorkloadParamsForm.value.workloadType;
      switch (workloadType) {
        case 'Deployment': {
          const tempPortList = [];
          for (const container of this.allContainers) {
            if (container.ports) {
              for (const port of container.ports) {
                tempPortList.push(port)
              }
            }
          }
          this.workloadPortsList = tempPortList;
          break;
        }
        default:
          this.workloadPortsList = [];
          break;
      }
    }
    this.wizardeditcomponent.next();
  }

  get checkServiceInvalid (): boolean {
    if (this.serviceList.length > 0) {
      for (const control of this.serviceList.controls) {
        if (control.invalid) {
          return true
        }
      }
    }
    // no service or no invalid
    return false;
  }

  doServicePageCustomPrevious (): void {
    // reset dom
    this.workloadPortsList = null;
    for (let i = 0; i < this.serviceList.length; i++) {
      for (const control of this.allPortList[i].controls) {
        control.controls.workloadPortName.setValue('');
        control.controls.workloadPortName.updateValueAndValidity();
      }
    }
  }

  // add other resource
  checkHpaReplicasVaild (update?: boolean): void {
    this.hpaReplicasInvaild = false;
    if (update) {
      this.hpaReplicasInvaild = this.editOtherResourceForm.value.maxReplicas < this.editOtherResourceForm.value.minReplicas;
    } else {
      this.hpaReplicasInvaild = this.addOtherResourceForm.value.maxReplicas < this.addOtherResourceForm.value.minReplicas;
    }
  }

  resetHpaForm (enable: boolean, update?: boolean): void {
    if (enable) {
      this.checkHpaReplicasVaild(update);
    } else {
      this.hpaReplicasInvaild = false;
      if(update){
        this.editOtherResourceForm.patchValue({
          enableHpa: false,
          hpaMetrics: 'CPU',
          percentage: 80,
          minReplicas: 1,
          maxReplicas: 10
        })
        this.updateEditOtherResourceForm();
      } else {
        this.addOtherResourceForm.patchValue({
          enableHpa: false,
          hpaMetrics: 'CPU',
          percentage: 80,
          minReplicas: 1,
          maxReplicas: 10
        })
        this.updateOtherResourceForm();
      }
    }
  }

  // add component
  showCreateComponent (): void {
    if (this.selectedApplication !== 'ALL') {
      this.basicComponentForm.patchValue({
        application: this.selectedApplication
      })
      this.basicComponentForm.updateValueAndValidity();
    }
    this.allComponentApplications = deepCopy(this.allApplications);
    this.showNoApplicationAlert = this.allComponentApplications.length <= 0;
    this.createComponentOpen = true;
  }

  doCreateComponentFinish (): void {
    const type = this.basicComponentForm.value.type;
    const name = this.basicComponentForm.value.name;
    const desc = this.basicComponentForm.value.description;
    switch (type) {
      case 'RAW': {
        const rawParam: CreateComponent = {
          name: name,
          description: desc,
          template: {
            raw: this.manifests
          },
          application: this.basicComponentForm.value.application
        }
        this.componentService.createComponent(rawParam).subscribe(data => {
          if (data.errors) {
            for (const each of data.errors) {
              if (each.message) {
                this.toast.show(this.translate.instant('MESSAGE.SHARED.CREATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
              }
            }
          }
          if (data.data?.CreateComponent) {
            this.toast.show(`${data.data.CreateComponent.name} ` + this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
          }
        }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
          this.doCreateRawComponentReset();
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
        });
        break;
      }
      case 'BASIC': {
        const secretStorageList: AddBasicComponentTemplateSecret[] = [];
        const configmapStorageList: AddBasicComponentTemplateConfigMap[] = [];
        const pvcStorageList: AddBasicComponentTemplatePersistentVolumeClaim[] = [];
        for (const storage of this.allStorages) {
          switch (storage.storageType) {
            case 'configMap':
              configmapStorageList.push({
                name: storage.name,
                data: storage.configMap.data
              })
              break;
            case 'secret':
              secretStorageList.push({
                name: storage.name,
                generic: {
                  type: storage.secretType,
                  data: storage.secret.data
                }
              })
              break;
            case 'persistentVolumeClaim':
              pvcStorageList.push({
                name: storage.name,
                spec: storage.persistentVolumeClaim.spec
              })
              break;
            default:
              console.log('not support')
          }
        }
        // add lables to workload and service
        const labelsMap: JSON = JSON.parse(JSON.stringify({}));
        if (this.addOtherResourceForm.value.labels) {
          for (const each of this.addOtherResourceForm.value.labels) {
            labelsMap[each.key] = each.value;
          }
        }
        const workloads: AddBasicComponentTemplateDeployment[] = [{
          name: name,
          labels: labelsMap,
          spec: {
            replicas: this.workloadParamsForm.value.replicas,
            template: {
              spec: {
                initContainers: (this.workloadParamsForm.value.advancedParams && this.allInitContainers.length > 0) ? this.allInitContainers : null,
                containers: this.allContainers,
                volumes: this.allVolumes,
                imagePullSecrets: this.workloadParamsForm.value.imagePullSecrets ? this.workloadParamsForm.value.imagePullSecrets : ['vmware-system-map-registry'],
                dnsPolicy: this.workloadParamsForm.value.dnsPolicy ? this.workloadParamsForm.value.dnsPolicy : 'ClusterFirst',
                restartPolicy: this.workloadParamsForm.value.restartPolicy ? this.workloadParamsForm.value.restartPolicy : 'Always',
                schedulerName: !this.workloadParamsForm.value.schedulerName ? this.workloadParamsForm.value.schedulerName : 'default-scheduler',
                serviceAccount: this.workloadParamsForm.value.serviceAccount ? this.workloadParamsForm.value.serviceAccount : 'default',
                terminationGracePeriodSeconds: this.workloadParamsForm.value.terminationGracePeriodSeconds ? this.workloadParamsForm.value.terminationGracePeriodSeconds : 30
              }
            },
            progressDeadlineSeconds: this.workloadParamsForm.value.progressDeadlineSeconds ? this.workloadParamsForm.value.progressDeadlineSeconds : 600,
            revisionHistoryLimit: this.workloadParamsForm.value.revisionHistoryLimit ? this.workloadParamsForm.value.revisionHistoryLimit : 10,
            strategy: {
              type: this.workloadParamsForm.value.strategyType,
              rollingUpdate: this.workloadParamsForm.value.strategyType === 'RollingUpdate'
                ? {
                    maxUnavailable: this.workloadParamsForm.value.maxUnavailable ? this.workloadParamsForm.value.maxUnavailable : '25%',
                    maxSurge: this.workloadParamsForm.value.maxSurge ? this.workloadParamsForm.value.maxSurge : '25%'
                  }
                : null
            }
          }
        }]
        // add hpa
        if (this.addOtherResourceForm.value.enableHpa) {
          const hpaParams: CreateHorizontalPodAutoscalerSpec = {
            maxReplicas: this.addOtherResourceForm.value.maxReplicas,
            minReplicas: this.addOtherResourceForm.value.minReplicas
          }
          if (this.addOtherResourceForm.value.hpaMetrics === 'CPU') {
            hpaParams.targetCPUUtilizationPercentage = this.addOtherResourceForm.value.percentage;
          } else {
            hpaParams.targetMemoryUtilizationPercentage = this.addOtherResourceForm.value.percentage;
          }
          workloads[0].hpa = hpaParams;
        }
        // add service
        const componentServiceList: AddBasicComponentTemplateService[] = [];
        if (this.addServiceForm.value.services) {
          for (let i = 0; i < this.serviceList.length; i++) {
            const service = this.serviceList.controls[i];
            const ports: AddServiceSpecPort[] = [];
            for (const port of this.allPortList[i].controls) {
              ports.push({
                name: port.value.workloadPortName + '-' + port.value.servicePort,
                port: port.value.servicePort,
                protocol: port.value.protocol,
                targetPort: port.value.workloadPortName
              })
            }
            componentServiceList.push({
              name: service.value.name,
              labels: labelsMap,
              spec: {
                type: service.value.serviceType,
                ports: ports
              }
            })
          }
        }
        const basicParam: CreateComponent = {
          name: name,
          description: desc,
          template: {
            basic: {
              services: componentServiceList,
              deployments: workloads,
              secrets: secretStorageList,
              configMaps: configmapStorageList,
              persistentVolumeClaims: pvcStorageList
            }
          },
          application: this.basicComponentForm.value.application
        }
        this.componentService.createComponent(basicParam).subscribe(data => {
          if (data.errors) {
            for (const each of data.errors) {
              if (each.message) {
                this.toast.show(this.translate.instant('MESSAGE.SHARED.CREATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
              }
            }
          }
          if (data.data?.CreateComponent) {
            this.toast.show(`${data.data.CreateComponent.name} ` + this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
          }
        }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
          this.doCreateBasicComponentReset();
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
        });
        break;
      }
      default:
        console.log('not support other type')
    }
  }

  doCreateComponentCancel (): void {
    this.doCreateRawComponentReset();
    this.doCreateBasicComponentReset();
  }

  doCreateRawComponentReset (): void {
    this.createComponentOpen = false;
    this.wizardnewcomponent.reset();
    this.basicComponentForm.reset();
    this.manifests = '';
    this.checkFormateOk = false;
    this.parseError = '';
    this.showCheckedAlert = false;
    this.allComponentApplications = null;
    this.showNoApplicationAlert = false;
  }

  doCreateBasicComponentReset (): void {
    this.createComponentOpen = false;
    this.wizardnewcomponent.reset();
    this.basicComponentForm.reset();
    this.allStorages = [];
    this.allVolumes = [];
    this.workloadParamsForm.reset();
    this.workloadParamsForm.patchValue({
      workloadType: 'Deployment',
      replicas: 1,
      advancedParams: false,
      imagePullSecrets: ['vmware-system-map-registry'],
      serviceAccount: 'default',
      restartPolicy: 'Always',
      dnsPolicy: 'ClusterFirst',
      schedulerName: 'default-scheduler',
      terminationGracePeriodSeconds: 30,
      progressDeadlineSeconds: 600,
      revisionHistoryLimit: 10,
      strategyType: 'RollingUpdate',
      maxUnavailable: '25%',
      maxSurge: '25%'
    })
    this.dockerConfigSecretList = null;
    this.allContainers = [];
    this.allInitContainers = [];
    this.addServiceForm.reset();
    while (this.serviceList.length !== 0) {
      this.serviceList.removeAt(0);
    }
    this.allPortList = [];
    this.workloadPortsList = null;
    this.addOtherResourceForm.reset();
    this.addOtherResourceForm.patchValue({
      enableHpa: false,
      hpaMetrics: 'CPU',
      percentage: 80,
      minReplicas: 1,
      maxReplicas: 10
    })
    this.hpaReplicasInvaild = false;
    while (this.labelsList.length !== 0) {
      this.labelsList.removeAt(0);
    }
    this.allComponentApplications = null;
    this.showNoApplicationAlert = false;
  }

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  initComponentNext ($event: any): void {
    const type = $event.target.value;
    switch (type) {
      case 'RAW':
        this.manifests = '';
        break;
      case 'BASIC':
        // add some component init
        break;
      default:
        console.log('not support.')
    }
  }

  modelValueChange (): void {
    this.showCheckedAlert = false;
    this.checkFormateOk = false;
  }

  checkFileFormat (): void {
    [this.checkFormateOk, this.parseError] = checkYamlFormat(this.manifests)
    this.showCheckedAlert = true;
  }

  showAddFileData (): void {
    this.showAddFile = true;
  }

  addFileToData (): void {
    const fileUpload = this.fileUpload.nativeElement;
    const file = fileUpload.files[0];
    if (file) {
      const reader = new FileReader();
      reader.readAsText(file, 'UTF-8');
      reader.onload = () => {
        this.manifests = reader.result + '';
      };
    }
    this.fileUpload.nativeElement.value = '';
    this.filename = null;
    this.showAddFile = false;
  }

  cancelAddFile (): void {
    this.fileUpload.nativeElement.value = null;
    this.filename = null;
    this.showAddFile = false;
  }

  getFileName (): void {
    const fileUpload = this.fileUpload.nativeElement;
    const file = fileUpload.files[0];
    this.filename = file ? file.name : '';
  }

  showEditFileData (): void {
    this.showEditAddFile = true;
  }

  addEditFileToData (): void {
    const fileUpload = this.editFileUpload.nativeElement;
    const file = fileUpload.files[0];
    if (file) {
      const reader = new FileReader();
      reader.readAsText(file, 'UTF-8');
      reader.onload = () => {
        this.manifests = reader.result + '';
      };
    }
    this.editFileUpload.nativeElement.value = '';
    this.filename = null;
    this.showEditAddFile = false;
  }

  cancelAddEditFile (): void {
    this.editFileUpload.nativeElement.value = null;
    this.filename = null;
    this.showEditAddFile = false;
  }

  getEditFileName (): void {
    const fileUpload = this.editFileUpload.nativeElement;
    const file = fileUpload.files[0];
    this.filename = file ? file.name : '';
  }

  // edit HELM component function
  checkValueFormat (): void {
    [this.checkFormateOk, this.parseError] = checkOneYamlFormat(this.newValues)
    this.showCheckedAlert = true;
  }

  resetValues (): void {
    this.newValues = this.originValues;
    this.showCheckedAlert = false;
    this.checkFormateOk = false;
    this.parseError = '';
  }

  getChartVersionList (): void {
    this.noChartVersionAlert = false;
    this.repositoryService.getAllChartRepoVersion(this.chartVersionForm.value.registry, this.chartVersionForm.value.project, this.chartVersionForm.value.name).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.CHARTS-DETAILS.GET-CHARTDETAILS-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data.ChartRepos[0].versions) {
        this.chartVersionListSorted = data.data.ChartRepos[0].versions.sort();
        this.noChartVersionAlert = !this.chartVersionListSorted.find(x => x.version === this.chartVersionForm.value.version)
      } else {
        this.noChartVersionAlert = true;
      }
    })
  }

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  getChartValues ($event: any): void {
    this.valueload = true;
    this.checkFormateOk = false;
    this.showCheckedAlert = false;
    const version = $event.target.value;
    this.repositoryService.getChartDetailValues(this.chartVersionForm.value.registry, this.chartVersionForm.value.project, this.chartVersionForm.value.name, version).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.CHARTS-DETAILS.GET-CHARTDETAILS-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data?.data?.ChartDetails) {
        this.chartVersionForm.patchValue({
          version: version,
          url: data.data.ChartDetails.download
        })
        this.newValues = data.data.ChartDetails.files['values.yaml'];
        this.originValues = data.data.ChartDetails.files['values.yaml'];
      } else {
        this.originValues = '';
        this.newValues = '';
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.CHARTS-DETAILS.GET-CHARTDETAILS-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => { this.valueload = false; });
  }

  // update component desc
  showUpdateComponentDesc (component: CustomComponent): void {
    this.selectedComponent = component;
    this.updateComponentDescForm.setValue({
      description: component.description
    });
    this.showUpdateComponentDescDialog = true;
  }

  doUpdateComponentDescription (): void {
    this.showUpdateComponentDescDialog = false;
    const input: UpdateComponent = {
      id: this.selectedComponent?.id,
      description: this.updateComponentDescForm.value.description
    }
    this.componentService.updateComponent(input).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.UPDATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.UpdateComponent) {
        this.toast.show(`${data.data.UpdateComponent.name} ` + this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.refresh();
    });
  }

  // update component
  showUpdateComponent (cpt: CustomComponent): void {
    this.selectedComponent = cpt;
    this.componentService.getComponentById(cpt.id).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.GET-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      const component: CustomComponent = data?.data?.Component ? data.data.Component : null;
      if (component) {
        this.updateComponentOpen = true;
        switch (component.type) {
          case 'RAW': {
            this.manifests = component.template.raw
            break;
          }
          case 'HELM': {
            const params = component.template.extraInfo;
            this.chartVersionForm.patchValue({
              registry: params.currentChartRegistry,
              project: params.currentChartProject,
              name: params.currentChartName,
              version: params.currentChartVersion,
              url: component.template.url
            })
            this.getChartVersionList();
            this.newValues = YAML.stringify(component.template.values);
            this.originValues = component.template.chart.files['values.yaml']
            break;
          }
          case 'BASIC': {
            this.removeDeployment.push({
              name: component.name
            })
            // handle all storage
            if (component.template.configMaps) {
              let configMap: BasicComponentTemplateConfigMap;
              for (configMap of component.template.configMaps) {
                this.allStorages.push({
                  name: configMap.name,
                  storageType: 'configMap',
                  configMap: {
                    data: configMap.data
                  }
                })
                this.removeConfigMapStorage.push({
                  name: configMap.name
                })
              }
            }
            if (component.template.persistentVolumeClaims) {
              let pvc: BasicComponentTemplatePersistentVolumeClaim;
              for (pvc of component.template.persistentVolumeClaims) {
                this.allStorages.push({
                  name: pvc.name,
                  storageType: 'persistentVolumeClaim',
                  persistentVolumeClaim: {
                    spec: {
                      storageClass: pvc.spec.storageClass ? pvc.spec.storageClass : '',
                      accessModes: pvc.spec.accessModes,
                      resources: {
                        requests: {
                          storage: pvc.spec.resources.requests.storage
                        },
                        limits: {
                          storage: pvc.spec.resources.limits.storage ? pvc.spec.resources.limits.storage : ''
                        }
                      },
                      volumeMode: pvc.spec.volumeMode
                    }
                  }
                })
                this.removePVCStprage.push({
                  name: pvc.name
                })
              }
            }
            if (component.template.secrets) {
              for (const secret of component.template.secrets) {
                this.allStorages.push({
                  name: secret.name,
                  storageType: 'secret',
                  secretType: secret.type,
                  secret: {
                    data: secret.data
                  }
                })
                this.removeSecretStorage.push({
                  name: secret.name
                })
              }
            }
            // handle all volumes of workload
            if (component.template.deployments[0].spec.template.spec.volumes) {
              for (const volume of component.template.deployments[0].spec.template.spec.volumes) {
                const volumeItemsList: TemplateSpecVolumeKeyToPathInput[] = [];
                if (volume.volume.__typename === 'ConfigMapVolume' || volume.volume.__typename === 'SecretVolume') {
                  if (volume.volume.items) {
                    for (const item of volume.volume.items) {
                      volumeItemsList.push({
                        key: item.key,
                        path: item.path,
                        mode: item.mode
                      })
                    }
                  }
                }
                this.allVolumes.push({
                  name: volume.name,
                  emptyDir: volume.volume.__typename === 'EmptyDir'
                    ? {
                        medium: !!volume.volume.medium ? volume.volume.medium : '',
                        size: volume.volume.size
                      }
                    : null,
                  configMap: volume.volume.__typename === 'ConfigMapVolume'
                    ? {
                        name: volume.volume.name,
                        defaultMode: volume.volume.defaultMode,
                        items: volumeItemsList
                      }
                    : null,
                  secret: volume.volume.__typename === 'SecretVolume'
                    ? {
                        name: volume.volume.name,
                        defaultMode: volume.volume.defaultMode,
                        items: volumeItemsList
                      }
                    : null,
                  persistentVolumeClaim: volume.volume.__typename === 'PersistentVolumeClaimVolume'
                    ? {
                        name: volume.volume.name,
                        readOnly: volume.volume.readOnly
                      }
                    : null
                })
              }
            }
            // handle all containers of workload
            for (const container of component.template.deployments[0].spec.template.spec.containers) {
              // handle resource of container
              let cpuRequest = null;
              let memRequest = null;
              let cpuLimit = null;
              let memLimit = null;
              if (container.resources) {
                if (container.resources.requests.cpu) {
                  const cpuRequestList = container.resources.requests.cpu.split('m', 2);
                  cpuRequest = cpuRequestList.length === 2 ? cpuRequestList[0] : Number(cpuRequestList[0]) * 1000;
                }
                if (container.resources.requests.memory) {
                  const memRequestList = container.resources.requests.memory.split('Mi', 2);
                  memRequest = memRequestList.length === 2 ? memRequestList[0] : Number(container.resources.requests.memory.split('Gi', 2)[0]) * 1024;
                }
                if (container.resources.limits.cpu) {
                  const cpuLimitList = container.resources.limits.cpu.split('m', 2);
                  cpuLimit = cpuLimitList.length === 2 ? cpuLimitList[0] : Number(cpuLimitList[0]) * 1000;
                }
                if (container.resources.limits.memory) {
                  const memLimitList = container.resources.requests.memory.split('Mi', 2);
                  memLimit = memLimitList.length === 2 ? memLimitList[0] : Number(container.resources.limits.memory.split('Gi', 2)[0]) * 1024;
                }
              }
              // handle ports of container
              const allPortsOfContainer: AddPodSpecContainerPort[] = [];
              if (container.ports) {
                for (const port of container.ports) {
                  allPortsOfContainer.push({
                    name: port.name,
                    protocol: port.protocol,
                    containerPort: port.containerPort
                  })
                }
              }
              // handle volumeMount of container
              const allVolumeMountOfContainer: AddPodSpecContainerVolumeMount[] = [];
              if (container.volumeMounts) {
                for (const volumeMount of container.volumeMounts) {
                  allVolumeMountOfContainer.push({
                    name: volumeMount.name,
                    mountPath: volumeMount.mountPath,
                    subPath: volumeMount.subPath
                  })
                }
              }
              this.allContainers.push({
                name: container.name,
                image: container.image,
                command: container.command ? container.command : null,
                args: container.args ? container.args : null,
                workingDir: container.workingDir ? container.workingDir : null,
                ports: allPortsOfContainer,
                env: container.env,
                volumeMounts: allVolumeMountOfContainer,
                resources: {
                  requests: {
                    cpu: cpuRequest ? cpuRequest + 'm' : null,
                    memory: memRequest ? memRequest + 'Mi' : null
                  },
                  limits: {
                    cpu: cpuLimit ? cpuLimit + 'm' : null,
                    memory: memLimit ? memLimit + 'Mi' : null
                  }
                },
                readinessProbe: container.readinessProbe
                  ? {
                      exec: container.readinessProbe.exec
                        ? {
                            command: container.readinessProbe.exec.command
                          }
                        : null,
                      httpGet: container.readinessProbe.httpGet
                        ? {
                            path: container.readinessProbe.httpGet.path,
                            port: container.readinessProbe.httpGet.port,
                            scheme: container.readinessProbe.httpGet.scheme,
                            httpHeaders: container.readinessProbe.httpGet.httpHeaders
                          }
                        : null,
                      tcpSocket: container.readinessProbe.tcpSocket
                        ? {
                            port: container.readinessProbe.tcpSocket.port
                          }
                        : null,
                      initialDelaySeconds: container.readinessProbe.initialDelaySeconds,
                      timeoutSeconds: container.readinessProbe.timeoutSeconds,
                      periodSeconds: container.readinessProbe.periodSeconds,
                      successThreshold: container.readinessProbe.successThreshold,
                      failureThreshold: container.readinessProbe.failureThreshold
                    }
                  : null,
                livenessProbe: container.livenessProbe
                  ? {
                      exec: container.livenessProbe.exec
                        ? {
                            command: container.livenessProbe.exec.command
                          }
                        : null,
                      httpGet: container.livenessProbe.httpGet
                        ? {
                            path: container.livenessProbe.httpGet.path,
                            port: container.livenessProbe.httpGet.port,
                            scheme: container.livenessProbe.httpGet.scheme,
                            httpHeaders: container.livenessProbe.httpGet.httpHeaders
                          }
                        : null,
                      tcpSocket: container.livenessProbe.tcpSocket
                        ? {
                            port: container.livenessProbe.tcpSocket.port
                          }
                        : null,
                      initialDelaySeconds: container.livenessProbe.initialDelaySeconds,
                      timeoutSeconds: container.livenessProbe.timeoutSeconds,
                      periodSeconds: container.livenessProbe.periodSeconds,
                      successThreshold: container.livenessProbe.successThreshold,
                      failureThreshold: container.livenessProbe.failureThreshold
                    }
                  : null,
                startupProbe: container.startupProbe
                  ? {
                      exec: container.startupProbe.exec
                        ? {
                            command: container.startupProbe.exec.command
                          }
                        : null,
                      httpGet: container.startupProbe.httpGet
                        ? {
                            path: container.startupProbe.httpGet.path,
                            port: container.startupProbe.httpGet.port,
                            scheme: container.startupProbe.httpGet.scheme,
                            httpHeaders: container.startupProbe.httpGet.httpHeaders
                          }
                        : null,
                      tcpSocket: container.startupProbe.tcpSocket
                        ? {
                            port: container.startupProbe.tcpSocket.port
                          }
                        : null,
                      initialDelaySeconds: container.startupProbe.initialDelaySeconds,
                      timeoutSeconds: container.startupProbe.timeoutSeconds,
                      periodSeconds: container.startupProbe.periodSeconds,
                      successThreshold: container.startupProbe.successThreshold,
                      failureThreshold: container.startupProbe.failureThreshold
                    }
                  : null,
                imagePullPolicy: container.imagePullPolicy,
                lifecycle: container.lifecycle
                  ? {
                      postStart: container.lifecycle.postStart
                        ? {
                            exec: container.lifecycle.postStart.exec
                              ? {
                                  command: container.lifecycle.postStart.exec.command
                                }
                              : null,
                            httpGet: container.lifecycle.postStart.httpGet
                              ? {
                                  path: container.lifecycle.postStart.httpGet.path,
                                  port: container.lifecycle.postStart.httpGet.port,
                                  scheme: container.lifecycle.postStart.httpGet.scheme,
                                  httpHeaders: container.lifecycle.postStart.httpGet.httpHeaders
                                }
                              : null,
                            tcpSocket: container.lifecycle.postStart.tcpSocket
                              ? {
                                  port: container.lifecycle.postStart.tcpSocket.port
                                }
                              : null
                          }
                        : null,
                      preStop: container.lifecycle.preStop
                        ? {
                            exec: container.lifecycle.preStop.exec
                              ? {
                                  command: container.lifecycle.preStop.exec.command
                                }
                              : null,
                            httpGet: container.lifecycle.preStop.httpGet
                              ? {
                                  path: container.lifecycle.preStop.httpGet.path,
                                  port: container.lifecycle.preStop.httpGet.port,
                                  scheme: container.lifecycle.preStop.httpGet.scheme,
                                  httpHeaders: container.lifecycle.preStop.httpGet.httpHeaders
                                }
                              : null,
                            tcpSocket: container.lifecycle.preStop.tcpSocket
                              ? {
                                  port: container.lifecycle.preStop.tcpSocket.port
                                }
                              : null
                          }
                        : null
                    }
                  : null
              })
            }
            // handle all init container of workload
            if (component.template.deployments[0].spec.template.spec.initContainers) {
              for (const initContainer of component.template.deployments[0].spec.template.spec.initContainers) {
                // handle resource of init container
                let initCpuRequest = null;
                let initMemRequest = null;
                let initCpuLimit = null;
                let initMemLimit = null;
                if (initContainer.resources) {
                  if (initContainer.resources.requests.cpu) {
                    const initCpuRequestList = initContainer.resources.requests.cpu.split('m', 2);
                    initCpuRequest = initCpuRequestList.length === 2 ? initCpuRequestList[0] : Number(initCpuRequestList[0]) * 1000;
                  }
                  if (initContainer.resources.requests.memory) {
                    const initMemRequestList = initContainer.resources.requests.memory.split('Mi', 2);
                    initMemRequest = initMemRequestList.length === 2 ? initMemRequestList[0] : Number(initContainer.resources.requests.memory.split('Gi', 2)[0]) * 1024;
                  }
                  if (initContainer.resources.limits.cpu) {
                    const initCpuLimitList = initContainer.resources.limits.cpu.split('m', 2);
                    initCpuLimit = initCpuLimitList.length === 2 ? initCpuLimitList[0] : Number(initCpuLimitList[0]) * 1000;
                  }
                  if (initContainer.resources.limits.memory) {
                    const initMemLimitList = initContainer.resources.requests.memory.split('Mi', 2);
                    initMemLimit = initMemLimitList.length === 2 ? initMemLimitList[0] : Number(initContainer.resources.limits.memory.split('Gi', 2)[0]) * 1024;
                  }
                }
                // handle volumeMount of container
                const allVolumeMountOfInitContainer: AddPodSpecContainerVolumeMount[] = [];
                if (initContainer.volumeMounts) {
                  for (const volumeMount of initContainer.volumeMounts) {
                    allVolumeMountOfInitContainer.push({
                      name: volumeMount.name,
                      mountPath: volumeMount.mountPath,
                      subPath: volumeMount.subPath
                    })
                  }
                }
                this.allInitContainers.push({
                  name: initContainer.name,
                  image: initContainer.image,
                  command: initContainer.command ? initContainer.command : null,
                  args: initContainer.args ? initContainer.args : null,
                  workingDir: initContainer.workingDir ? initContainer.workingDir : null,
                  env: initContainer.env,
                  volumeMounts: allVolumeMountOfInitContainer,
                  resources: {
                    requests: {
                      cpu: initCpuRequest ? initCpuRequest + 'm' : null,
                      memory: initMemRequest ? initMemRequest + 'Mi' : null
                    },
                    limits: {
                      cpu: initCpuLimit ? initCpuLimit + 'm' : null,
                      memory: initMemLimit ? initMemLimit + 'Mi' : null
                    }
                  },
                  imagePullPolicy: initContainer.imagePullPolicy
                })
              }
            }
            // set compare string of advancedParams
            const compare = 'vmware-system-map-registry'+'default' + 'Always' + 'ClusterFirst' + 'default-scheduler' + '30' + '600' + '10' + 'RollingUpdate' + '25%' + '25%';
            let currentParam = component.template.deployments[0].spec.template.spec.imagePullSecrets.join() +
              component.template.deployments[0].spec.template.spec.serviceAccount +
              component.template.deployments[0].spec.template.spec.restartPolicy +
              component.template.deployments[0].spec.template.spec.dnsPolicy +
              component.template.deployments[0].spec.template.spec.schedulerName +
              component.template.deployments[0].spec.template.spec.terminationGracePeriodSeconds +
              component.template.deployments[0].spec.progressDeadlineSeconds +
              component.template.deployments[0].spec.revisionHistoryLimit +
              component.template.deployments[0].spec.strategy.type
            if (component.template.deployments[0].spec.strategy.type === 'RollingUpdate') {
              currentParam = currentParam + component.template.deployments[0].spec.strategy.rollingUpdate.maxUnavailable + component.template.deployments[0].spec.strategy.rollingUpdate.maxSurge
            }
            if (compare !== currentParam || !!component.template.deployments[0].spec.template.spec.initContainers) {
              this.editWorkloadParamsForm.patchValue({
                advancedParams: true
              })
            }
            // set workload params
            this.editWorkloadParamsForm.patchValue({
              workloadType: 'Deployment',
              replicas: component.template.deployments[0].spec.replicas,
              imagePullSecrets: component.template.deployments[0].spec.template.spec.imagePullSecrets,
              serviceAccount: component.template.deployments[0].spec.template.spec.serviceAccount,
              restartPolicy: component.template.deployments[0].spec.template.spec.restartPolicy,
              dnsPolicy: component.template.deployments[0].spec.template.spec.dnsPolicy,
              schedulerName: component.template.deployments[0].spec.template.spec.schedulerName,
              terminationGracePeriodSeconds: component.template.deployments[0].spec.template.spec.terminationGracePeriodSeconds,
              progressDeadlineSeconds: component.template.deployments[0].spec.progressDeadlineSeconds,
              revisionHistoryLimit: component.template.deployments[0].spec.revisionHistoryLimit,
              strategyType: component.template.deployments[0].spec.strategy.type,
              maxUnavailable: component.template.deployments[0].spec.strategy.rollingUpdate ? component.template.deployments[0].spec.strategy.rollingUpdate.maxUnavailable : '25%',
              maxSurge: component.template.deployments[0].spec.strategy.rollingUpdate ? component.template.deployments[0].spec.strategy.rollingUpdate.maxSurge : '25%'
            })

            // handle service
            if (component.template.services) {
              for (const svc of component.template.services) {
                // handle ports
                this.allPortList[this.serviceList.length] = new FormArray([]);
                for (const port of svc.spec.ports) {
                  const portControl = this.fb.group({
                    servicePort: new FormControl(port.port, Validators.required),
                    protocol: new FormControl(port.protocol),
                    workloadPortName: new FormControl(port.targetPort)
                  })
                  this.allPortList[this.serviceList.length].push(portControl)
                }
                const svcControl = this.fb.group({
                  name: svc.name,
                  serviceType: svc.spec.type,
                  ports: this.allPortList[this.serviceList.length]
                })
                this.serviceList.push(svcControl)
                this.removeService.push({
                  name: svc.name
                })
              }
            }
            // handle other params
            const labels = component.template.deployments[0].labels
            for (const key of Object.keys(labels)) {
              if (!this.disAllowLabels.find(x => x === key)) {
                this.labelsList.push(new FormGroup({
                  key: new FormControl(key),
                  value: new FormControl(labels[key])
                }))
              }
            }
            if (component.template.deployments[0].hpa) {
              this.editOtherResourceForm.patchValue({
                enableHpa: true,
                hpaMetrics: component.template.deployments[0].hpa.targetCPUUtilizationPercentage ? 'CPU' : 'MEMORY',
                percentage: component.template.deployments[0].hpa.targetCPUUtilizationPercentage ? component.template.deployments[0].hpa.targetCPUUtilizationPercentage : component.template.deployments[0].hpa.targetMemoryUtilizationPercentage,
                minReplicas: component.template.deployments[0].hpa.minReplicas,
                maxReplicas: component.template.deployments[0].hpa.maxReplicas
              })
            }
            break;
          }
          default:
            console.log('not support.')
        }
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.GET-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }))
  }

  doUpdateComponentFinish (): void {
    switch (this.selectedComponent?.type) {
      case 'RAW': {
        const rawInput: UpdateComponent = {
          id: this.selectedComponent?.id,
          template: {
            raw: this.manifests
          }
        }
        this.componentService.updateComponent(rawInput).subscribe(data => {
          if (data.errors) {
            for (const each of data.errors) {
              if (each.message) {
                this.toast.show(this.translate.instant('MESSAGE.SHARED.UPDATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
              }
            }
          }
          if (data.data?.UpdateComponent) {
            this.toast.show(`${data.data.UpdateComponent.name} ` + this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
          }
        }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
          this.doUpdateComponentReset();
          this.refresh();
        });
        break;
      }
      case 'HELM': {
        const extraParam: JSON = JSON.parse(JSON.stringify({}));
        extraParam['currentChartRegistry'] = this.chartVersionForm.value.registry;
        extraParam['currentChartProject'] = this.chartVersionForm.value.project;
        extraParam['currentChartName'] = this.chartVersionForm.value.name;
        extraParam['currentChartVersion'] = this.chartVersionForm.value.version;
        const helmInput: UpdateComponent = {
          id: this.selectedComponent?.id,
          template: {
            helm: {
              url: this.chartVersionForm.value.url,
              values: YAML.parse(this.newValues),
              extraInfo: extraParam
            }
          }
        }
        this.componentService.updateComponent(helmInput).subscribe(data => {
          if (data.errors) {
            for (const each of data.errors) {
              if (each.message) {
                this.toast.show(this.translate.instant('MESSAGE.SHARED.CREATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
              }
            }
          }
          if (data.data?.UpdateComponent) {
            this.toast.show(`${data.data.UpdateComponent.name} ` + this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
          }
        }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
          this.doUpdateComponentReset();
          this.refresh();
        });
        break;
      }
      case 'BASIC': {
        const addSecretStorageList: AddBasicComponentTemplateSecret[] = [];
        const addConfigmapStorageList: AddBasicComponentTemplateConfigMap[] = [];
        const addPVCStorageList: AddBasicComponentTemplatePersistentVolumeClaim[] = [];
        for (const storage of this.allStorages) {
          switch (storage.storageType) {
            case 'configMap':
              addConfigmapStorageList.push({
                name: storage.name,
                data: storage.configMap.data
              })
              break;
            case 'secret':
              addSecretStorageList.push({
                name: storage.name,
                generic: {
                  type: storage.secretType,
                  data: storage.secret.data
                }
              })
              break;
            case 'persistentVolumeClaim':
              addPVCStorageList.push({
                name: storage.name,
                spec: storage.persistentVolumeClaim.spec
              })
              break;
            default:
              console.log('not support')
          }
        }
        // add lables to workload and service
        const labelsMap: JSON = JSON.parse(JSON.stringify({}));
        if (this.editOtherResourceForm.value.labels) {
          for (const each of this.editOtherResourceForm.value.labels) {
            labelsMap[each.key] = each.value;
          }
        }
        const workloads: AddBasicComponentTemplateDeployment[] = [{
          name: this.selectedComponent?.name,
          labels: labelsMap,
          spec: {
            replicas: this.editWorkloadParamsForm.value.replicas,
            template: {
              spec: {
                initContainers: (this.editWorkloadParamsForm.value.advancedParams && this.allInitContainers.length > 0) ? this.allInitContainers : null,
                containers: this.allContainers,
                volumes: this.allVolumes,
                imagePullSecrets: this.editWorkloadParamsForm.value.imagePullSecrets ? this.editWorkloadParamsForm.value.imagePullSecrets : ['vmware-system-map-registry'],
                dnsPolicy: this.editWorkloadParamsForm.value.dnsPolicy ? this.editWorkloadParamsForm.value.dnsPolicy : 'ClusterFirst',
                restartPolicy: this.editWorkloadParamsForm.value.restartPolicy ? this.editWorkloadParamsForm.value.restartPolicy : 'Always',
                schedulerName: !this.editWorkloadParamsForm.value.schedulerName ? this.editWorkloadParamsForm.value.schedulerName : 'default-scheduler',
                serviceAccount: this.editWorkloadParamsForm.value.serviceAccount ? this.editWorkloadParamsForm.value.serviceAccount : 'default',
                terminationGracePeriodSeconds: this.editWorkloadParamsForm.value.terminationGracePeriodSeconds ? this.editWorkloadParamsForm.value.terminationGracePeriodSeconds : 30
              }
            },
            progressDeadlineSeconds: this.editWorkloadParamsForm.value.progressDeadlineSeconds ? this.editWorkloadParamsForm.value.progressDeadlineSeconds : 600,
            revisionHistoryLimit: this.editWorkloadParamsForm.value.revisionHistoryLimit ? this.editWorkloadParamsForm.value.revisionHistoryLimit : 10,
            strategy: {
              type: this.editWorkloadParamsForm.value.strategyType,
              rollingUpdate: this.editWorkloadParamsForm.value.strategyType === 'RollingUpdate'
                ? {
                    maxUnavailable: this.editWorkloadParamsForm.value.maxUnavailable ? this.editWorkloadParamsForm.value.maxUnavailable : '25%',
                    maxSurge: this.editWorkloadParamsForm.value.maxSurge ? this.editWorkloadParamsForm.value.maxSurge : '25%'
                  }
                : null
            }
          }
        }]
        // add hpa
        if (this.editOtherResourceForm.value.enableHpa) {
          const hpaParams: CreateHorizontalPodAutoscalerSpec = {
            maxReplicas: this.editOtherResourceForm.value.maxReplicas,
            minReplicas: this.editOtherResourceForm.value.minReplicas
          }
          if (this.editOtherResourceForm.value.hpaMetrics === 'CPU') {
            hpaParams.targetCPUUtilizationPercentage = this.editOtherResourceForm.value.percentage;
          } else {
            hpaParams.targetMemoryUtilizationPercentage = this.editOtherResourceForm.value.percentage;
          }
          workloads[0].hpa = hpaParams;
        }
        // add service
        const componentServiceList: AddBasicComponentTemplateService[] = [];
        if (this.editServiceForm.value.services) {
          for (let i = 0; i < this.serviceList.length; i++) {
            const service = this.serviceList.controls[i];
            const ports: AddServiceSpecPort[] = [];
            for (const port of this.allPortList[i].controls) {
              ports.push({
                name: port.value.workloadPortName + '-' + port.value.servicePort,
                port: port.value.servicePort,
                protocol: port.value.protocol,
                targetPort: port.value.workloadPortName
              })
            }
            componentServiceList.push({
              name: service.value.name,
              labels: labelsMap,
              spec: {
                type: service.value.serviceType,
                ports: ports
              }
            })
          }
        }
        const basicInput: UpdateComponent = {
          id: this.selectedComponent?.id,
          template: {
            basic: {
              addDeployments: workloads,
              removeDeployments: this.removeDeployment,
              addServices: componentServiceList,
              removeServices: this.removeService,
              addSecrets: addSecretStorageList,
              removeSecrets: this.removeSecretStorage,
              addConfigMaps: addConfigmapStorageList,
              removeConfigMaps: this.removeConfigMapStorage,
              addPersistentVolumeClaims: addPVCStorageList,
              removePersistentVolumeClaims: this.removePVCStprage
            }
          }
        }
        this.componentService.updateComponent(basicInput).subscribe(data => {
          if (data.errors) {
            for (const each of data.errors) {
              if (each.message) {
                this.toast.show(this.translate.instant('MESSAGE.SHARED.UPDATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
              }
            }
          }
          if (data.data?.UpdateComponent) {
            this.toast.show(`${data.data.UpdateComponent.name} ` + this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
          }
        }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT-DETAILS.UPDATE-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
          this.doUpdateComponentReset();
          this.refresh();
        });
        break;
      }
      default:
        console.log('not support')
    }
  }

  doUpdateComponentCancel (): void {
    this.doUpdateComponentReset();
  }

  doUpdateComponentReset (): void {
    this.wizardeditcomponent.reset();
    this.updateComponentOpen = false;
    this.chartVersionForm.reset();
    this.chartVersionListSorted = [];
    this.noChartVersionAlert = false;
    this.valueload = false;
    this.newValues = '';
    this.originValues = '';
    this.manifests = '';
    this.showCheckedAlert = false;
    this.checkFormateOk = false;
    this.parseError = '';
    this.allStorages = [];
    this.allVolumes = [];
    this.allContainers = [];
    this.allInitContainers = [];
    this.removeService = [];
    this.dockerConfigSecretList = null;
    this.editOtherResourceForm.reset();
    this.editOtherResourceForm.patchValue({
      enableHpa: false,
      hpaMetrics: 'CPU',
      percentage: 80,
      minReplicas: 1,
      maxReplicas: 10
    })
    this.hpaReplicasInvaild = false;
    while (this.labelsList.length > 0) {
      this.labelsList.removeAt(0)
    }
    this.editServiceForm.reset();
    while (this.serviceList.length > 0) {
      this.serviceList.removeAt(0)
    }
    this.allPortList = [];
    this.editWorkloadParamsForm.reset();
    this.editWorkloadParamsForm.patchValue({
      workloadType: 'Deployment',
      replicas: 1,
      advancedParams: false,
      imagePullSecrets: ['vmware-system-map-registry'],
      serviceAccount: 'default',
      restartPolicy: 'Always',
      dnsPolicy: 'ClusterFirst',
      schedulerName: 'default-scheduler',
      terminationGracePeriodSeconds: 30,
      progressDeadlineSeconds: 600,
      revisionHistoryLimit: 10,
      strategyType: 'RollingUpdate',
      maxUnavailable: '25%',
      maxSurge: '25%'
    })
  }

  //clone component
  showCloneComponent(cpt: CustomComponent): void {
    this.selectedComponent = cpt;
    this.componentService.getComponentById(cpt.id).subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.GET-CPT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      const component: CustomComponent = !!data?.data?.Component? data.data.Component : null;
      if(!!component){
         this.allComponentApplications = deepCopy(this.allApplications);
         this.showNoApplicationAlert = this.allComponentApplications.length <= 0;
         this.cloneComponentForm.setValue({
           name: component.name + '-clone',
           description: component.description,
           application: component.application.id,
           type: component.type
         })
         this.cloneComponentOpen = true;
         switch (component.type) {
          case 'RAW':
            this.manifests = component.template.raw
            break;
          case 'HELM':
            const params = component.template.extraInfo;
            this.chartVersionForm.patchValue({
              registry: params['currentChartRegistry'],
              project: params['currentChartProject'],
              name: params['currentChartName'],
              version: params['currentChartVersion'],
              url: component.template.url
            })
            this.getChartVersionList();
            this.newValues = YAML.stringify(component.template.values);
            this.originValues = component.template.chart.files['values.yaml']
            break;
          case 'BASIC':
            // handle all storage
            if(!!component.template.configMaps){
              let configMap: BasicComponentTemplateConfigMap;
              for (configMap of component.template.configMaps){
                this.allStorages.push({
                  name: configMap.name,
                  storageType: 'configMap',
                  configMap: {
                    data: configMap.data
                  }
                })
              }
            }
            if(!!component.template.persistentVolumeClaims){
              let pvc: BasicComponentTemplatePersistentVolumeClaim;
              for (pvc of component.template.persistentVolumeClaims){
                this.allStorages.push({
                  name: pvc.name,
                  storageType: 'persistentVolumeClaim',
                  persistentVolumeClaim: {
                    spec: {
                      storageClass: !!pvc.spec.storageClass ? pvc.spec.storageClass : '',
                      accessModes: pvc.spec.accessModes,
                      resources: {
                        requests: {
                          storage: pvc.spec.resources.requests.storage
                        },
                        limits: {
                          storage: !!pvc.spec.resources.limits.storage ? pvc.spec.resources.limits.storage : ''
                        }
                      },
                      volumeMode: pvc.spec.volumeMode
                    }
                  }
                })
              }
            }
            if(!!component.template.secrets){
              for (let secret of component.template.secrets){
                this.allStorages.push({
                  name: secret.name,
                  storageType: 'secret',
                  secretType: secret.type,
                  secret: {
                    data: secret.data,
                  }
                })
              }
            }
            //handle all volumes of workload
            if(!!component.template.deployments[0].spec.template.spec.volumes){
              for(const volume of component.template.deployments[0].spec.template.spec.volumes){
                const volumeItemsList: TemplateSpecVolumeKeyToPathInput[] = [];
                if(volume.volume.__typename === 'ConfigMapVolume' || volume.volume.__typename === 'SecretVolume'){
                  if(!!volume.volume.items){
                    for (const item of volume.volume.items){
                      volumeItemsList.push({
                        key: item.key,
                        path: item.path,
                        mode: item.mode
                      })
                    }
                  }
                }
                this.allVolumes.push({
                  name: volume.name,
                  emptyDir: volume.volume.__typename === 'EmptyDir' ? {
                      medium: !!volume.volume.medium ? volume.volume.medium : '',
                      size: volume.volume.size
                    } : null,
                  configMap: volume.volume.__typename === 'ConfigMapVolume' ? {
                      name: volume.volume.name,
                      defaultMode: volume.volume.defaultMode,
                      items: volumeItemsList
                    } : null,
                  secret: volume.volume.__typename === 'SecretVolume' ? {
                      name: volume.volume.name,
                      defaultMode: volume.volume.defaultMode,
                      items: volumeItemsList
                    } : null,
                  persistentVolumeClaim: volume.volume.__typename === 'PersistentVolumeClaimVolume' ? {
                      name: volume.volume.name,
                      readOnly: volume.volume.readOnly
                    } : null,
                })
              }
            }
            // handle all containers of workload
            for(const container of component.template.deployments[0].spec.template.spec.containers){
                // handle resource of container
              let cpuRequest = null;
              let memRequest = null;
              let cpuLimit = null;
              let memLimit = null;
              if(!!container.resources){
                if(!!container.resources.requests.cpu){
                  const cpuRequestList = container.resources.requests.cpu.split('m', 2);
                  cpuRequest = cpuRequestList.length === 2 ? cpuRequestList[0] : Number(cpuRequestList[0]) * 1000;
                }
                if(!!container.resources.requests.memory){
                  const memRequestList = container.resources.requests.memory.split('Mi', 2);
                  memRequest = memRequestList.length === 2 ? memRequestList[0] : Number(container.resources.requests.memory.split('Gi', 2)[0]) * 1024;
                }
                if(!!container.resources.limits.cpu){
                  const cpuLimitList = container.resources.limits.cpu.split('m', 2);
                  cpuLimit = cpuLimitList.length === 2 ? cpuLimitList[0] : Number(cpuLimitList[0]) * 1000;
                }
                if(!!container.resources.limits.memory){
                  const memLimitList = container.resources.requests.memory.split('Mi', 2);
                  memLimit = memLimitList.length === 2 ? memLimitList[0] : Number(container.resources.limits.memory.split('Gi', 2)[0]) * 1024;
                }
              }
              //handle ports of container
              const allPortsOfContainer: AddPodSpecContainerPort[] = [];
              if(!!container.ports){
                for (let port of container.ports){
                  allPortsOfContainer.push({
                    name: port.name,
                    protocol: port.protocol,
                    containerPort: port.containerPort
                  })
                }
              }
              //handle volumeMount of container
              const allVolumeMountOfContainer: AddPodSpecContainerVolumeMount[] = [];
              if(!!container.volumeMounts){
                for (let volumeMount of container.volumeMounts) {
                  allVolumeMountOfContainer.push({
                    name: volumeMount.name,
                    mountPath: volumeMount.mountPath,
                    subPath: volumeMount.subPath
                  })
                }
              }
              this.allContainers.push({
                name: container.name,
                image: container.image,
                command: !!container.command ? container.command : null,
                args: !!container.args ? container.args : null,
                workingDir: !!container.workingDir ? container.workingDir : null,
                ports: allPortsOfContainer,
                env: container.env,
                volumeMounts: allVolumeMountOfContainer,
                resources: {
                  requests: {
                    cpu: !!cpuRequest ? cpuRequest + 'm' : null,
                    memory: !!memRequest ? memRequest + 'Mi' : null,
                  },
                  limits: {
                    cpu: !!cpuLimit ? cpuLimit + 'm' : null,
                    memory: !!memLimit ? memLimit + 'Mi' : null
                  }
                },
                readinessProbe: !!container.readinessProbe ? {
                  exec: !!container.readinessProbe.exec ? {
                    command: container.readinessProbe.exec.command
                  } : null,
                  httpGet: !!container.readinessProbe.httpGet ? {
                    path: container.readinessProbe.httpGet.path,
                    port: container.readinessProbe.httpGet.port,
                    scheme: container.readinessProbe.httpGet.scheme,
                    httpHeaders: container.readinessProbe.httpGet.httpHeaders
                  } : null,
                  tcpSocket: !!container.readinessProbe.tcpSocket ? {
                    port: container.readinessProbe.tcpSocket.port
                  } : null,
                  initialDelaySeconds: container.readinessProbe.initialDelaySeconds,
                  timeoutSeconds: container.readinessProbe.timeoutSeconds,
                  periodSeconds: container.readinessProbe.periodSeconds,
                  successThreshold: container.readinessProbe.successThreshold,
                  failureThreshold: container.readinessProbe.failureThreshold
                } : null,
                livenessProbe: !!container.livenessProbe ? {
                  exec: !!container.livenessProbe.exec ? {
                    command: container.livenessProbe.exec.command
                  } : null,
                  httpGet: !!container.livenessProbe.httpGet ? {
                    path: container.livenessProbe.httpGet.path,
                    port: container.livenessProbe.httpGet.port,
                    scheme: container.livenessProbe.httpGet.scheme,
                    httpHeaders: container.livenessProbe.httpGet.httpHeaders
                  } : null,
                  tcpSocket: !!container.livenessProbe.tcpSocket ? {
                    port: container.livenessProbe.tcpSocket.port
                  } : null,
                  initialDelaySeconds: container.livenessProbe.initialDelaySeconds,
                  timeoutSeconds: container.livenessProbe.timeoutSeconds,
                  periodSeconds: container.livenessProbe.periodSeconds,
                  successThreshold: container.livenessProbe.successThreshold,
                  failureThreshold: container.livenessProbe.failureThreshold
                } : null,
                startupProbe: !!container.startupProbe ? {
                  exec: !!container.startupProbe.exec ? {
                    command: container.startupProbe.exec.command
                  } : null,
                  httpGet: !!container.startupProbe.httpGet ? {
                    path: container.startupProbe.httpGet.path,
                    port: container.startupProbe.httpGet.port,
                    scheme: container.startupProbe.httpGet.scheme,
                    httpHeaders: container.startupProbe.httpGet.httpHeaders
                  } : null,
                  tcpSocket: !!container.startupProbe.tcpSocket ? {
                    port: container.startupProbe.tcpSocket.port
                  } : null,
                  initialDelaySeconds: container.startupProbe.initialDelaySeconds,
                  timeoutSeconds: container.startupProbe.timeoutSeconds,
                  periodSeconds: container.startupProbe.periodSeconds,
                  successThreshold: container.startupProbe.successThreshold,
                  failureThreshold: container.startupProbe.failureThreshold
                } : null,
                imagePullPolicy: container.imagePullPolicy,
                lifecycle: !!container.lifecycle ? {
                  postStart: !!container.lifecycle.postStart ? {
                    exec: !!container.lifecycle.postStart.exec ? {
                      command: container.lifecycle.postStart.exec.command
                    } : null,
                    httpGet: !!container.lifecycle.postStart.httpGet ? {
                      path: container.lifecycle.postStart.httpGet.path,
                      port: container.lifecycle.postStart.httpGet.port,
                      scheme: container.lifecycle.postStart.httpGet.scheme,
                      httpHeaders: container.lifecycle.postStart.httpGet.httpHeaders
                    } : null,
                    tcpSocket: !!container.lifecycle.postStart.tcpSocket ? {
                      port: container.lifecycle.postStart.tcpSocket.port
                    } : null
                  } : null,
                  preStop: !!container.lifecycle.preStop ? {
                    exec: !!container.lifecycle.preStop.exec ? {
                      command: container.lifecycle.preStop.exec.command
                    } : null,
                    httpGet: !!container.lifecycle.preStop.httpGet ? {
                      path: container.lifecycle.preStop.httpGet.path,
                      port: container.lifecycle.preStop.httpGet.port,
                      scheme: container.lifecycle.preStop.httpGet.scheme,
                      httpHeaders: container.lifecycle.preStop.httpGet.httpHeaders
                    } : null,
                    tcpSocket: !!container.lifecycle.preStop.tcpSocket ? {
                      port: container.lifecycle.preStop.tcpSocket.port
                    } : null
                  } : null
                } : null
              })
            }
            // handle all init container of workload
            if(!!component.template.deployments[0].spec.template.spec.initContainers){
              for(const initContainer of component.template.deployments[0].spec.template.spec.initContainers){
                // handle resource of init container
                let initCpuRequest = null;
                let initMemRequest = null;
                let initCpuLimit = null;
                let initMemLimit = null;
                if(!!initContainer.resources){
                  if(!!initContainer.resources.requests.cpu){
                    const initCpuRequestList = initContainer.resources.requests.cpu.split('m', 2);
                    initCpuRequest = initCpuRequestList.length === 2 ? initCpuRequestList[0] : Number(initCpuRequestList[0]) * 1000;
                  }
                  if(!!initContainer.resources.requests.memory){
                    const initMemRequestList = initContainer.resources.requests.memory.split('Mi', 2);
                    initMemRequest = initMemRequestList.length === 2 ? initMemRequestList[0] : Number(initContainer.resources.requests.memory.split('Gi', 2)[0]) * 1024;
                  }
                  if(!!initContainer.resources.limits.cpu){
                    const initCpuLimitList = initContainer.resources.limits.cpu.split('m', 2);
                    initCpuLimit = initCpuLimitList.length === 2 ? initCpuLimitList[0] : Number(initCpuLimitList[0]) * 1000;
                  }
                  if(!!initContainer.resources.limits.memory){
                    const initMemLimitList = initContainer.resources.requests.memory.split('Mi', 2);
                    initMemLimit = initMemLimitList.length === 2 ? initMemLimitList[0] : Number(initContainer.resources.limits.memory.split('Gi', 2)[0]) * 1024;
                  }
                }
                //handle volumeMount of container
                const allVolumeMountOfInitContainer: AddPodSpecContainerVolumeMount[] = [];
                if(!!initContainer.volumeMounts){
                  for (let volumeMount of initContainer.volumeMounts) {
                    allVolumeMountOfInitContainer.push({
                      name: volumeMount.name,
                      mountPath: volumeMount.mountPath,
                      subPath: volumeMount.subPath
                    })
                  }
                }
                this.allInitContainers.push({
                  name: initContainer.name,
                  image: initContainer.image,
                  command: !!initContainer.command ? initContainer.command : null,
                  args: !!initContainer.args ? initContainer.args : null,
                  workingDir: !!initContainer.workingDir ? initContainer.workingDir : null,
                  env: initContainer.env,
                  volumeMounts: allVolumeMountOfInitContainer,
                  resources: {
                    requests: {
                      cpu: !!initCpuRequest ? initCpuRequest + 'm' : null,
                      memory: !!initMemRequest ? initMemRequest + 'Mi' : null,
                    },
                    limits: {
                      cpu: !!initCpuLimit ? initCpuLimit + 'm' : null,
                      memory: !!initMemLimit ? initMemLimit + 'Mi' : null
                    }
                  },
                  imagePullPolicy: initContainer.imagePullPolicy
                })
              }
            }
            // set compare string of advancedParams
            const compare = 'vmware-system-map-registry'+'default'+'Always'+'ClusterFirst'+'default-scheduler'+'30'+'600'+'10'+'RollingUpdate'+'25%'+'25%';
            let currentParam = component.template.deployments[0].spec.template.spec.imagePullSecrets.join()+
              component.template.deployments[0].spec.template.spec.serviceAccount+
              component.template.deployments[0].spec.template.spec.restartPolicy+
              component.template.deployments[0].spec.template.spec.dnsPolicy+
              component.template.deployments[0].spec.template.spec.schedulerName+
              component.template.deployments[0].spec.template.spec.terminationGracePeriodSeconds+
              component.template.deployments[0].spec.progressDeadlineSeconds+
              component.template.deployments[0].spec.revisionHistoryLimit+
              component.template.deployments[0].spec.strategy.type
            if(component.template.deployments[0].spec.strategy.type === 'RollingUpdate'){
              currentParam = currentParam + component.template.deployments[0].spec.strategy.rollingUpdate.maxUnavailable + component.template.deployments[0].spec.strategy.rollingUpdate.maxSurge
            }
            if(compare !== currentParam || !!component.template.deployments[0].spec.template.spec.initContainers){
              this.editWorkloadParamsForm.patchValue({
                advancedParams: true
              })
            }
            // set workload params
            this.editWorkloadParamsForm.patchValue({
              workloadType: 'Deployment',
              replicas: component.template.deployments[0].spec.replicas,
              imagePullSecrets: component.template.deployments[0].spec.template.spec.imagePullSecrets,
              serviceAccount: component.template.deployments[0].spec.template.spec.serviceAccount,
              restartPolicy: component.template.deployments[0].spec.template.spec.restartPolicy,
              dnsPolicy: component.template.deployments[0].spec.template.spec.dnsPolicy,
              schedulerName: component.template.deployments[0].spec.template.spec.schedulerName,
              terminationGracePeriodSeconds: component.template.deployments[0].spec.template.spec.terminationGracePeriodSeconds,
              progressDeadlineSeconds: component.template.deployments[0].spec.progressDeadlineSeconds,
              revisionHistoryLimit: component.template.deployments[0].spec.revisionHistoryLimit,
              strategyType: component.template.deployments[0].spec.strategy.type,
              maxUnavailable: !!component.template.deployments[0].spec.strategy.rollingUpdate ? component.template.deployments[0].spec.strategy.rollingUpdate.maxUnavailable : '25%',
              maxSurge: !!component.template.deployments[0].spec.strategy.rollingUpdate ? component.template.deployments[0].spec.strategy.rollingUpdate.maxSurge : '25%'
            })

            // handle service
            if(!!component.template.services){
              for(const svc of component.template.services){
                //handle ports
                this.allPortList[this.serviceList.length] = new FormArray([]);
                for (const port of svc.spec.ports){
                  const portControl = this.fb.group({
                    servicePort: new FormControl(port.port, Validators.required),
                    protocol: new FormControl(port.protocol),
                    workloadPortName: new FormControl(port.targetPort)
                  })
                  this.allPortList[this.serviceList.length].push(portControl)
                }
                const svcControl = this.fb.group({
                  name: svc.name,
                  serviceType: svc.spec.type,
                  ports: this.allPortList[this.serviceList.length]
                })
                this.serviceList.push(svcControl)
              }
            }
            // handle other params
            const labels = component.template.deployments[0].labels
            for (const key of Object.keys(labels)) {
              if(!this.disAllowLabels.find(x => x === key)){
                this.labelsList.push(new FormGroup({
                  key: new FormControl(key),
                  value: new FormControl(labels[key])
                }))
              }
            }
            if(!!component.template.deployments[0].hpa){
              this.editOtherResourceForm.patchValue({
                enableHpa: true,
                hpaMetrics: !!component.template.deployments[0].hpa.targetCPUUtilizationPercentage ? 'CPU' : 'MEMORY',
                percentage: !!component.template.deployments[0].hpa.targetCPUUtilizationPercentage ? component.template.deployments[0].hpa.targetCPUUtilizationPercentage : component.template.deployments[0].hpa.targetMemoryUtilizationPercentage,
                minReplicas: component.template.deployments[0].hpa.minReplicas,
                maxReplicas: component.template.deployments[0].hpa.maxReplicas
              })
            }
            break;
          default:
            console.log('not support.')
         }
      }
    },(e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.GET-CPT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
  }

  doCloneComponentFinish(): void {
    const type = this.cloneComponentForm.value.type;
    const name = this.cloneComponentForm.value.name;
    const desc = this.cloneComponentForm.value.description;
    const app = this.cloneComponentForm.value.application;
    switch (type) {
      case 'RAW':
        const rawInput: CreateComponent = {
          name: name,
          description: desc,
          template: {
            raw: this.manifests
          },
          application: app
        }
        this.componentService.createComponent(rawInput).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.CreateComponent){
            this.toast.show(`${data.data.CreateComponent.name} `+this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          this.doCloneComponentReset();
          this.refresh();
        });
        break;
      case 'HELM':
        const extraParam: JSON = JSON.parse(JSON.stringify({}));
        extraParam['currentChartRegistry'] = this.chartVersionForm.value.registry;
        extraParam['currentChartProject'] = this.chartVersionForm.value.project;
        extraParam['currentChartName'] = this.chartVersionForm.value.name;
        extraParam['currentChartVersion'] = this.chartVersionForm.value.version;
        const helmInput: CreateComponent = {
          name: name,
          description: desc,
          template: {
            helm: {
              url: this.chartVersionForm.value.url,
              values: YAML.parse(this.newValues),
              extraInfo: extraParam
            }
          },
          application: app
        }
        this.componentService.createComponent(helmInput).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.CreateComponent){
            this.toast.show(`${data.data.CreateComponent.name} `+this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          this.doCloneComponentReset();
          this.refresh();
        });
        break;
      case 'BASIC':
        const addSecretStorageList: AddBasicComponentTemplateSecret[] = [];
        const addConfigmapStorageList: AddBasicComponentTemplateConfigMap[] = [];
        const addPVCStorageList: AddBasicComponentTemplatePersistentVolumeClaim[] = [];
        for (let storage of this.allStorages){
          switch (storage.storageType) {
            case 'configMap':
              addConfigmapStorageList.push({
                name: storage.name,
                data: storage.configMap.data
              })
              break;
            case 'secret':
              addSecretStorageList.push({
                name: storage.name,
                generic: {
                  type: storage.secretType,
                  data: storage.secret.data
                }
              })
              break;
            case 'persistentVolumeClaim':
              addPVCStorageList.push({
                name: storage.name,
                spec: storage.persistentVolumeClaim.spec
              })
              break;
            default:
              console.log('not support')
          }
        }
        // add lables to workload and service
         const labelsMap: JSON = JSON.parse(JSON.stringify({}));
         if (!!this.editOtherResourceForm.value.labels) {
           for (const each of this.editOtherResourceForm.value.labels) {
             labelsMap[each.key] = each.value;
           }
         }
         const workloads: AddBasicComponentTemplateDeployment[] = [{
           name: name,
           labels: labelsMap,
           spec: {
             replicas: this.editWorkloadParamsForm.value.replicas,
             template: {
               spec: {
                 initContainers: (this.editWorkloadParamsForm.value.advancedParams && this.allInitContainers.length > 0) ? this.allInitContainers : null,
                 containers: this.allContainers,
                 volumes: this.allVolumes,
                 imagePullSecrets: !!this.editWorkloadParamsForm.value.imagePullSecrets ? this.editWorkloadParamsForm.value.imagePullSecrets : ['vmware-system-map-registry'],
                 dnsPolicy: !!this.editWorkloadParamsForm.value.dnsPolicy ? this.editWorkloadParamsForm.value.dnsPolicy : 'ClusterFirst',
                 restartPolicy: !!this.editWorkloadParamsForm.value.restartPolicy ? this.editWorkloadParamsForm.value.restartPolicy : 'Always',
                 schedulerName: !this.editWorkloadParamsForm.value.schedulerName ? this.editWorkloadParamsForm.value.schedulerName : 'default-scheduler',
                 serviceAccount: !!this.editWorkloadParamsForm.value.serviceAccount ? this.editWorkloadParamsForm.value.serviceAccount : 'default',
                 terminationGracePeriodSeconds: !!this.editWorkloadParamsForm.value.terminationGracePeriodSeconds ? this.editWorkloadParamsForm.value.terminationGracePeriodSeconds : 30
               }
             },
             progressDeadlineSeconds: !!this.editWorkloadParamsForm.value.progressDeadlineSeconds ? this.editWorkloadParamsForm.value.progressDeadlineSeconds : 600,
             revisionHistoryLimit: !!this.editWorkloadParamsForm.value.revisionHistoryLimit ? this.editWorkloadParamsForm.value.revisionHistoryLimit : 10,
             strategy: {
               type: this.editWorkloadParamsForm.value.strategyType,
               rollingUpdate: this.editWorkloadParamsForm.value.strategyType === 'RollingUpdate' ? {
                 maxUnavailable: this.editWorkloadParamsForm.value.maxUnavailable ? this.editWorkloadParamsForm.value.maxUnavailable : '25%',
                 maxSurge: this.editWorkloadParamsForm.value.maxSurge ? this.editWorkloadParamsForm.value.maxSurge : '25%',
               } : null
             }
           }
         }]
         // add hpa
         if(this.editOtherResourceForm.value.enableHpa){
           const hpaParams: CreateHorizontalPodAutoscalerSpec = {
             maxReplicas: this.editOtherResourceForm.value.maxReplicas,
             minReplicas: this.editOtherResourceForm.value.minReplicas,
           }
           if(this.editOtherResourceForm.value.hpaMetrics === 'CPU'){
             hpaParams.targetCPUUtilizationPercentage = this.editOtherResourceForm.value.percentage;
           } else {
             hpaParams.targetMemoryUtilizationPercentage = this.editOtherResourceForm.value.percentage;
           }
           workloads[0].hpa = hpaParams;
         }
         // add service
         const componentServiceList: AddBasicComponentTemplateService[] = [];
         if(!!this.editServiceForm.value.services){
            for(var i=0;i < this.serviceList.length; i++){
              const service = this.serviceList.controls[i];
              const ports: AddServiceSpecPort[] = [];
              for(let port of this.allPortList[i].controls){
                ports.push({
                  name: port.value.workloadPortName + '-' + port.value.servicePort,
                  port: port.value.servicePort,
                  protocol: port.value.protocol,
                  targetPort: port.value.workloadPortName
                })
              }
              componentServiceList.push({
                name: service.value.name,
                labels: labelsMap,
                spec: {
                  type: service.value.serviceType,
                  ports: ports
                }
              })
            }
         }
        const basicInput: CreateComponent = {
          name: name,
          description: desc,
          template: {
            basic: {
              services: componentServiceList,
              deployments: workloads,
              secrets: addSecretStorageList,
              configMaps: addConfigmapStorageList,
              persistentVolumeClaims: addPVCStorageList
            }
          },
          application: app
        }
        this.componentService.createComponent(basicInput).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.CreateComponent){
            this.toast.show(`${data.data.CreateComponent.name} `+this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          this.doCloneComponentReset();
          this.refresh();
        });
        break;
      default:
        console.log('not support')
    }
  }

  doCloneStoragePageCustomNext(buttonType: string): void {
    this.allConfigMapStorages = [];
    this.allSecretStorages = [];
    this.allPVCStorages = [];
    this.dockerConfigSecretList = null;
    if ("custom-next" === buttonType) {
      this.configMapStorageExist = !!this.allStorages.find(x => x.storageType === 'configMap');
      this.secretStorageExist = !!this.allStorages.find(x => x.storageType === 'secret');
      this.pvcStorageExist = !!this.allStorages.find(x => x.storageType === 'persistentVolumeClaim');
    }
    this.getDockerConfigSecret();
    this.wizardclonecomponent.next();
  }

  doCloneWorkloadPageCustomNext(buttonType: string): void {
    this.workloadPortsList = null;
    if ("custom-next" === buttonType) {
      const workloadType = this.editWorkloadParamsForm.value.workloadType;
      switch (workloadType) {
        case 'Deployment':
          const tempPortList = [];
          for(let container of this.allContainers){
            if(!!container.ports){
              for (let port of container.ports){
                tempPortList.push(port)
              }
            }
          }
          this.workloadPortsList = tempPortList;
          break;
        default:
          this.workloadPortsList = [];
          break;
      }
    }
    this.wizardclonecomponent.next();
  }

  doCloneComponentCancel(): void {
    this.doCloneComponentReset();
  }

  doCloneComponentReset(): void {
    this.wizardclonecomponent.reset();
    this.cloneComponentOpen = false;
    this.allComponentApplications = null;
    this.chartVersionForm.reset();
    this.chartVersionListSorted = [];
    this.noChartVersionAlert = false;
    this.valueload = false;
    this.newValues = '';
    this.originValues = '';
    this.manifests = '';
    this.showCheckedAlert = false;
    this.checkFormateOk = false;
    this.parseError = '';
    this.allStorages = [];
    this.allVolumes = [];
    this.allContainers = [];
    this.allInitContainers = [];
    this.dockerConfigSecretList = null;
    this.editOtherResourceForm.reset();
    this.editOtherResourceForm.patchValue({
      enableHpa: false,
      hpaMetrics: 'CPU',
      percentage: 80,
      minReplicas: 1,
      maxReplicas: 10
    })
    this.hpaReplicasInvaild = false;
    while (this.labelsList.length > 0){
      this.labelsList.removeAt(0)
    }
    this.editServiceForm.reset();
    while (this.serviceList.length > 0){
      this.serviceList.removeAt(0)
    }
    this.allPortList = [];
    this.editWorkloadParamsForm.reset();
    this.editWorkloadParamsForm.patchValue({
      workloadType: 'Deployment',
      replicas: 1,
      advancedParams: false,
      imagePullSecrets: ['vmware-system-map-registry'],
      serviceAccount: 'default',
      restartPolicy: 'Always',
      dnsPolicy: 'ClusterFirst',
      schedulerName: 'default-scheduler',
      terminationGracePeriodSeconds: 30,
      progressDeadlineSeconds: 600,
      revisionHistoryLimit: 10,
      strategyType: 'RollingUpdate',
      maxUnavailable: '25%',
      maxSurge: '25%'
    })
  }

  // remove component
  showDeleteComponent (cpt: CustomComponent): void {
    this.removeDialog = true;
    this.selectedComponent = cpt;
  }

  deleteComponentById (): void {
    this.removeDialog = false;
    this.componentService.deleteComponent(this.selectedComponent?.id).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            if (each.message === 'object is busy') {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ' 请先解除组件部署关联后再执行此操作！', this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.DELETE-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            } else {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.DELETE-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            }
          }
        }
      }
      if (data.data?.DeleteComponent) {
        this.toast.show(`${data.data.DeleteComponent.name} ` + this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.DELETE-CPT-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.DELETE-CPT-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.DELETE-CPT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  // fix "ng-reflect-model shows correct value but not reflecting in input", see https://stackoverflow.com/questions/40663198/ng-reflect-model-shows-correct-value-but-not-reflecting-in-input
  customTrackBy (index: number, obj: any): any {
    return index;
  }
}
