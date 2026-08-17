import {Component, OnInit, ViewChild} from '@angular/core';
import {AuthService} from "../../../services/auth.service";
import {ClustersService} from "../../../services/clusters.service";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {OrganizationService} from "../../../services/organization.service";
import {Cluster, handleObject, ClusterExtension, EnableClusterExtension, ModifyClusterExtension, CreateShortcut, ClusterExtensionState, ClusterExtensionName, UpdateShortcut} from "../../../classes";
import {FormControl, FormGroup, Validators} from "@angular/forms";
import {ClrWizard} from "@clr/angular";
import { environment} from '../../../../environments/environment';
import {ClusterExtensionService} from "../../../services/cluster-extension.service";
import {CustomValidatorDirective, nameValidator} from "../../../utils/validators/custom-validator.directive";



@Component({
  selector: 'app-operation',
  templateUrl: './operation.component.html',
  styleUrls: ['./operation.component.css']
})
export class OperationComponent implements OnInit {

  //admin role can handle cluster create/update/delete
  get canWrite(): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER);
  }

  constructor(
    public authService: AuthService,
    private clusterService: ClustersService,
    private toast: ToastrService,
    private translate: TranslateService,
    private orgService: OrganizationService,
    private clusterExtensionsService: ClusterExtensionService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  backgroupImage = environment.assetsPath + 'images/tanzu-back.png';
  imagePath: string = environment.assetsPath + 'images';

  orgId: string;
  selectedClusterName: string;
  selectedExtensionName: string;
  selectedExtensionId: string;
  currentPrometheusStorage: number;
  currentElasticsearchStorage: number;
  sharedClusterList: Cluster[] = [];
  showSearch = false;
  searchName: string;
  showExtensionOpen = false;

  allExtension: ClusterExtension[] = [];
  selectedExtension: ClusterExtension;
  extensionEnabledClusters: Cluster[] = [];
  extensionDisableClusters: Cluster[] = [];
  loading = false;

  showEditOctant = false;
  showEditGrafana = false;
  showEditPrometheus = false;
  showEditElasticsearch = false;
  showEditFluentbit = false;
  removeRelateClusterList: string[] = [];
  removeDialog = false;
  addOperationOpen = false;
  showEditOtherService = false;

  @ViewChild('wizardoperation') wizardnewoperation: ClrWizard;
  serviceType: string;
  octantParamForm = new FormGroup({
    cluster: new FormControl('', Validators.required),
    sharedCluster: new FormControl('')
  });
  updateOctantRelatedForm = new FormGroup({
    sharedCluster: new FormControl('')
  })
  prometheusParamForm = new FormGroup({
    cluster: new FormControl('', Validators.required),
    storage: new FormControl(256, Validators.min(20))
  });
  updatePrometheusParamForm = new FormGroup({
    storage: new FormControl(256, Validators.min(20))
  });
  grafanaParamForm = new FormGroup({
    cluster: new FormControl('', Validators.required),
    prometheusCluster: new FormControl(''),
    password: new FormControl(''),
    confirmPassword: new FormControl('')
  });
  updateGrafanaRelatedForm = new FormGroup({
    prometheusCluster: new FormControl('')
  })
  esParamForm = new FormGroup({
    cluster: new FormControl('', Validators.required),
    esDataNodeCount: new FormControl(1, Validators.min(1)),
    esRequestCPU: new FormControl(1, Validators.min(1)),
    esRequestMEM: new FormControl(2, Validators.min(2)),
    esDataVolume: new FormControl(50, Validators.min(10)),
    esLogsRetainDays: new FormControl(7, Validators.min(1)),
  });
  updateESParamForm = new FormGroup({
    esDataNodeCount: new FormControl(1, Validators.min(1)),
    esRequestCPU: new FormControl(1, Validators.min(1)),
    esRequestMEM: new FormControl(2, Validators.min(2)),
    esDataVolume: new FormControl(50, Validators.min(10)),
    esLogsRetainDays: new FormControl(7, Validators.min(1)),
  });
  fluentbitParamForm = new FormGroup({
    cluster: new FormControl('', Validators.required),
    esCluster: new FormControl(''),
    enableAudit: new FormControl(false)
  })
  updateFluentbitParamForm = new FormGroup({
    esCluster: new FormControl(''),
    enableAudit: new FormControl(false)
  })
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  readonly broadEndpointPattern = CustomValidatorDirective.broadEndpointPattern;
  otherParamForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    endpoint: new FormControl('', [Validators.required, Validators.pattern(this.broadEndpointPattern)])
  })
  updateOtherParamForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    endpoint: new FormControl('', [Validators.required, Validators.pattern(this.broadEndpointPattern)])
  })
  deleteOperationForm = new FormGroup({
    confirm: new FormControl('', Validators.required)
  })

  ngOnInit(): void {
    this.refresh();
  }

  refresh(): void {
    this.loading = true;
    const filters = [{ key: 'organization', value: [this.orgId] }];
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    } else {
      filters.push({ key: 'name', value: ['OCTANT', 'PROMETHEUS', 'GRAFANA', 'ELASTICSEARCH', 'FLUENTBIT'] })
    }
    this.clusterExtensionsService.getAllClusterExtension(filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.GET-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      const tempList = !!data.data?.ClusterExtensions?.edge?.nodes ? data.data.ClusterExtensions.edge.nodes : [];
      this.allExtension = tempList.sort((n1, n2) => {
          if(n1.name > n2.name){return 1;}
          if(n1.name < n2.name){return -1}
          return 0
      });
      this.clusterExtensionsService.getAllShotcuts(this.orgId).subscribe( data => {
        if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.GET-SHOTCUT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
        }
        if(!!data.data?.Shortcuts?.edge?.nodes){
          for (let shotcut of data.data.Shortcuts.edge.nodes){
            this.allExtension.push({
              id: shotcut.id,
              name: ClusterExtensionName.OTHER,
              endpoints: [shotcut.endpoint],
              state: ClusterExtensionState.ENABLED,
              cluster: {
                name: shotcut.name,
                description: shotcut.description
              }
            })
          }
        }
      }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.GET-SHOTCUT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {this.loading = false;})
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.GET-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
  }

  getAllHealthyCluster(): void {
    this.clusterService.getAllHealthyClusterId(this.orgId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.sharedClusterList = !!data.data?.Clusters?.edge?.nodes ? data.data.Clusters.edge.nodes : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
  }

  getDisableExtensionCluster(extensionName: string): void {
    this.clusterService.getExtensionDisabledClusterId(this.orgId, extensionName).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.extensionDisableClusters = !!data.data?.Clusters?.edge?.nodes ? data.data.Clusters.edge.nodes : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
  }

  getEnableExtensionCluster(extensionName: string): void {
    this.clusterService.getExtensionEnabledClusterId(this.orgId, extensionName).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.extensionEnabledClusters = !!data.data?.Clusters?.edge?.nodes ? data.data.Clusters.edge.nodes : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
  }

  showCreateOperator(): void {
    this.addOperationOpen = true;
  }

  setServiceType(type: string): void {
    this.serviceType = type;
    //get all disable clusters
    this.extensionDisableClusters = null;
    this.getDisableExtensionCluster(type);
    switch (type) {
      case 'OCTANT':
        this.getAllHealthyCluster();
        break;
      case 'GRAFANA':
        this.getEnableExtensionCluster('PROMETHEUS');
        break;
      case 'FLUENTBIT':
        this.getEnableExtensionCluster('ELASTICSEARCH');
        break;
    }
  }

  checkPasswdValidation(pass: string, repass: string): boolean {
    return pass === repass;
  }

  doOperationFinish(): void {
    switch (this.serviceType) {
      case 'OCTANT':
        const relateClusterList: string[] = [];
        for(let cluster of this.octantParamForm.value.sharedCluster){
          relateClusterList.push(cluster.id);
        }
        const octantParam: EnableClusterExtension = {
          name: 'OCTANT',
          related: relateClusterList,
          cluster: this.octantParamForm.value.cluster
        }
        this.clusterExtensionsService.enableClusterExtensions(octantParam).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.EnableClusterExtension){
            this.toast.show(`${data.data.EnableClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-MSG'), this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
          this.doOperationReset();
          this.refresh();
         })
        break;
      case 'PROMETHEUS':
        const prometheusParamsMap: JSON = JSON.parse(JSON.stringify({}));
        prometheusParamsMap['PROMETHEUS_DATASTORE_SIZE_GIB'] = !!this.prometheusParamForm.value.storage ? this.prometheusParamForm.value.storage+'' : '256';
        const prometheusParam: EnableClusterExtension = {
          name: 'PROMETHEUS',
          cluster: this.prometheusParamForm.value.cluster,
          spec: prometheusParamsMap
        }
        this.clusterExtensionsService.enableClusterExtensions(prometheusParam).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.EnableClusterExtension){
            this.toast.show(`${data.data.EnableClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-MSG'), this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
          this.doOperationReset();
          this.refresh();
         })
        break;
      case 'GRAFANA':
        const grafanaParamsMap: JSON = JSON.parse(JSON.stringify({}));
        grafanaParamsMap['GRAFANA_ADMIN_PASSWORD'] = !!this.grafanaParamForm.value.confirmPassword ? this.grafanaParamForm.value.confirmPassword : 'vmware';
        const relatePrometheusClusterList: string[] = [];
        if (!!this.grafanaParamForm.value.prometheusCluster){
          for(let cluster of this.grafanaParamForm.value.prometheusCluster){
            relatePrometheusClusterList.push(cluster.id);
          }
        }
        const grafanaParam: EnableClusterExtension = {
          name: 'GRAFANA',
          related: relatePrometheusClusterList,
          spec: grafanaParamsMap,
          cluster: this.grafanaParamForm.value.cluster
        }
        this.clusterExtensionsService.enableClusterExtensions(grafanaParam).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.EnableClusterExtension){
            this.toast.show(`${data.data.EnableClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-MSG'), this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
          this.doOperationReset();
          this.refresh();
         })
        break;
      case 'FLUENTBIT':
        const fluentbitParamsMap: JSON = JSON.parse(JSON.stringify({}));
        fluentbitParamsMap['FLUENTBIT_ENABLE_AUDIT_LOGGING'] = this.fluentbitParamForm.value.enableAudit+'';
        const fluentbitParam: EnableClusterExtension = {
          name: 'FLUENTBIT',
          related: !!this.fluentbitParamForm.value.esCluster ? [this.fluentbitParamForm.value.esCluster] : null,
          cluster: this.fluentbitParamForm.value.cluster,
          spec: fluentbitParamsMap
        }
        this.clusterExtensionsService.enableClusterExtensions(fluentbitParam).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.EnableClusterExtension){
            this.toast.show(`${data.data.EnableClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-MSG'), this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
          this.doOperationReset();
          this.refresh();
         })
        break;
      case 'ELASTICSEARCH':
        const esParamsMap: JSON = JSON.parse(JSON.stringify({}));
        esParamsMap['ES_INSTANCE_REPLICAS'] = !!this.esParamForm.value.esDataNodeCount ? this.esParamForm.value.esDataNodeCount+'' : '1';
        esParamsMap['ES_DATA_HEAPSIZE_GIB'] = !!this.esParamForm.value.esRequestMEM ? Math.round(this.esParamForm.value.esRequestMEM/2)+'' : '1';
        esParamsMap['ES_RESOURCE_LIMIT_CPU'] = !!this.esParamForm.value.esRequestCPU ? this.esParamForm.value.esRequestCPU*2+'' : '2';
        esParamsMap['ES_RESOURCE_LIMIT_MEMORY_GIB'] = !!this.esParamForm.value.esRequestMEM ? this.esParamForm.value.esRequestMEM*2+'' : '4';
        esParamsMap['ES_RESOURCE_REQUEST_CPU'] = !!this.esParamForm.value.esRequestCPU ? this.esParamForm.value.esRequestCPU+'' : '1';
        esParamsMap['ES_RESOURCE_REQUEST_MEMORY_GIB'] = !!this.esParamForm.value.esRequestMEM ? this.esParamForm.value.esRequestMEM+'' : '2';
        esParamsMap['ES_DATASTORE_SIZE_GIB'] = !!this.esParamForm.value.esDataVolume ? this.esParamForm.value.esDataVolume+'' : '50';
        esParamsMap['ES_PRESERVED_UNIT_COUNT'] = !!this.esParamForm.value.esLogsRetainDays ? this.esParamForm.value.esLogsRetainDays+'' : '7';
        const esParam: EnableClusterExtension = {
          name: 'ELASTICSEARCH',
          cluster: this.esParamForm.value.cluster,
          spec: esParamsMap
        }
        this.clusterExtensionsService.enableClusterExtensions(esParam).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.EnableClusterExtension){
            this.toast.show(`${data.data.EnableClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-MSG'), this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
          this.doOperationReset();
          this.refresh();
         })
        break;
      case 'OTHER':
        const otherParam: CreateShortcut = {
          name: this.otherParamForm.value.name,
          description: this.otherParamForm.value.description,
          endpoint: this.otherParamForm.value.endpoint,
          organization: this.orgId
        }
        this.clusterExtensionsService.createShortcut(otherParam).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-SHOTCUT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.CreateShortcut){
            this.toast.show(`${data.data.CreateShortcut.name}`+this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-SHOTCUT-MSG'), this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-SHOTCUT-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.CREATE-SHOTCUT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
          this.doOperationReset();
          this.refresh();
         })
        break;
    }
  }

  doOperationCancel(): void {
    this.doOperationReset();
  }

  private doOperationReset(): void {
    this.addOperationOpen = false;
    this.wizardnewoperation.reset();
    this.serviceType = '';
    this.sharedClusterList = [];
    this.extensionDisableClusters = [];
    this.extensionEnabledClusters = [];
    this.octantParamForm.reset();
    this.prometheusParamForm.reset();
    this.prometheusParamForm.patchValue({
      storage: 256
    })
    this.grafanaParamForm.reset();
    this.fluentbitParamForm.reset();
    this.fluentbitParamForm.patchValue({
      enableAudit: false
    })
    this.esParamForm.reset();
    this.esParamForm.patchValue({
      esDataNodeCount: 1,
      esRequestCPU: 1,
      esRequestMEM: 2,
      esDataVolume: 50,
      esLogsRetainDays: 7
    })
    this.otherParamForm.reset();
  }

  showEditExtensionService(extension: ClusterExtension): void {
    this.selectedClusterName = extension.cluster.name;
    this.selectedExtensionId = extension.id;
    switch (extension.name) {
      case 'OCTANT':
        this.getAllHealthyCluster();
        this.updateOctantRelatedForm.setValue({
          sharedCluster: extension.related.edge.nodes
        })
        if(!!extension?.related?.edge?.nodes){
          for(let cluster of extension.related.edge.nodes){
            this.removeRelateClusterList.push(cluster.id);
          }
        }
        this.showEditOctant = true;
        break;
      case 'GRAFANA':
        this.getEnableExtensionCluster('PROMETHEUS');
        this.updateGrafanaRelatedForm.setValue({
          prometheusCluster: extension.related.edge.nodes
        })
        if(!!extension?.related?.edge?.nodes){
          for(let cluster of extension.related.edge.nodes){
            this.removeRelateClusterList.push(cluster.id);
          }
        }
        this.showEditGrafana = true;
        break;
      case 'PROMETHEUS':
        for (const key of Object.keys(extension.spec)) {
          if (key === 'PROMETHEUS_DATASTORE_SIZE_GIB') {
            this.currentPrometheusStorage = Number(extension.spec[key]);
            this.updatePrometheusParamForm.patchValue({
              storage: Number(extension.spec[key])
            })
          }
        }
        this.showEditPrometheus = true;
        break;
      case 'ELASTICSEARCH':
        for (const key of Object.keys(extension.spec)) {
          if (key === 'ES_INSTANCE_REPLICAS') {
            this.updateESParamForm.patchValue({
              esDataNodeCount: Number(extension.spec[key])
            })
          } else if (key === 'ES_RESOURCE_REQUEST_CPU'){
            this.updateESParamForm.patchValue({
              esRequestCPU: Number(extension.spec[key])
            })
          } else if (key === 'ES_RESOURCE_REQUEST_MEMORY_GIB'){
            this.updateESParamForm.patchValue({
              esRequestMEM: Number(extension.spec[key])
            })
          } else if (key === 'ES_PRESERVED_UNIT_COUNT'){
            this.updateESParamForm.patchValue({
              esLogsRetainDays: Number(extension.spec[key])
            })
          } else if (key === 'ES_DATASTORE_SIZE_GIB') {
            this.currentElasticsearchStorage = Number(extension.spec[key]);
            this.updateESParamForm.patchValue({
              esDataVolume: Number(extension.spec[key])
            })
          }
        }
        this.showEditElasticsearch = true;
        break;
      case 'FLUENTBIT':
        this.getEnableExtensionCluster('ELASTICSEARCH');
        for (const key of Object.keys(extension.spec)) {
          if (key === 'FLUENTBIT_ENABLE_AUDIT_LOGGING') {
            this.updateFluentbitParamForm.patchValue({
              enableAudit: extension.spec[key] === 'true'
            })
          }
        }
        this.updateFluentbitParamForm.patchValue({
          esCluster: !!extension?.related?.edge?.nodes ? extension.related.edge.nodes[0].id : ''
        })
        if(!!extension?.related?.edge?.nodes){
          this.removeRelateClusterList.push(extension.related.edge.nodes[0].id)
        }
        this.showEditFluentbit = true;
        break;
      case 'OTHER':
        this.updateOtherParamForm.patchValue({
          name: extension.cluster.name,
          description: extension.cluster.description,
          endpoint: extension.endpoints[0]
        })
        this.showEditOtherService = true;
        break;
    }
  }

  doUpdateOctantRelated(): void {
    this.showEditOctant = false;
    const addRelateClusterList: string[] = [];
    if(!!this.updateOctantRelatedForm.value.sharedCluster){
      for(let cluster of this.updateOctantRelatedForm.value.sharedCluster){
        addRelateClusterList.push(cluster.id);
      }
    }
    const param: ModifyClusterExtension = {
      id: this.selectedExtensionId,
      addRelated: addRelateClusterList,
      removeRelated: this.removeRelateClusterList
    }
    this.clusterExtensionsService.modifyClusterExtensions(param).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.ModifyClusterExtension){
        this.toast.show(`${data.data.ModifyClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-MSG'), this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
      this.doUpdateOctantReset();
      this.refresh();
     })
  }

  doUpdateOctantReset(): void {
    this.selectedClusterName = '';
    this.selectedExtensionId = '';
    this.updateOctantRelatedForm.reset();
    this.sharedClusterList = [];
    this.removeRelateClusterList = [];
  }

  doUpdateGrafanaRelated(): void {
    this.showEditGrafana = false;
    const addRelateClusterList: string[] = [];
    if(!!this.updateGrafanaRelatedForm.value.prometheusCluster){
      for(let cluster of this.updateGrafanaRelatedForm.value.prometheusCluster){
        addRelateClusterList.push(cluster.id);
      }
    }
    const param: ModifyClusterExtension = {
      id: this.selectedExtensionId,
      addRelated: addRelateClusterList,
      removeRelated: this.removeRelateClusterList
    }
    this.clusterExtensionsService.modifyClusterExtensions(param).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.ModifyClusterExtension){
        this.toast.show(`${data.data.ModifyClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-MSG'), this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
      this.doUpdateGrafanaReset();
      this.refresh();
     })
  }

  doUpdateGrafanaReset(): void {
    this.selectedClusterName = '';
    this.selectedExtensionId = '';
    this.updateGrafanaRelatedForm.reset();
    this.extensionEnabledClusters = [];
    this.removeRelateClusterList = [];
  }

  doUpdatePrometheusStorage(): void {
    this.showEditPrometheus = false;
    const paramsMap: JSON = JSON.parse(JSON.stringify({}));
    paramsMap['PROMETHEUS_DATASTORE_SIZE_GIB'] = !!this.updatePrometheusParamForm.value.storage ? this.updatePrometheusParamForm.value.storage+'' : '256';
    const param: ModifyClusterExtension = {
      id: this.selectedExtensionId,
      spec: paramsMap
    }
    this.clusterExtensionsService.modifyClusterExtensions(param).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.ModifyClusterExtension){
        this.toast.show(`${data.data.ModifyClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-MSG'), this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
      this.doUpdatePrometheusReset();
      this.refresh();
     })
  }

  doUpdatePrometheusReset(): void {
    this.selectedClusterName = '';
    this.selectedExtensionId = '';
    this.updatePrometheusParamForm.reset();
    this.currentPrometheusStorage = 0;
  }

  doUpdateElasticsearchParams(): void {
    this.showEditElasticsearch = false;
    const paramsMap: JSON = JSON.parse(JSON.stringify({}));
    paramsMap['ES_INSTANCE_REPLICAS'] = !!this.updateESParamForm.value.esDataNodeCount ? this.updateESParamForm.value.esDataNodeCount+'' : '1';
    paramsMap['ES_DATA_HEAPSIZE_GIB'] = !!this.updateESParamForm.value.esRequestMEM ? Math.round(this.updateESParamForm.value.esRequestMEM/2)+'' : '1';
    paramsMap['ES_RESOURCE_LIMIT_CPU'] = !!this.updateESParamForm.value.esRequestCPU ? this.updateESParamForm.value.esRequestCPU*2+'' : '2';
    paramsMap['ES_RESOURCE_LIMIT_MEMORY_GIB'] = !!this.updateESParamForm.value.esRequestMEM ? this.updateESParamForm.value.esRequestMEM*2+'' : '4';
    paramsMap['ES_RESOURCE_REQUEST_CPU'] = !!this.updateESParamForm.value.esRequestCPU ? this.updateESParamForm.value.esRequestCPU+'' : '1';
    paramsMap['ES_RESOURCE_REQUEST_MEMORY_GIB'] = !!this.updateESParamForm.value.esRequestMEM ? this.updateESParamForm.value.esRequestMEM+'' : '2';
    paramsMap['ES_DATASTORE_SIZE_GIB'] = !!this.updateESParamForm.value.esDataVolume ? this.updateESParamForm.value.esDataVolume+'' : '50';
    paramsMap['ES_PRESERVED_UNIT_COUNT'] = !!this.updateESParamForm.value.esLogsRetainDays ? this.updateESParamForm.value.esLogsRetainDays+'' : '7';
    const param: ModifyClusterExtension = {
      id: this.selectedExtensionId,
      spec: paramsMap
    }
    this.clusterExtensionsService.modifyClusterExtensions(param).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.ModifyClusterExtension){
        this.toast.show(`${data.data.ModifyClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-MSG'), this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
      this.doUpdateElasticsearchReset();
      this.refresh();
     })
  }

  doUpdateElasticsearchReset(): void {
    this.selectedClusterName = '';
    this.selectedExtensionId = '';
    this.updateESParamForm.reset();
    this.currentElasticsearchStorage = 0;
  }

  doUpdateFluentbitParams(): void {
    this.showEditFluentbit = false;
    const fluentbitParamsMap: JSON = JSON.parse(JSON.stringify({}));
    fluentbitParamsMap['FLUENTBIT_ENABLE_AUDIT_LOGGING'] = this.updateFluentbitParamForm.value.enableAudit+'';
    const param: ModifyClusterExtension = {
      id: this.selectedExtensionId,
      addRelated: !!this.updateFluentbitParamForm.value.esCluster ? [this.updateFluentbitParamForm.value.esCluster] : null,
      removeRelated: this.removeRelateClusterList,
      spec: fluentbitParamsMap
    }
    this.clusterExtensionsService.modifyClusterExtensions(param).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.ModifyClusterExtension){
        this.toast.show(`${data.data.ModifyClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-MSG'), this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
      this.doUpdateFluentbitReset();
      this.refresh();
     })
  }

  doUpdateFluentbitReset(): void {
    this.selectedClusterName = '';
    this.selectedExtensionId = '';
    this.updateFluentbitParamForm.reset();
    this.extensionEnabledClusters = [];
    this.removeRelateClusterList = [];
  }

  doUpdateOtherServiceParams(): void {
    this.showEditOtherService = false;
    const param: UpdateShortcut = {
      id: this.selectedExtensionId,
      description: this.updateOtherParamForm.value.description,
      endpoint: this.updateOtherParamForm.value.endpoint
    }
    this.clusterExtensionsService.updateShortcut(param).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-SHOTCUT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.UpdateShortcut){
        this.toast.show(`${data.data.UpdateShortcut.name}`+this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-SHOTCUT-MSG'), this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-SHOTCUT-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.UPDATE-SHOTCUT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
      this.doUpdateOtherServiceReset();
      this.refresh();
     })
  }

  doUpdateOtherServiceReset(): void {
    this.selectedClusterName = '';
    this.selectedExtensionId = '';
    this.updateOtherParamForm.reset();
  }

  showExtensionService(extension: ClusterExtension): void {
    this.selectedClusterName = extension.cluster.name;
    this.selectedExtension = extension;
    this.showExtensionOpen = true;
  }

  showExtensionrReset(): void {
    this.selectedClusterName = '';
    this.selectedExtension = null;
    this.showExtensionOpen = false;
  }

  showRemoveOperationService(extension: ClusterExtension): void {
    this.selectedExtensionName = extension.name;
    this.selectedClusterName = extension.cluster.name;
    this.selectedExtensionId = extension.id;
    this.removeDialog = true;
  }

  deleteOperationById(): void {
    this.removeDialog = false;
    switch (this.selectedExtensionName) {
      case 'OTHER':
        this.clusterExtensionsService.deleteShortcut(this.selectedExtensionId).subscribe( data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.DELETE-SHOTCUT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.DeleteShortcut){
            this.toast.show(`${data.data.DeleteShortcut.name}`+this.translate.instant('MESSAGE.CLUSTER-EXTENSION.DELETE-SHOTCUT-MSG'), this.translate.instant('MESSAGE.CLUSTER-EXTENSION.DELETE-SHOTCUT-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.DELETE-SHOTCUT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
           this.deleteOperationReset();
           this.refresh();
        })
        break;
      default:
        this.clusterExtensionsService.disableClusterExtensions(this.selectedExtensionId).subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.DELETE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.DisableClusterExtension){
        this.toast.show(`${data.data.DisableClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER-EXTENSION.DELETE-EXTENSION-MSG'), this.translate.instant('MESSAGE.CLUSTER-EXTENSION.DELETE-EXTENSION-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.DELETE-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
           this.deleteOperationReset();
           this.refresh();
        })
    }
  }

  deleteOperationReset(): void {
    this.selectedClusterName = '';
    this.selectedExtensionId = '';
    this.deleteOperationForm.reset();
  }

}
