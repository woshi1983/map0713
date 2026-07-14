import { Component, OnInit, ViewChild } from '@angular/core';
import {ClrDatagridPagination, ClrDatagridStateInterface, ClrWizard} from '@clr/angular';
import {FormControl, FormGroup, Validators, FormArray, FormBuilder} from '@angular/forms';
import { ClustersService } from '../../../services/clusters.service';
import {
  ImageRegistry,
  LoadBalancer,
  Cluster,
  ClusterProvider,
  VMClass,
  CreateCluster,
  AttachCluster, handleObject, AddClusterProviderImageRegistry, UpdateCluster
} from '../../../classes';
import { AuthService} from '../../../services/auth.service';
import {ProviderService} from '../../../services/provider.service';
import {
  podCIDRMinNetmaskValidator,
  CustomValidatorDirective,
  CAFormatValidator,
  nameValidator
} from '../../../utils/validators/custom-validator.directive';
import {LoadbalancerService} from "../../../services/loadbalancer.service";
import {RepositorysService} from "../../../services/repositorys.service";
import semver from 'semver';
import { ToastrService } from 'ngx-toastr';
import {TranslateService} from "@ngx-translate/core";
import {OrganizationService} from "../../../services/organization.service";
import {FEATUREGATES} from "../../../utils/static-entries/shared.const";
import { checkCerts, certCheckResult } from "../../../utils/operator/check-certs";


@Component({
  selector: 'app-management',
  templateUrl: './management.component.html',
  styleUrls: ['./management.component.css']
})
export class ManagementComponent implements OnInit {

  featureGates = FEATUREGATES

  //admin role can handle cluster create/update/delete
  get canWrite(): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER);
  }

  //only superadmin can handle cluster force delete
  get canForceDelete(): boolean {
    return this.authService.can('NULL', handleObject.CLUSTER);
  }

  orgId: string
  allClusters: Cluster[] = [];
  total = 0;
  selectedCluster: Cluster;
  selectedClusterName: string;
  selectedClusterId: string;

  createClusterOpen = false;
  updateClusterOpen = false;
  attachClusterOpen = false;
  deleteClusterDialog = false;
  removeClusterDialog = false;
  forceDeleteDialog = false;

  showSearch = true;
  searchName: string;
  allowUpdateClusterMaster =false;
  showUpdateWorkerDialog = false;
  updateWorkerChanged = false;
  showUpdateTKGSMasterDialog = false;
  showUpdateTKGMMasterDialog = false;
  showUpdateTKGSVMclassDialog = false;
  updateTKGSVMclassChanged = false;
  showUpdateTKGMVMclassDialog = false;
  updateTKGMVMclassChanged = false;
  showUpdateClusterDescDialog = false;

  providerClass: string;
  providerList: ClusterProvider[] = [];
  tkcVersionList: string[] = [];
  sortedTkcVersionList: string[] = [];
  tkcvmClassList: VMClass[] = [];
  storageClassList: string[] = [];
  clusterNamespace: string;

  ClusterConnectSuccess = false;
  showClusterConnectAlert = false;
  noVMclassOnNamespaceAlert = false;
  noStorageOnNamespaceAlert = false;
  clusterCIDRInvaildAlert = false;

  registryList: ImageRegistry[] = [];
  lbList: LoadBalancer[] = [];

  tkgmNodeType = [{name: "small", cpu: 4, mem: 4, storage: 20},
    {name: "medium", cpu: 4, mem: 8, storage: 40},
    {name: "large", cpu: 4, mem: 16, storage: 40},
    {name: "extra-large", cpu: 8, mem: 32, storage: 80}
  ]

  certResult: certCheckResult[] = [];

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardcluster') wizardnewcluster: ClrWizard;
  @ViewChild('wizardaddcluster') wizardaddcluster: ClrWizard;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  readonly endpointPattern = CustomValidatorDirective.endpointPattern;
  readonly CIDRPattern = CustomValidatorDirective.CIDRPattern;
  readonly IPPattern = CustomValidatorDirective.IPPattern;
  basicClusterForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(21), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    providerId: new FormControl('', Validators.required),
    enableContour: new FormControl(true, Validators.required),
    enableTBS: new FormControl(false, Validators.required)
  });
  tkgsClusterParamForm = new FormGroup({
    tkcVersion: new FormControl('', [Validators.required]),
    masterNum: new FormControl('1', [Validators.required]),
    masterClass: new FormControl('', [Validators.required]),
    masterStoragePolicy: new FormControl('', [Validators.required]),
    masterEtcdSize: new FormControl('10'),
    workerNum: new FormControl('1', [Validators.required, Validators.min(0), Validators.max(150)]),
    workerClass: new FormControl('', [Validators.required]),
    workerStoragePolicy: new FormControl('', [Validators.required]),
    workerContainerdSize: new FormControl('50'),
    defaultStorageClass: new FormControl('', [Validators.required]),
    cniType: new FormControl('antrea'),
    serviceCIDR: new FormControl('10.96.0.0/12', [Validators.pattern(this.CIDRPattern)]),
    serviceDomain: new FormControl('cluster.local'),
    podCIDR: new FormControl('192.168.0.0/16', [Validators.pattern(this.CIDRPattern), podCIDRMinNetmaskValidator(23)]),
  });
  tkgmClusterParamForm = new FormGroup({
    tkcVersion: new FormControl('', [Validators.required]),
    masterNum: new FormControl('1', [Validators.required]),
    masterClass: new FormControl('small', [Validators.required]),
    masterCPU: new FormControl(2),
    masterRAM: new FormControl(4),
    masterDisk: new FormControl(20),
    workerNum: new FormControl('1', [Validators.required, Validators.min(1), Validators.max(100)]),
    workerClass: new FormControl('medium', [Validators.required]),
    workerCPU: new FormControl(2),
    workerRAM: new FormControl(4),
    workerDisk: new FormControl(20),
    enableVIP: new FormControl(false),
    masterVIP: new FormControl(''),
    cniType: new FormControl('antrea'),
    serviceCIDR: new FormControl('100.64.0.0/13', [Validators.pattern(this.CIDRPattern)]),
    podCIDR: new FormControl('100.96.0.0/11', [Validators.pattern(this.CIDRPattern), podCIDRMinNetmaskValidator(23)]),
  });
  basicAttachClusterForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(48), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    endpoint: new FormControl('', [Validators.required, Validators.pattern(this.endpointPattern)]),
    caCert: new FormControl(''),
    skipTLS: new FormControl(false),
    token: new FormControl('', Validators.required),
    registry: new FormControl('', Validators.required),
    defaultRegistry: new FormControl('', Validators.required),
    enableGlobalLB: new FormControl(false),
    loadBalancerId: new FormControl(''),
    enableContour: new FormControl(false, Validators.required),
    enableTBS: new FormControl(false, Validators.required)
  });
  caList = new FormArray([]);
  trustCAFrom = new FormGroup({
    trustCA: this.caList
  });
  updateTKGSClusterWorkerForm = new FormGroup({
    workerCount: new FormControl(1, [Validators.required, Validators.min(1), Validators.max(150)])
  });
  updateTKGMClusterMasterForm = new FormGroup({
    masterCount: new FormControl(1, [Validators.required])
  });
  updateTKGSClusterVMClassForm = new FormGroup({
    masterVMClass: new FormControl('', [Validators.required]),
    workerVMClass: new FormControl('', [Validators.required])
  });
  updateTKGMClusterVMClassForm = new FormGroup({
    masterCPU: new FormControl(''),
    masterRAM: new FormControl(''),
    masterDisk: new FormControl(''),
    workerCPU: new FormControl(''),
    workerRAM: new FormControl(''),
    workerDisk: new FormControl(''),
  });
  updateClusterDescForm = new FormGroup({
    description: new FormControl('')
  })
  deleteClusterForm = new FormGroup({
    confirm: new FormControl('', [Validators.required]),
  });

  loading = true;

  constructor(
    public authService: AuthService,
    private clusterService: ClustersService,
    private providerService: ProviderService,
    private glbService: LoadbalancerService,
    private repositoryService: RepositorysService,
    private toast: ToastrService,
    private translate: TranslateService,
    private orgService: OrganizationService,
    private fb: FormBuilder
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  ngOnInit(): void {
  }

  refresh(state?: ClrDatagridStateInterface): void {
    this.loading = true;
    const filters = [];
    filters.push({ key: 'organization', value: [this.orgId] });
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.clusterService.getAllClusters(!!state?.page?.current ? state.page.current - 1 : 0, !!state?.page?.size ? state.page.size : Number(this.pagination.pageSize), filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.allClusters = !!data.data?.Clusters?.edge?.nodes ? data.data.Clusters.edge.nodes : [];
      this.selectedCluster = null;
      this.total = !!data.data?.Clusters?.totalCount ? data.data.Clusters.totalCount : 0;
      this.certResult = checkCerts(this.allClusters);
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false);
  }

  getAllProvider(): void {
    this.providerService.getProviderNames(this.orgId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-PROVIDER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.providerList = !!data.data?.ClusterProviders?.edge?.nodes ? data.data.ClusterProviders.edge.nodes : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-PROVIDER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  getAllLBName(): void {
    this.glbService.getLbNames(this.orgId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-LOADBALANCER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.lbList = !!data.data?.LoadBalancers?.edge?.nodes ? data.data.LoadBalancers.edge.nodes : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-LOADBALANCER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  getALlRegistryName(): void {
    this.repositoryService.getImageRegistrysNames(this.orgId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-REPOSITORY-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.registryList = !!data.data?.ImageRegistries?.edge?.nodes ? data.data.ImageRegistries.edge.nodes : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-REPOSITORY-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  refreshNamespaceInProvider($event: any): void {
    const providerId = $event.target.value;
    for (const provider of this.providerList){
      if(providerId === provider.id){
        this.providerClass = provider.class;
      }
    }
    switch (this.providerClass) {
      case 'TKGS':
        this.providerService.getTKGSClusterProfileByProvider(providerId).subscribe( data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-PROVIDER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.ClusterProvider) {
            this.clusterNamespace = data.data.ClusterProvider.spec.defaultNamespace;
            this.tkcVersionList = data.data.ClusterProvider.spec.clusterReleases;
            this.tkcvmClassList = data.data.ClusterProvider.spec.vmClasses;
            this.noVMclassOnNamespaceAlert = !this.tkcvmClassList;
            for ( const namespaceInfo of data.data.ClusterProvider.spec.namespaces) {
              if (namespaceInfo.name === this.clusterNamespace){
                this.storageClassList = namespaceInfo.storages;
              }
            }
            this.noStorageOnNamespaceAlert = !this.storageClassList;
            // add sorted tkc version list
            this.sortedTkcVersionList = this.tkcVersionList.sort((n1, n2) => {
              if(semver.gt(n1, n2)){return -1;}
              if(semver.lt(n1,n2)){return 1;}
              return 0;
            });
            // remove duplicate item
            for(let i =0;i < this.sortedTkcVersionList.length - 1;i++){
              if(semver.eq(this.sortedTkcVersionList[i], this.sortedTkcVersionList[i+1])){
                this.sortedTkcVersionList.splice(i,1);
              }
            }
          } else {
            this.clusterNamespace = '';
            this.tkcvmClassList = [];
            this.tkcVersionList = [];
            this.storageClassList = [];
            this.noVMclassOnNamespaceAlert = true;
            this.noStorageOnNamespaceAlert = true;
          }
          }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-PROVIDER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
        break;
      case 'TKG2M':
      case 'TKGM':
        this.providerService.getTKGMClusterProfileByProvider(providerId).subscribe( data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-PROVIDER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.ClusterProvider) {
            this.clusterNamespace = data.data.ClusterProvider.spec.defaultNamespace;
            this.tkcVersionList = data.data.ClusterProvider.spec.clusterReleases;
            // add sorted tkc version list
            this.sortedTkcVersionList = this.tkcVersionList.sort((n1, n2) => {
              if(semver.gt(n1, n2)){return -1;}
              if(semver.lt(n1,n2)){return 1;}
              return 0;
            });
            // remove duplicate item
            for(let i =0;i < this.sortedTkcVersionList.length - 1;i++){
              if(semver.eq(this.sortedTkcVersionList[i], this.sortedTkcVersionList[i+1])){
                this.sortedTkcVersionList.splice(i,1);
              }
            }
          } else {
            this.tkcVersionList = [];
            this.clusterNamespace = '';
          }
          }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-PROVIDER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
        break;
      default:
        console.log('not support.')
    }
  }

  setVIPVaildtor(enabled: boolean){
    if(enabled){
      this.tkgmClusterParamForm.controls['masterVIP'].setValidators([Validators.required, Validators.pattern(this.IPPattern)]);
      this.tkgmClusterParamForm.controls['masterVIP'].updateValueAndValidity();
    } else {
      this.tkgmClusterParamForm.controls['masterVIP'].clearValidators();
      this.tkgmClusterParamForm.controls['masterVIP'].updateValueAndValidity();
    }
  }

  // check cluster CIDR is match worker node count
  clusterCIDRAndWorkerNumValidator(providerClass: string): void {
    this.clusterCIDRInvaildAlert = false;
    let networkCIDR;
    let workerNum;
    if(providerClass === 'TKGS'){
      networkCIDR = !!this.tkgsClusterParamForm.value.podCIDR ? this.tkgsClusterParamForm.value.podCIDR : '192.168.0.0/16';
      workerNum = this.tkgsClusterParamForm.value.workerNum;
    } else {
      networkCIDR = !!this.tkgmClusterParamForm.value.podCIDR ? this.tkgmClusterParamForm.value.podCIDR : '100.96.0.0/11';
      workerNum = this.tkgmClusterParamForm.value.workerNum;
    }
    const netCIDRList = networkCIDR.split('/');
    if (netCIDRList.length === 2){
      const maxWorkerNum = Math.pow(2, 24 - Number(netCIDRList[1]))
      this.clusterCIDRInvaildAlert = maxWorkerNum < workerNum;
    }
  }

  showCreateCluster(): void {
    this.getAllProvider();
    this.createClusterOpen = true;
  }

  showAttachCluster(): void {
    this.getAllLBName();
    this.getALlRegistryName();
    this.attachClusterOpen = true;
  }

  setLBVaildtor(enabled: boolean){
    if(enabled){
      this.basicAttachClusterForm.controls['loadBalancerId'].setValidators(Validators.required);
      this.basicAttachClusterForm.controls['loadBalancerId'].updateValueAndValidity();
    } else {
      this.basicAttachClusterForm.controls['loadBalancerId'].clearValidators();
      this.basicAttachClusterForm.controls['loadBalancerId'].updateValueAndValidity();
    }
  }

  deleteCA(index: number): void {
    this.caList.removeAt(index);
  }

  addCA(): void {
    this.caList.push(this.fb.group({
      caCert: new FormControl('', [Validators.required, CAFormatValidator()])
    }));
  }

  private doClusterReset(): void {
    this.wizardnewcluster.reset();
    this.basicClusterForm.reset();
    this.basicClusterForm.patchValue({
      enableContour: true,
      enableTBS: false
    });
    this.setLBVaildtor(false);
    this.tkgsClusterParamForm.reset();
    this.tkgsClusterParamForm.patchValue({
      masterNum: 1,
      masterEtcdSize: 10,
      workerNum: 1,
      workerContainerdSize: 50,
      cniType: 'antrea',
      serviceCIDR: '10.96.0.0/12',
      serviceDomain: 'cluster.local',
      podCIDR: '192.168.0.0/16'
    });
    this.createClusterOpen = false;
    this.updateClusterOpen = false;
    this.providerList = [];
    this.tkcvmClassList = [];
    this.tkcVersionList = [];
    this.sortedTkcVersionList = [];
    this.providerClass = '';
    this.clusterNamespace = '';
    this.noVMclassOnNamespaceAlert = false;
    this.noStorageOnNamespaceAlert = false;
    this.tkgmClusterParamForm.reset();
    this.tkgmClusterParamForm.patchValue({
      masterNum: 1,
      masterClass: 'small',
      masterCPU: 2,
      masterRAM: 4,
      masterDisk: 20,
      workerNum: 1,
      workerClass: 'medium',
      workerCPU: 2,
      workerRAM: 4,
      workerDisk: 20,
      enableVIP: false,
      cniType: 'antrea',
      serviceCIDR: '100.64.0.0/13',
      podCIDR: '100.96.0.0/11'
    })
    this.setVIPVaildtor(false);
  }

  public doClusterFinish(): void {
    // enable web-terminal for all clusters
    const defaultExtensionsEnabled = [{name: 'WEB_TERMINAL'}];
    if(this.basicClusterForm.value.enableContour){
      defaultExtensionsEnabled.push({name: 'CONTOUR'})
    }
    if(this.basicClusterForm.value.enableTBS){
      defaultExtensionsEnabled.push({name: 'TBS'})
    }
    switch (this.providerClass) {
      case 'TKGS':
        const tkgsInput: CreateCluster = {
          name: this.basicClusterForm.value.name,
          description: this.basicClusterForm.value.description,
          version: this.tkgsClusterParamForm.value.tkcVersion,
          spec: {
            tkgs: {
              namespace: this.clusterNamespace,
              cni: this.tkgsClusterParamForm.value.cniType ? this.tkgsClusterParamForm.value.cniType : 'antrea',
              podCIDR: this.tkgsClusterParamForm.value.podCIDR ? this.tkgsClusterParamForm.value.podCIDR : '192.168.0.0/16',
              svcCIDR: this.tkgsClusterParamForm.value.serviceCIDR ? this.tkgsClusterParamForm.value.serviceCIDR : '10.96.0.0/12',
              svcDomain: this.tkgsClusterParamForm.value.serviceDomain ? this.tkgsClusterParamForm.value.serviceDomain : 'cluster.local',
              defStorage: this.tkgsClusterParamForm.value.defaultStorageClass,
              ctlPlaneClass: this.tkgsClusterParamForm.value.masterClass,
              ctlPlaneCount: this.tkgsClusterParamForm.value.masterNum,
              ctlPlaneStorage: this.tkgsClusterParamForm.value.masterStoragePolicy,
              ctlPlaneStorageCapGi: this.tkgsClusterParamForm.value.masterEtcdSize ? this.tkgsClusterParamForm.value.masterEtcdSize : 10,
              dataPlaneClass: this.tkgsClusterParamForm.value.workerClass,
              dataPlaneCount: this.tkgsClusterParamForm.value.workerNum,
              dataPlaneStorage: this.tkgsClusterParamForm.value.workerStoragePolicy,
              dataPlaneStorageCapGi: this.tkgsClusterParamForm.value.workerContainerdSize ? this.tkgsClusterParamForm.value.workerContainerdSize : 50
            }
          },
          extensions: defaultExtensionsEnabled,
          provider: this.basicClusterForm.value.providerId
        };
        this.clusterService.createCluster(tkgsInput).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.CREATE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.CreateCluster){
            this.toast.show(`${data.data.CreateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.CREATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.CREATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.CREATE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          this.doClusterReset();
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
        });
        break;
      case 'TKG2M':
      case 'TKGM':
        const tkgmInput: CreateCluster = {
          name: this.basicClusterForm.value.name,
          description: this.basicClusterForm.value.description,
          version: this.tkgmClusterParamForm.value.tkcVersion,
          spec: this.providerClass === 'TKG2M' ? {
            tkg2m: {
              namespace: this.clusterNamespace,
              osName: 'ubuntu',
              cni: this.tkgmClusterParamForm.value.cniType ? this.tkgmClusterParamForm.value.cniType : 'antrea',
              clusterCIDR: this.tkgmClusterParamForm.value.podCIDR ? this.tkgmClusterParamForm.value.podCIDR : '172.23.0.0/16',
              serviceCIDR: this.tkgmClusterParamForm.value.serviceCIDR ? this.tkgmClusterParamForm.value.serviceCIDR : '172.22.0.0/16',
              ctlPlaneReplicas: this.tkgmClusterParamForm.value.masterNum,
              workerNodePool0Replicas: this.tkgmClusterParamForm.value.workerNum,
            }
          } : {
            tkgm: {
              namespace: this.clusterNamespace,
              cni: this.tkgmClusterParamForm.value.cniType ? this.tkgmClusterParamForm.value.cniType : 'antrea',
              podCIDR: this.tkgmClusterParamForm.value.podCIDR ? this.tkgmClusterParamForm.value.podCIDR : '192.168.0.0/16',
              svcCIDR: this.tkgmClusterParamForm.value.serviceCIDR ? this.tkgmClusterParamForm.value.serviceCIDR : '10.96.0.0/12',
              ctlPlaneCount: this.tkgmClusterParamForm.value.masterNum,
              dataPlaneCount: this.tkgmClusterParamForm.value.workerNum,
            }
          },
          extensions: defaultExtensionsEnabled,
          provider: this.basicClusterForm.value.providerId
        };
        const tkgmClusterSpec: any = this.providerClass === 'TKG2M' ? tkgmInput.spec.tkg2m : tkgmInput.spec.tkgm;
        switch (this.tkgmClusterParamForm.value.masterClass) {
          case 'Custom':
            if (this.providerClass === 'TKG2M') {
              tkgmClusterSpec.ctlPlaneNumCpus = !!this.tkgmClusterParamForm.value.masterCPU ? this.tkgmClusterParamForm.value.masterCPU : 2;
              tkgmClusterSpec.ctlPlaneMemMiB = !!this.tkgmClusterParamForm.value.masterRAM ? this.tkgmClusterParamForm.value.masterRAM*1024 : 4*1024;
              tkgmClusterSpec.ctlPlaneDiskGiB = !!this.tkgmClusterParamForm.value.masterDisk ? this.tkgmClusterParamForm.value.masterDisk : 20;
            } else {
              tkgmClusterSpec.ctlPlaneCPUs = !!this.tkgmClusterParamForm.value.masterCPU ? this.tkgmClusterParamForm.value.masterCPU : 2;
              tkgmClusterSpec.ctlPlaneMemMi = !!this.tkgmClusterParamForm.value.masterRAM ? this.tkgmClusterParamForm.value.masterRAM*1024 : 4*1024;
              tkgmClusterSpec.ctlPlaneStorageCapGi = !!this.tkgmClusterParamForm.value.masterDisk ? this.tkgmClusterParamForm.value.masterDisk : 20;
            }
            break;
          default:
            const masterIndex =  this.tkgmNodeType.findIndex(x => x.name === this.tkgmClusterParamForm.value.masterClass)
            if (this.providerClass === 'TKG2M') {
              tkgmClusterSpec.ctlPlaneNumCpus = this.tkgmNodeType[masterIndex].cpu;
              tkgmClusterSpec.ctlPlaneMemMiB = this.tkgmNodeType[masterIndex].mem*1024;
              tkgmClusterSpec.ctlPlaneDiskGiB = this.tkgmNodeType[masterIndex].storage;
            } else {
              tkgmClusterSpec.ctlPlaneCPUs = this.tkgmNodeType[masterIndex].cpu;
              tkgmClusterSpec.ctlPlaneMemMi = this.tkgmNodeType[masterIndex].mem*1024;
              tkgmClusterSpec.ctlPlaneStorageCapGi = this.tkgmNodeType[masterIndex].storage;
            }
            break;
        }
        switch (this.tkgmClusterParamForm.value.workerClass) {
          case 'Custom':
            if (this.providerClass === 'TKG2M') {
              tkgmClusterSpec.workerNumCpus = !!this.tkgmClusterParamForm.value.workerCPU ? this.tkgmClusterParamForm.value.workerCPU : 2;
              tkgmClusterSpec.workerMemMiB = !!this.tkgmClusterParamForm.value.workerRAM ? this.tkgmClusterParamForm.value.workerRAM*1024 : 4*1024;
              tkgmClusterSpec.workerDiskGiB = !!this.tkgmClusterParamForm.value.workerDisk ? this.tkgmClusterParamForm.value.workerDisk : 20;
            } else {
              tkgmClusterSpec.dataPlaneCPUs = !!this.tkgmClusterParamForm.value.workerCPU ? this.tkgmClusterParamForm.value.workerCPU : 2;
              tkgmClusterSpec.dataPlaneMemMi = !!this.tkgmClusterParamForm.value.workerRAM ? this.tkgmClusterParamForm.value.workerRAM*1024 : 4*1024;
              tkgmClusterSpec.dataPlaneStorageCapGi = !!this.tkgmClusterParamForm.value.workerDisk ? this.tkgmClusterParamForm.value.workerDisk : 20;
            }
            break;
          default:
            const nodeIndex =  this.tkgmNodeType.findIndex(x => x.name === this.tkgmClusterParamForm.value.workerClass)
            if (this.providerClass === 'TKG2M') {
              tkgmClusterSpec.workerNumCpus = this.tkgmNodeType[nodeIndex].cpu;
              tkgmClusterSpec.workerMemMiB = this.tkgmNodeType[nodeIndex].mem*1024;
              tkgmClusterSpec.workerDiskGiB = this.tkgmNodeType[nodeIndex].storage;
            } else {
              tkgmClusterSpec.dataPlaneCPUs = this.tkgmNodeType[nodeIndex].cpu;
              tkgmClusterSpec.dataPlaneMemMi = this.tkgmNodeType[nodeIndex].mem*1024;
              tkgmClusterSpec.dataPlaneStorageCapGi = this.tkgmNodeType[nodeIndex].storage;
            }
            break;
        }
        if (this.providerClass === 'TKGM' && this.tkgmClusterParamForm.value.enableVIP){
          const paramsMap: JSON = JSON.parse(JSON.stringify({}));
          paramsMap['AVI_CONTROL_PLANE_HA_PROVIDER'] = 'false';
          paramsMap['VSPHERE_CONTROL_PLANE_ENDPOINT'] = this.tkgmClusterParamForm.value.masterVIP;
          tkgmClusterSpec.env = paramsMap;
        }
        this.clusterService.createCluster(tkgmInput).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.CREATE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.CreateCluster){
            this.toast.show(`${data.data.CreateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.CREATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.CREATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.CREATE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          this.doClusterReset();
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
        });
        break;
      default:
        console.log('not support.')
    }
  }

  public doClusterCancel(): void {
    this.doClusterReset();
  }

  getSelectedRegistry(): ImageRegistry[] {
    return this.basicAttachClusterForm.value.registry;
  }

  private doAttachClusterReset(): void {
    this.wizardaddcluster.reset();
    this.basicAttachClusterForm.reset();
    this.basicAttachClusterForm.patchValue({
      skipTLS: false,
      enableGlobalLB: false,
      enableContour: false,
      enableTBS: false
    })
    while (this.caList.length !== 0) {
      this.caList.removeAt(0);
    }
    this.trustCAFrom.reset();
    this.attachClusterOpen = false;
    this.ClusterConnectSuccess = false;
    this.showClusterConnectAlert = false;
    this.registryList = [];
    this.lbList = [];
  }

  public doAttachClusterFinish(): void {
    const defaultExtensionsEnabled = [{name: 'WEB_TERMINAL'}];
    if(this.basicAttachClusterForm.value.enableContour){
      defaultExtensionsEnabled.push({name: 'CONTOUR'})
    }
    if(this.basicAttachClusterForm.value.enableTBS){
      defaultExtensionsEnabled.push({name: 'TBS'})
    }
    const addImageRegistry: AddClusterProviderImageRegistry[] = [];
    for(let registry of this.basicAttachClusterForm.value.registry){
      if(registry.id !== this.basicAttachClusterForm.value.defaultRegistry){
        addImageRegistry.push({
          id: registry.id,
          default: false
        })
      } else {
        addImageRegistry.push({
          id: registry.id,
          default: true
        })
      }
    }
    const attachParam: AttachCluster = {
      name: this.basicAttachClusterForm.value.name,
      description: this.basicAttachClusterForm.value.description,
      endpoint: this.basicAttachClusterForm.value.endpoint,
      caCert: !this.basicAttachClusterForm.value.skipTLS ? this.basicAttachClusterForm.value.caCert : null,
      insecureSkipTLSVerify: this.basicAttachClusterForm.value.skipTLS,
      authToken: this.basicAttachClusterForm.value.token,
      spec: {
        rootCAs: []
      },
      extensions: defaultExtensionsEnabled,
      imageRegistries: addImageRegistry,
      loadBalancer: this.basicAttachClusterForm.value.enableGlobalLB ? this.basicAttachClusterForm.value.loadBalancerId : null
    }
    if(this.basicAttachClusterForm.value.enableTBS && !!this.trustCAFrom.value.trustCA){
      for(const caTrust of this.trustCAFrom.value.trustCA){
        attachParam.spec.rootCAs.push(caTrust.caCert)
      }
    }
    this.clusterService.attachCluster(attachParam).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.ADDERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.ATTACH-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.AttachCluster) {
          this.toast.show(`${data.data.AttachCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.ATTACH-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.ATTACH-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.ATTACH-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doAttachClusterReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }


  public doAttachClusterCancel(): void {
    this.doAttachClusterReset();
  }

  public checkAttachClusterConnect(server: string, token: string, registryId: string): void {
    this.clusterService.checkAttachClusterToken(server, token, registryId).subscribe(
      data => {
        if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.CHECKERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.CHECK-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
        }
        this.ClusterConnectSuccess = !!data.data?.AttachCluster;
        this.showClusterConnectAlert = true;
      }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.CHECK-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
    );
  }


  showDeleteClusterDialog(cluster: Cluster): void {
    this.deleteClusterDialog = true;
    this.selectedClusterName = cluster.name;
    this.selectedClusterId = cluster.id;
  }

  deleteClusterById(id: string): void {
    this.deleteClusterDialog = false;
    this.clusterService.deleteCluster(id).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             if(each.message === 'object is busy'){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-ALERT'), this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             } else {
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
           }
        }
      }
      if (!!data.data?.DeleteCluster){
          this.toast.show(`${data.data.DeleteCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  showRemoveClusterDialog(cluster: Cluster): void {
    this.removeClusterDialog = true;
    this.selectedClusterName = cluster.name;
    this.selectedClusterId = cluster.id;
  }

  removeClusterById(id: string): void {
    this.removeClusterDialog = false;
    this.clusterService.deleteCluster(id).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             if(each.message === 'object is busy'){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-ALERT'), this.translate.instant('MESSAGE.CLUSTER.REMOVE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             } else {
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.REMOVE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
           }
        }
      }
      if (!!data.data?.DeleteCluster){
          this.toast.show(`${data.data.DeleteCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.REMOVE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.REMOVE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.REMOVE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  showForceDeleteDialog(cluster: Cluster): void {
    this.forceDeleteDialog = true;
    this.selectedClusterName = cluster.name;
    this.selectedClusterId = cluster.id;
  }

  forceDeleteClusterById(id: string): void {
    this.forceDeleteDialog = false;
    this.clusterService.forceDeleteCluster(id).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             if(each.message === 'object is busy'){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-ALERT'), this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             } else {
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
           }
        }
      }
      if (!!data.data?.DeleteCluster){
          this.toast.show(`${data.data.DeleteCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doReinitCluster(cluster: Cluster): void {
    this.clusterService.ReinitCluster(cluster.id).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-REINIT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.UpdateCluster){
          this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-REINIT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
     })
  }

  showUpdateClusterDesc(cluster: Cluster): void {
    this.updateClusterDescForm.setValue({
      description: cluster.description,
    });
    this.selectedClusterId = cluster.id;
    this.showUpdateClusterDescDialog = true;
  }

  doUpdateClusterDescription(): void {
    this.showUpdateClusterDescDialog = false;
    this.clusterService.updateClusterDesc(this.selectedClusterId, this.updateClusterDescForm.value.description).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-DESC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.UpdateCluster){
          this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-DESC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
     })
  }

  showUpdateClusterVMClass(cluster: Cluster): void {
    this.selectedCluster = cluster;
    switch (this.selectedCluster?.class) {
      case 'TKGS':
        this.showUpdateTKGSVMclassDialog = true;
        this.updateTKGSVMclassChanged = false;
        this.tkcvmClassList = cluster.provider.spec.vmClasses;
        this.updateTKGSClusterVMClassForm.patchValue({
          masterVMClass: cluster.spec.ctlPlaneClass,
          workerVMClass: cluster.spec.dataPlaneClass
        });
        break;
      case 'TKG2M':
      case 'TKGM':
        this.showUpdateTKGMVMclassDialog = true;
        this.updateTKGMVMclassChanged = false;
        this.updateTKGMClusterVMClassForm.patchValue({
          masterCPU: cluster.class === 'TKG2M' ? cluster.spec.ctlPlaneNumCpus : cluster.spec.ctlPlaneCPUs,
          masterRAM: cluster.class === 'TKG2M' ? cluster.spec.ctlPlaneMemMiB / 1024 : cluster.spec.ctlPlaneMemMi / 1024,
          masterDisk: cluster.class === 'TKG2M' ? cluster.spec.ctlPlaneDiskGiB : cluster.spec.ctlPlaneStorageCapGi,
          workerCPU: cluster.class === 'TKG2M' ? cluster.spec.workerNumCpus : cluster.spec.dataPlaneCPUs,
          workerRAM: cluster.class === 'TKG2M' ? cluster.spec.workerMemMiB / 1024 : cluster.spec.dataPlaneMemMi / 1024,
          workerDisk: cluster.class === 'TKG2M' ? cluster.spec.workerDiskGiB : cluster.spec.dataPlaneStorageCapGi
        })
        break;
      default:
        console.log('not support')
    }
  }

  doUpdateClusterVMClass(): void {
    switch (this.selectedCluster?.class) {
      case 'TKGS': {
        this.tkcvmClassList = [];
        this.showUpdateTKGSVMclassDialog = false;
        let input: UpdateCluster = {
          id: this.selectedCluster.id,
          spec: {
            tkgs: {
              ctlPlaneClass: this.updateTKGSClusterVMClassForm.value.masterVMClass,
              dataPlaneClass: this.updateTKGSClusterVMClassForm.value.workerVMClass
            }
          }
        };
        this.clusterService.updateClusterVMClass(input).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-VMCLASS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.UpdateCluster){
              this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-VMCLASS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
         })
        break;
      }
      case 'TKG2M':
      case 'TKGM': {
        this.showUpdateTKGMVMclassDialog = false;
        const oldCtlResourceSetting = this.selectedCluster.class === 'TKG2M'
          ? this.selectedCluster.spec.ctlPlaneNumCpus + this.selectedCluster.spec.ctlPlaneMemMiB / 1024 + this.selectedCluster.spec.ctlPlaneDiskGiB
          : this.selectedCluster.spec.ctlPlaneCPUs + this.selectedCluster.spec.ctlPlaneMemMi / 1024 + this.selectedCluster.spec.ctlPlaneStorageCapGi;
        const oldDataResourceSetting = this.selectedCluster.class === 'TKG2M'
          ? this.selectedCluster.spec.workerNumCpus + this.selectedCluster.spec.workerMemMiB / 1024 + this.selectedCluster.spec.workerDiskGiB
          : this.selectedCluster.spec.dataPlaneCPUs + this.selectedCluster.spec.dataPlaneMemMi / 1024 + this.selectedCluster.spec.dataPlaneStorageCapGi;
        const newCtlResourceSetting = this.updateTKGMClusterVMClassForm.value.masterCPU + this.updateTKGMClusterVMClassForm.value.masterRAM + this.updateTKGMClusterVMClassForm.value.masterDisk;
        const newDataResourceSetting = this.updateTKGMClusterVMClassForm.value.workerCPU + this.updateTKGMClusterVMClassForm.value.workerRAM + this.updateTKGMClusterVMClassForm.value.workerDisk;
        let input: UpdateCluster = {
          id: this.selectedCluster.id,
          spec: this.selectedCluster.class === 'TKG2M' ? {
            tkg2m: {
              ctlPlaneNumCpus: oldCtlResourceSetting !== newCtlResourceSetting ? this.updateTKGMClusterVMClassForm.value.masterCPU : null,
              ctlPlaneMemMiB: oldCtlResourceSetting !== newCtlResourceSetting ? this.updateTKGMClusterVMClassForm.value.masterRAM * 1024 : null,
              ctlPlaneDiskGiB: oldCtlResourceSetting !== newCtlResourceSetting ? this.updateTKGMClusterVMClassForm.value.masterDisk : null,
              workerNumCpus: oldDataResourceSetting != newDataResourceSetting ? this.updateTKGMClusterVMClassForm.value.workerCPU : null,
              workerMemMiB: oldDataResourceSetting != newDataResourceSetting ? this.updateTKGMClusterVMClassForm.value.workerRAM * 1024 : null,
              workerDiskGiB: oldDataResourceSetting != newDataResourceSetting ? this.updateTKGMClusterVMClassForm.value.workerDisk : null
            }
          } : {
            tkgm: {
              ctlPlaneCPUs: oldCtlResourceSetting !== newCtlResourceSetting ? this.updateTKGMClusterVMClassForm.value.masterCPU : null,
              ctlPlaneMemMi: oldCtlResourceSetting !== newCtlResourceSetting ? this.updateTKGMClusterVMClassForm.value.masterRAM * 1024 : null,
              ctlPlaneStorageCapGi: oldCtlResourceSetting !== newCtlResourceSetting ? this.updateTKGMClusterVMClassForm.value.masterDisk : null,
              dataPlaneCPUs: oldDataResourceSetting != newDataResourceSetting ? this.updateTKGMClusterVMClassForm.value.workerCPU : null,
              dataPlaneMemMi: oldDataResourceSetting != newDataResourceSetting ? this.updateTKGMClusterVMClassForm.value.workerRAM * 1024 : null,
              dataPlaneStorageCapGi: oldDataResourceSetting != newDataResourceSetting ? this.updateTKGMClusterVMClassForm.value.workerDisk : null
            }
          }
        };
        if (oldCtlResourceSetting !== newCtlResourceSetting || oldDataResourceSetting != newDataResourceSetting){
          this.clusterService.updateClusterVMClass(input).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-VMCLASS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.UpdateCluster){
              this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-VMCLASS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
         })
        } else {
          this.toast.show(this.translate.instant('MESSAGE.CLUSTER.UPDATE-RESOURCE-NOCHANGE'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
        }
        break;
      }
      default:
        console.log('not support.')
    }
  }

  showUpdateClustrMaster(cluster: Cluster): void {
    this.allowUpdateClusterMaster = false;
    this.selectedClusterId = cluster.id;
    this.selectedClusterName = cluster.name;
    this.providerClass = cluster.class;
    switch (this.providerClass) {
      case 'TKGS':
        this.showUpdateTKGSMasterDialog = true;
        if (cluster.spec.ctlPlaneCount === 1) {
          this.allowUpdateClusterMaster = true;
        }
        break;
      case 'TKG2M':
      case 'TKGM':
        this.showUpdateTKGMMasterDialog = true;
        this.updateTKGMClusterMasterForm.setValue({
          masterCount: cluster.class === 'TKG2M' ? cluster.spec.ctlPlaneReplicas : cluster.spec.ctlPlaneCount
        })
        break;
      default:
        console.log('not support.')
        break;
    }
  }

  doUpdateClusterMaster(id: string): void {
    switch (this.providerClass) {
      case 'TKGS':
        this.showUpdateTKGSMasterDialog = false;
        this.clusterService.updateTKGSClusterMasterCount(id).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSATER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.UpdateCluster){
              this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSATER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
         })
        break;
      case 'TKG2M':
      case 'TKGM':
        this.showUpdateTKGMMasterDialog = false;
        const updateMaster = this.providerClass === 'TKG2M'
          ? this.clusterService.updateTKG2MClusterMasterCount(id, this.updateTKGMClusterMasterForm.value.masterCount)
          : this.clusterService.updateTKGMClusterMasterCount(id, this.updateTKGMClusterMasterForm.value.masterCount);
        updateMaster.subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSATER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.UpdateCluster){
              this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSATER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
         })
        break;
      default:
        console.log('not support')
    }
  }

  showUpdateClusterWorker(cluster: Cluster): void {
    this.showUpdateWorkerDialog = true;
    this.updateWorkerChanged = false;
    this.selectedClusterName = cluster.name;
    this.selectedClusterId = cluster.id;
    this.providerClass = cluster.class;
    this.updateTKGSClusterWorkerForm.patchValue({
      workerCount: cluster.class === 'TKG2M' ? cluster.spec.workerNodePool0Replicas : cluster.spec.dataPlaneCount
    });
  }

  doUpdateClusterWorker(id: string, count: number): void {
    this.showUpdateWorkerDialog = false;
    switch (this.providerClass) {
      case 'TKGS':
         this.clusterService.updateTKGSClusterWorkerCount(id, count).subscribe(data => {
           if (!!data.errors){
              for (const each of data.errors){
                 if(!!each.message){
                   this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-WORKER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
                 }
              }
            }
          if (!!data.data?.UpdateCluster){
              this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-WORKER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
         })
        break;
      case 'TKG2M':
      case 'TKGM':
         const updateWorker = this.providerClass === 'TKG2M'
           ? this.clusterService.updateTKG2MClusterWorkerCount(id, count)
           : this.clusterService.updateTKGMClusterWorkerCount(id, count);
         updateWorker.subscribe(data => {
           if (!!data.errors){
              for (const each of data.errors){
                 if(!!each.message){
                   this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-WORKER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
                 }
              }
            }
          if (!!data.data?.UpdateCluster){
              this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-WORKER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
         })
        break;
      default:
        console.log('not support')
    }
  }
}
