import {Component, OnInit, ViewChild} from '@angular/core';
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {OrganizationService} from "../../../../services/organization.service";
import {AuthService} from "../../../../services/auth.service";
import {ClustersService} from "../../../../services/clusters.service";
import {VolumeService} from "../../../../services/volume.service";
import {
  Cluster, ClusterExtension,
  CreateStorageClass, EnableClusterExtension,
  handleObject,
  StorageClass,
  UpdateStorageClass
} from "../../../../classes";
import {ClrDatagridPagination,ClrDatagridStateInterface, ClrWizard} from "@clr/angular";
import {CustomValidatorDirective, nameValidator} from "../../../../utils/validators/custom-validator.directive";
import {FormArray, FormControl, FormGroup, Validators} from "@angular/forms";
import {ClusterExtensionService} from "../../../../services/cluster-extension.service";

@Component({
  selector: 'app-storageclass',
  templateUrl: './storageclass.component.html',
  styleUrls: ['./storageclass.component.css']
})
export class StorageclassComponent implements OnInit {

  //admin and user role can handle cluster resources
  get canWriteResources(): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER_RESOURCE);
  }

  constructor(
    public authService: AuthService,
    private clusterService: ClustersService,
    private toast: ToastrService,
    private translate: TranslateService,
    private orgService: OrganizationService,
    private volumeService: VolumeService,
    private clusterExtensionsService: ClusterExtensionService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  orgId: string;
  clusterList: Cluster[];
  allStorageClass: StorageClass[] = [];
  selectedStorageClass: StorageClass;

  showSearch = true;
  searchName: string;
  loading = false;

  createSCOpen = false;
  showUpdateSCLabels = false;
  showUpdateSCAnnos = false;
  showUpdateSCOptions = false;
  showUpdateSCDefault = false;
  removeDialog = false;

  @ViewChild('csipagination') csipagination: ClrDatagridPagination;
  nfsControllerInstalled = false;
  allCSIController: ClusterExtension[] = [];
  selectedCSIController: ClusterExtension;
  extensionDisableClusters: Cluster[] = [];
  showEnableCSI = false;
  showDisableCSI = false;
  csiLoading = false;
  csiTotal = 0;
  enableCSIControllerForm = new FormGroup({
    name: new FormControl('CSI_DRIVER_NFS', [Validators.required]),
    cluster: new FormControl('', Validators.required),
  });

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardsc') wizardnewsc: ClrWizard;
  readonly disAllowLabels = CustomValidatorDirective.disAllowLabels;
  readonly commonNamePattern = CustomValidatorDirective.commonNamePattern;
  labelsList = new FormArray([]);
  annotationsList = new FormArray([]);
  basicSCForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.commonNamePattern), nameValidator(['default'])]),
    cluster: new FormControl('', Validators.required),
    type: new FormControl('Common'),
    labels: this.labelsList,
    annotations: this.annotationsList
  });
  parameters = new FormArray([])
  options = new FormArray([])
  scParamsForm = new FormGroup({
    enableCSIController: new FormControl(true),
    nfsServer: new FormControl(''),
    nfsPath: new FormControl(''),
    nfsServerVersion: new FormControl('4.1'),
    vsphereStoragePolicy: new FormControl(''),
    cephMonitors: new FormControl(''),
    cephAdminId: new FormControl('admin'),
    cephAdminSecretName: new FormControl(''),
    cephAdminSecretNamespace: new FormControl(''),
    cephPool: new FormControl('rbd'),
    cephUserId: new FormControl('admin'),
    cephUserSecretName: new FormControl(''),
    cephUserSecretNamespace: new FormControl(''),
    provisioner: new FormControl(''),
    reclaimPolicy: new FormControl('Delete'),
    volumeBindingMode: new FormControl('Immediate'),
    allowVolumeExpansion: new FormControl(false),
    parameters: this.parameters,
    mountOptions: this.options
  });

  editLabelsList = new FormArray([]);
  editAnnotationsList = new FormArray([]);
  editSCForm = new FormGroup({
    labels: this.editLabelsList,
    annotations: this.editAnnotationsList
  });
  editoptions = new FormArray([])
  editSCParamsForm = new FormGroup({
    allowVolumeExpansion: new FormControl(false),
    mountOptions: this.editoptions
  });

  deleteSCForm = new FormGroup({
    confirm: new FormControl('', [Validators.required]),
  });

  ngOnInit(): void {
  }

  refresh(): void {
    this.loading = true;
    const filters = [{ key: 'organization', value: [this.orgId] }];
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.volumeService.getAllStorageClasses(filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.SC.GET-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.allStorageClass = !!data.data?.StorageClasses ? data.data.StorageClasses : [];
      this.selectedStorageClass = null;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SC.GET-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false);
  }

  refreshCSIController(state?: ClrDatagridStateInterface): void {
    this.csiLoading = true;
    this.clusterExtensionsService.getAllCSIControllerExtension(!!state?.page?.current ? state.page.current - 1 : 0, !!state?.page?.size ? state.page.size : Number(this.csipagination.pageSize), this.orgId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.GET-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.csiTotal = !!data.data?.ClusterExtensions?.totalCount ? data.data.ClusterExtensions.totalCount : 0;
      this.allCSIController = !!data.data?.ClusterExtensions?.edge?.nodes ? data.data.ClusterExtensions.edge.nodes : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-EXTENSION.GET-EXTENSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.csiLoading = false;
    })
  }

  checkDefaultSC(annotations: JSON): boolean {
    if(!!annotations){
      for (const key of Object.keys(annotations)) {
        if(key === 'storageclass.kubernetes.io/is-default-class'){
          return true;
        }
      }
    }
    return false;
  }

  getAllHealthyClusterIds(): void {
    this.clusterList = null;
    this.clusterService.getAllHealthyClusterId(this.orgId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.clusterList = !!data.data?.Clusters?.edge?.nodes ? data.data.Clusters.edge.nodes : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
  }

  checkCanEditable(labels: JSON): boolean {
    //check ingress with label 'map.tanzu.vmware.com/component'
    if(!!labels){
      for (const key of Object.keys(labels)) {
        if(key === 'map.tanzu.vmware.com/component'){
          return false;
        }
      }
    }
    return true;
  }

  showSetDefaultStorageClass(sc: StorageClass): void {
    this.selectedStorageClass = sc;
    this.showUpdateSCDefault = true;
  }

  updateSCDefault(): void {
    // check select sc
    const annosMap: JSON = JSON.parse(JSON.stringify({}));
    if(this.checkDefaultSC(this.selectedStorageClass.annotations)){
      // sc is default, remove
      for (const key of Object.keys(this.selectedStorageClass.annotations)) {
        if(key !== 'storageclass.kubernetes.io/is-default-class'){
          annosMap[key] = this.selectedStorageClass.annotations[key];
        }
      }
    } else {
      // sc is not default, want to set default
      // check cluster not has default sc
      const currentCluster = this.selectedStorageClass.cluster.id;
      for(const sc of this.allStorageClass){
        if(sc.cluster.id === currentCluster){
          if(this.checkDefaultSC(sc.annotations)){
            // has default
            this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` 当前集群已经存在默认存储类。`, this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
            this.showUpdateSCDefault = false;
            this.selectedStorageClass = null;
            return
          }
        }
      }
      // no default sc in cluster
      if(!!this.selectedStorageClass.annotations){
        for (const key of Object.keys(this.selectedStorageClass.annotations)) {
          annosMap[key] = this.selectedStorageClass.annotations[key];
        }
      }
      annosMap['storageclass.kubernetes.io/is-default-class'] = 'true';
    }
    const scParam: UpdateStorageClass = {
      name: this.selectedStorageClass.name,
      annotations: annosMap,
      cluster: this.selectedStorageClass.cluster.id
    }
    this.volumeService.updateStorageClass(scParam).subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors) {
          if (!!each.message) {
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
          }
        }
      }
      if(!!data.data?.updateStorageClass){
        this.toast.show(`${data.data.updateStorageClass.name} `+this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-MSG'), this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.refresh();
      this.showUpdateSCDefault = false;
      this.selectedStorageClass = null;
    })
  }

  showCreateStorageClass(): void {
    this.getAllHealthyClusterIds();
    this.createSCOpen = true;
  }

  initSCParams(type: string): void {
    switch (type) {
      case 'NFS':
        this.scParamsForm.reset();
        this.nfsControllerInstalled = false;
        //check csi-nfs controller installed or not
        for(let csi of this.allCSIController){
          if(csi.cluster.id === this.basicSCForm.value.cluster){
            this.nfsControllerInstalled = true;
          }
        }
        this.scParamsForm.controls['vsphereStoragePolicy'].clearValidators();
        this.scParamsForm.controls['vsphereStoragePolicy'].updateValueAndValidity();
        this.scParamsForm.controls['cephMonitors'].clearValidators();
        this.scParamsForm.controls['cephMonitors'].updateValueAndValidity();
        this.scParamsForm.controls['cephAdminSecretName'].clearValidators();
        this.scParamsForm.controls['cephAdminSecretName'].updateValueAndValidity();
        this.scParamsForm.controls['cephAdminSecretNamespace'].clearValidators();
        this.scParamsForm.controls['cephAdminSecretNamespace'].updateValueAndValidity();
        this.scParamsForm.controls['cephUserSecretName'].clearValidators();
        this.scParamsForm.controls['cephUserSecretName'].updateValueAndValidity();
        this.scParamsForm.controls['cephUserSecretNamespace'].clearValidators();
        this.scParamsForm.controls['cephUserSecretNamespace'].updateValueAndValidity();
        this.scParamsForm.patchValue({
          enableCSIController: true,
          nfsServerVersion: '4.1',
          reclaimPolicy: 'Delete',
          volumeBindingMode: 'Immediate',
          allowVolumeExpansion: false,
          provisioner: 'nfs.csi.k8s.io'
        })
        this.scParamsForm.controls['nfsServer'].setValidators(Validators.required);
        this.scParamsForm.controls['nfsServer'].updateValueAndValidity();
        this.scParamsForm.controls['nfsPath'].setValidators(Validators.required);
        this.scParamsForm.controls['nfsPath'].updateValueAndValidity();
        break;
      case 'vsphere-csi':
        this.scParamsForm.reset();
        this.scParamsForm.controls['nfsServer'].clearValidators();
        this.scParamsForm.controls['nfsServer'].updateValueAndValidity();
        this.scParamsForm.controls['nfsPath'].clearValidators();
        this.scParamsForm.controls['nfsPath'].updateValueAndValidity();
        this.scParamsForm.controls['cephMonitors'].clearValidators();
        this.scParamsForm.controls['cephMonitors'].updateValueAndValidity();
        this.scParamsForm.controls['cephAdminSecretName'].clearValidators();
        this.scParamsForm.controls['cephAdminSecretName'].updateValueAndValidity();
        this.scParamsForm.controls['cephAdminSecretNamespace'].clearValidators();
        this.scParamsForm.controls['cephAdminSecretNamespace'].updateValueAndValidity();
        this.scParamsForm.controls['cephUserSecretName'].clearValidators();
        this.scParamsForm.controls['cephUserSecretName'].updateValueAndValidity();
        this.scParamsForm.controls['cephUserSecretNamespace'].clearValidators();
        this.scParamsForm.controls['cephUserSecretNamespace'].updateValueAndValidity();
        this.scParamsForm.patchValue({
          enableCSIController: true,
          nfsServerVersion: '4.1',
          reclaimPolicy: 'Delete',
          volumeBindingMode: 'Immediate',
          allowVolumeExpansion: false,
          provisioner: 'csi.vsphere.vmware.com'
        })
        this.scParamsForm.controls['vsphereStoragePolicy'].setValidators(Validators.required);
        this.scParamsForm.controls['vsphereStoragePolicy'].updateValueAndValidity();
        break;
      case 'ceph-rbd':
        this.scParamsForm.reset();
        this.scParamsForm.controls['nfsServer'].clearValidators();
        this.scParamsForm.controls['nfsServer'].updateValueAndValidity();
        this.scParamsForm.controls['nfsPath'].clearValidators();
        this.scParamsForm.controls['nfsPath'].updateValueAndValidity();
        this.scParamsForm.controls['vsphereStoragePolicy'].clearValidators();
        this.scParamsForm.controls['vsphereStoragePolicy'].updateValueAndValidity();
        this.scParamsForm.patchValue({
          enableCSIController: true,
          nfsServerVersion: '4.1',
          reclaimPolicy: 'Delete',
          volumeBindingMode: 'Immediate',
          allowVolumeExpansion: false,
          cephMonitors: '',
          cephAdminId: 'admin',
          cephAdminSecretName: '',
          cephAdminSecretNamespace: '',
          cephPool: 'rbd',
          cephUserId: 'admin',
          cephUserSecretName: '',
          cephUserSecretNamespace: '',
          cephFstype: 'ext4',
          provisioner: 'kubernetes.io/rbd'
        })
        this.scParamsForm.controls['cephMonitors'].setValidators(Validators.required);
        this.scParamsForm.controls['cephMonitors'].updateValueAndValidity();
        this.scParamsForm.controls['cephAdminSecretName'].setValidators(Validators.required);
        this.scParamsForm.controls['cephAdminSecretName'].updateValueAndValidity();
        this.scParamsForm.controls['cephAdminSecretNamespace'].setValidators(Validators.required);
        this.scParamsForm.controls['cephAdminSecretNamespace'].updateValueAndValidity();
        this.scParamsForm.controls['cephUserSecretName'].setValidators(Validators.required);
        this.scParamsForm.controls['cephUserSecretName'].updateValueAndValidity();
        this.scParamsForm.controls['cephUserSecretNamespace'].setValidators(Validators.required);
        this.scParamsForm.controls['cephUserSecretNamespace'].updateValueAndValidity();
        break;
      case 'static':
        this.scParamsForm.reset();
        this.scParamsForm.controls['nfsServer'].clearValidators();
        this.scParamsForm.controls['nfsServer'].updateValueAndValidity();
        this.scParamsForm.controls['nfsPath'].clearValidators();
        this.scParamsForm.controls['nfsPath'].updateValueAndValidity();
        this.scParamsForm.controls['vsphereStoragePolicy'].clearValidators();
        this.scParamsForm.controls['vsphereStoragePolicy'].updateValueAndValidity();
        this.scParamsForm.controls['cephMonitors'].clearValidators();
        this.scParamsForm.controls['cephMonitors'].updateValueAndValidity();
        this.scParamsForm.controls['cephAdminSecretName'].clearValidators();
        this.scParamsForm.controls['cephAdminSecretName'].updateValueAndValidity();
        this.scParamsForm.controls['cephAdminSecretNamespace'].clearValidators();
        this.scParamsForm.controls['cephAdminSecretNamespace'].updateValueAndValidity();
        this.scParamsForm.controls['cephUserSecretName'].clearValidators();
        this.scParamsForm.controls['cephUserSecretName'].updateValueAndValidity();
        this.scParamsForm.controls['cephUserSecretNamespace'].clearValidators();
        this.scParamsForm.controls['cephUserSecretNamespace'].updateValueAndValidity();
        this.scParamsForm.patchValue({
          enableCSIController: true,
          nfsServerVersion: '4.1',
          reclaimPolicy: 'Delete',
          volumeBindingMode: 'WaitForFirstConsumer',
          allowVolumeExpansion: false,
          provisioner: 'kubernetes.io/no-provisioner'
        })
        break;
      default:
        this.scParamsForm.reset();
        this.scParamsForm.controls['nfsServer'].clearValidators();
        this.scParamsForm.controls['nfsServer'].updateValueAndValidity();
        this.scParamsForm.controls['nfsPath'].clearValidators();
        this.scParamsForm.controls['nfsPath'].updateValueAndValidity();
        this.scParamsForm.controls['vsphereStoragePolicy'].clearValidators();
        this.scParamsForm.controls['vsphereStoragePolicy'].updateValueAndValidity();
        this.scParamsForm.controls['cephMonitors'].clearValidators();
        this.scParamsForm.controls['cephMonitors'].updateValueAndValidity();
        this.scParamsForm.controls['cephAdminSecretName'].clearValidators();
        this.scParamsForm.controls['cephAdminSecretName'].updateValueAndValidity();
        this.scParamsForm.controls['cephAdminSecretNamespace'].clearValidators();
        this.scParamsForm.controls['cephAdminSecretNamespace'].updateValueAndValidity();
        this.scParamsForm.controls['cephUserSecretName'].clearValidators();
        this.scParamsForm.controls['cephUserSecretName'].updateValueAndValidity();
        this.scParamsForm.controls['cephUserSecretNamespace'].clearValidators();
        this.scParamsForm.controls['cephUserSecretNamespace'].updateValueAndValidity();
        while (this.parameters.length !== 0){
          this.parameters.removeAt(0)
        }
        while (this.options.length !== 0){
          this.options.removeAt(0)
        }
        this.scParamsForm.patchValue({
          enableCSIController: true,
          nfsServerVersion: '4.1',
          reclaimPolicy: 'Delete',
          volumeBindingMode: 'Immediate',
          allowVolumeExpansion: false,
          provisioner: ''
        });
        this.scParamsForm.controls['provisioner'].setValidators(Validators.required);
        this.scParamsForm.controls['provisioner'].updateValueAndValidity();
        break;
    }
  }

  addNewSCLabel(): void {
    this.labelsList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  deleteSCLable(index: number): void{
    this.labelsList.removeAt(index);
  }

  addNewSCAnnotation(): void {
    this.annotationsList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('')
    }));
  }

  deleteSCAnnotation(index: number): void{
    this.annotationsList.removeAt(index);
  }

  addNewSCParams(): void {
    this.parameters.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  deleteSCParams(index: number): void{
    this.parameters.removeAt(index);
  }

  addNewSCOptions(): void {
    this.options.push(new FormGroup({
      key: new FormControl(''),
    }));
  }

  deleteSCOptions(index: number): void{
    this.options.removeAt(index);
  }

  doCreateSCFinish(): void {
    const labelsMap: JSON = JSON.parse(JSON.stringify({}));
    if (!!this.basicSCForm.value.labels) {
      for (const each of this.basicSCForm.value.labels) {
        labelsMap[each.key] = each.value;
      }
    }
    const annotationMap: JSON = JSON.parse(JSON.stringify({}));
    if (!!this.basicSCForm.value.annotations) {
      for (const each of this.basicSCForm.value.annotations) {
        annotationMap[each.key] = each.value;
      }
    }
    const scParam: CreateStorageClass = {
      name: this.basicSCForm.value.name,
      labels: labelsMap,
      annotations: annotationMap,
      allowVolumeExpansion: !!this.scParamsForm.value.allowVolumeExpansion ? this.scParamsForm.value.allowVolumeExpansion : false,
      volumeBindingMode: !!this.scParamsForm.value.volumeBindingMode ? this.scParamsForm.value.volumeBindingMode : 'Immediate',
      reclaimPolicy: !!this.scParamsForm.value.reclaimPolicy ? this.scParamsForm.value.reclaimPolicy : 'reclaimPolicy',
      cluster: this.basicSCForm.value.cluster
    };
    switch (this.basicSCForm.value.type) {
      case 'NFS':
        scParam.provisioner = !!this.scParamsForm.value.provisioner ?  this.scParamsForm.value.provisioner : 'nfs.csi.k8s.io';
        const nfspParametersMap: JSON = JSON.parse(JSON.stringify({}));
        nfspParametersMap['server'] = this.scParamsForm.value.nfsServer;
        nfspParametersMap['share'] = this.scParamsForm.value.nfsPath;
        scParam.parameters = nfspParametersMap;
        scParam.mountOptions = ['hard', 'nfsvers=' + this.scParamsForm.value.nfsServerVersion ]
        break;
      case 'vsphere-csi':
        scParam.provisioner = !!this.scParamsForm.value.provisioner ?  this.scParamsForm.value.provisioner : 'csi.vsphere.vmware.com';
        const vcParametersMap: JSON = JSON.parse(JSON.stringify({}));
        vcParametersMap['svStorageClass'] = this.scParamsForm.value.vsphereStoragePolicy;
        scParam.parameters = vcParametersMap;
        break;
      case 'ceph-rbd':
        scParam.provisioner = !!this.scParamsForm.value.provisioner ?  this.scParamsForm.value.provisioner : 'kubernetes.io/rbd';
        const cephParametersMap: JSON = JSON.parse(JSON.stringify({}));
        cephParametersMap['monitors'] = this.scParamsForm.value.cephMonitors;
        cephParametersMap['adminId'] = !!this.scParamsForm.value.cephAdminId ? this.scParamsForm.value.cephAdminId : 'admin';
        cephParametersMap['adminSecretNamespace'] = this.scParamsForm.value.cephAdminSecretNamespace;
        cephParametersMap['adminSecretName'] = this.scParamsForm.value.cephAdminSecretName;
        cephParametersMap['userId'] = !!this.scParamsForm.value.cephUserId ? this.scParamsForm.value.cephUserId : 'admin';
        cephParametersMap['userSecretName'] = this.scParamsForm.value.cephUserSecretName;
        cephParametersMap['userSecretNamespace'] = this.scParamsForm.value.cephUserSecretNamespace;
        cephParametersMap['pool'] = !!this.scParamsForm.value.cephPool? this.scParamsForm.value.cephPool : 'rbd';
        scParam.parameters = cephParametersMap;
        break;
      case 'static':
        scParam.provisioner = !!this.scParamsForm.value.provisioner ?  this.scParamsForm.value.provisioner : 'kubernetes.io/no-provisioner';
        break;
      default:
        scParam.provisioner = this.scParamsForm.value.provisioner
        const paramsMap: JSON = JSON.parse(JSON.stringify({}));
        if (!!this.scParamsForm.value.parameters) {
          for (const each of this.scParamsForm.value.parameters) {
            paramsMap[each.key] = each.value;
          }
        }
        scParam.parameters = paramsMap;
        const stringList = [];
        if(!!this.scParamsForm.value.mountOptions){
          for (const each of this.scParamsForm.value.mountOptions) {
            stringList.push(each.key)
          }
        }
        scParam.mountOptions = stringList;
        break;
    }
    this.volumeService.createStorageClass(scParam).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.SC.CREATE-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.CreateStorageClass){
        this.toast.show(`${data.data.CreateStorageClass.name} `+this.translate.instant('MESSAGE.VOLUME.SC.CREATE-SC-MSG'), this.translate.instant('MESSAGE.VOLUME.SC.CREATE-SC-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
      // enable nfs csi
      if(!this.nfsControllerInstalled && this.scParamsForm.value.enableCSIController){
        const param: EnableClusterExtension = {
          name: 'CSI_DRIVER_NFS',
          cluster: this.basicSCForm.value.cluster
        }
        this.clusterExtensionsService.enableClusterExtensions(param).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-TBS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.EnableClusterExtension){
            this.toast.show(`${data.data.EnableClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-TBS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
           this.refreshCSIController();
         })
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SC.CREATE-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doCreateSCReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doCreateSCCancel(): void {
    this.doCreateSCReset();
  }

  doCreateSCReset(): void {
    this.wizardnewsc.reset();
    this.basicSCForm.reset();
    this.basicSCForm.patchValue({
      type: 'Common'
    })
    while (this.annotationsList.length > 0){
      this.annotationsList.removeAt(0);
    }
    while (this.labelsList.length > 0){
      this.labelsList.removeAt(0);
    }
    this.scParamsForm.reset();
    this.scParamsForm.patchValue({
      nfsServerVersion: '4.1',
      cephAdminId: 'admin',
      cephPool: 'rbd',
      cephUserId: 'admin',
      reclaimPolicy: 'Delete',
      volumeBindingMode: 'Immediate',
      allowVolumeExpansion: false
    })
    while (this.parameters.length > 0){
      this.parameters.removeAt(0);
    }
    while (this.options.length > 0){
      this.options.removeAt(0);
    }
    this.createSCOpen = false;
    this.clusterList = [];
  }

  showUpdateStorageClassLabels(sc: StorageClass): void {
    this.selectedStorageClass = sc;
    if (!!sc.labels) {
      const labels = sc.labels;
      for (const key of Object.keys(labels)) {
        this.editLabelsList.push(new FormGroup({
          key: new FormControl(key),
          value: new FormControl(labels[key])
        }));
      }
    }
    this.showUpdateSCLabels = true;
  }

  addEditLabel(): void {
    this.editLabelsList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  deleteEditLable(index: number): void{
    this.editLabelsList.removeAt(index);
  }

  updateSCLables(): void {
    const labelsMap: JSON = JSON.parse(JSON.stringify({}));
    if (!!this.editSCForm.value.labels) {
      for (const each of this.editSCForm.value.labels) {
        labelsMap[each.key] = each.value;
      }
    }
    const scParam: UpdateStorageClass = {
      name: this.selectedStorageClass.name,
      labels: labelsMap,
      cluster: this.selectedStorageClass.cluster.id
    }
    this.volumeService.updateStorageClass(scParam).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.updateStorageClass){
        this.toast.show(`${data.data.updateStorageClass.name} `+this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-MSG'), this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doUpdateSCReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  showUpdateStorageClassAnnotations(sc: StorageClass): void {
    this.selectedStorageClass = sc;
    if (!!sc.annotations) {
      const labels = sc.annotations;
      for (const key of Object.keys(labels)) {
        this.editAnnotationsList.push(new FormGroup({
          key: new FormControl(key),
          value: new FormControl(labels[key])
        }));
      }
    }
    this.showUpdateSCAnnos = true;
  }

  addEditAnno(): void {
    this.editAnnotationsList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('')
    }));
  }

  deleteEditAnno(index: number): void{
    this.editAnnotationsList.removeAt(index);
  }

  updateSCAnnos(): void {
    const annosMap: JSON = JSON.parse(JSON.stringify({}));
    if (!!this.editSCForm.value.annotations) {
      for (const each of this.editSCForm.value.annotations) {
        annosMap[each.key] = each.value;
      }
    }
    const scParam: UpdateStorageClass = {
      name: this.selectedStorageClass.name,
      annotations: annosMap,
      cluster: this.selectedStorageClass.cluster.id
    }
    this.volumeService.updateStorageClass(scParam).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.updateStorageClass){
        this.toast.show(`${data.data.updateStorageClass.name} `+this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-MSG'), this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doUpdateSCReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  showUpdateStorageClassOptions(sc: StorageClass): void {
    this.selectedStorageClass = sc;
    if (!!sc.mountOptions) {
      for (const each of sc.mountOptions) {
        this.editoptions.push(new FormGroup({
          key: new FormControl(each)
        }));
      }
    }
    this.editSCParamsForm.patchValue({
      allowVolumeExpansion: sc.allowVolumeExpansion
    })
    this.showUpdateSCOptions = true;
  }

  addEditOptions(): void {
    this.editoptions.push(new FormGroup({
      key: new FormControl('', Validators.required)
    }));
  }

  deleteEditOptions(index: number): void{
    this.editoptions.removeAt(index);
  }

  updateSCOptions(): void {
    const stringList = [];
    if(!!this.editSCParamsForm.value.mountOptions){
      for (const each of this.editSCParamsForm.value.mountOptions) {
        stringList.push(each.key)
      }
    }
    const scParam: UpdateStorageClass = {
      name: this.selectedStorageClass.name,
      mountOptions: stringList,
      allowVolumeExpansion: this.editSCParamsForm.value.allowVolumeExpansion,
      cluster: this.selectedStorageClass.cluster.id
    }
    this.volumeService.updateStorageClass(scParam).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.updateStorageClass){
        this.toast.show(`${data.data.updateStorageClass.name} `+this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-MSG'), this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SC.UPDATE-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doUpdateSCReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doUpdateSCReset(): void {
    this.editSCForm.reset();
    while (this.editLabelsList.length > 0){
      this.editLabelsList.removeAt(0)
    }
    while (this.editAnnotationsList.length > 0){
      this.editAnnotationsList.removeAt(0)
    }
    this.editSCParamsForm.reset();
    while (this.editoptions.length > 0){
      this.editoptions.removeAt(0)
    }
    this.showUpdateSCLabels = false;
    this.showUpdateSCAnnos = false;
    this.showUpdateSCOptions = false;
  }

  showRemoveDialog(sc: StorageClass): void {
    this.selectedStorageClass = sc;
    this.removeDialog = true;
  }

  deleteSCByName(): void {
    this.removeDialog = false;
    this.volumeService.deleteStorageClass(this.selectedStorageClass.name, this.selectedStorageClass.cluster.id).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.SC.DELETE-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.DeleteStorageClass){
          this.toast.show(`${data.data.DeleteStorageClass.name} `+this.translate.instant('MESSAGE.VOLUME.SC.DELETE-SC-MSG'), this.translate.instant('MESSAGE.VOLUME.SC.DELETE-SC-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SC.DELETE-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  showEnableCSIController(): void {
    this.clusterService.getExtensionDisabledClusterId(this.orgId, 'CSI_DRIVER_NFS').subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.extensionDisableClusters = !!data.data?.Clusters?.edge?.nodes ? data.data.Clusters.edge.nodes : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
    this.showEnableCSI = true;
  }

  doEnableCSIController(): void {
    const param: EnableClusterExtension = {
      name: this.enableCSIControllerForm.value.name,
      cluster: this.enableCSIControllerForm.value.cluster
    }
    this.clusterExtensionsService.enableClusterExtensions(param).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-TBS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.EnableClusterExtension){
        this.toast.show(`${data.data.EnableClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-TBS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
     this.refreshCSIController();
     this.doEnableCSIControllerReset();
   })
  }

  doEnableCSIControllerReset(): void {
    this.enableCSIControllerForm.reset();
    this.enableCSIControllerForm.patchValue({
      name: 'CSI_DRIVER_NFS'
    })
    this.showEnableCSI = false;
  }

  showDisableCSIController(): void {
    this.showDisableCSI = true;
  }

  disableCSIController(): void {
    this.showDisableCSI = false;
    this.clusterExtensionsService.disableClusterExtensions(this.selectedCSIController?.id).subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-TBS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.DisableClusterExtension){
        this.toast.show(`${data.data.DisableClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-TBS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
     this.refreshCSIController();
   })
  }

}
