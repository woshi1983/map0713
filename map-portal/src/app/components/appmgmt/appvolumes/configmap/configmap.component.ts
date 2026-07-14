import {Component, ElementRef, OnInit, ViewChild} from '@angular/core';
import {Cluster, ConfigMap, CreateConfigMap, handleObject, Namespace, UpdateConfigMap} from "../../../../classes";
import {FormArray, FormControl, FormGroup, Validators} from "@angular/forms";
import {AuthService} from "../../../../services/auth.service";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {VolumeService} from "../../../../services/volume.service";
import {ClustersService} from "../../../../services/clusters.service";
import {ClrDatagridPagination, ClrWizard} from "@clr/angular";
import {CustomValidatorDirective, nameValidator} from "../../../../utils/validators/custom-validator.directive";
import {OrganizationService} from "../../../../services/organization.service";
import {namespacediffset} from "../../../../utils/operator";

@Component({
  selector: 'app-configmap',
  templateUrl: './configmap.component.html',
  styleUrls: ['./configmap.component.css']
})
export class ConfigmapComponent implements OnInit {

  //admin and user role can handle cluster resources
  get canWriteResources(): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER_RESOURCE);
  }

  constructor(
     public authService: AuthService,
     private clusterService: ClustersService,
     private volumeService: VolumeService,
     private toast: ToastrService,
     private translate: TranslateService,
     private orgService: OrganizationService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  orgId: string;
  clusterList: Cluster[] = [];
  allConfigmap: ConfigMap[] = [];
  selectedConfigmap: ConfigMap;
  selectedConfigmapName: string;

  namespaceList: Namespace[] = [];
  showNoNonDefaultNamespaceAlert = false;

  showSearch = true;
  searchName: string;
  loading = false;

  createConfigmapOpen = false;
  updateConfigmapOpen = false;
  removeDialog = false;

  showAddFile = false;
  showAddEditFile = false;
  filename: string;
  editFilename: string;

  @ViewChild("fileUpload", {static: false}) fileUpload: ElementRef;
  @ViewChild("fileEditUpload", {static: false}) fileEditUpload: ElementRef;
  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardconfigmap') wizardnewconfigmap: ClrWizard;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  basicComfigmapForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    cluster: new FormControl('', Validators.required),
    namespace: new FormControl('', Validators.required)
  });
  dataList = new FormArray([]);
  dataForm = new FormGroup({
    cmData: this.dataList
  });
  editDataList = new FormArray([]);
  editDataForm = new FormGroup({
    cmData: this.editDataList
  });
  deleteConfigmapForm = new FormGroup({
    confirm: new FormControl('', [Validators.required]),
  });


  ngOnInit(): void {
  }

  refresh(): void {
    this.loading = true;
    const filters = [{ key: 'organization', value: [this.orgId] }, { key: 'omitKappChanges' }];
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.volumeService.getAllConfigmaps(filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.CONFIGMAP.GET-CONFIGMAP-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.allConfigmap = !!data.data?.ConfigMaps ? data.data.ConfigMaps : [];
      this.selectedConfigmap = null;
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.VOLUME.CONFIGMAP.GET-CONFIGMAP-FAILED'), {
      toastClass: 'alert-danger',
      closeButton: true,
      extendedTimeOut: 0,
      timeOut: 0
    }), () => this.loading = false);
  }

  getAllHealthyClusterIds(): void {
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

  getAllNamespaces($event: any): void {
    const clusterId = $event.target.value;
    this.showNoNonDefaultNamespaceAlert = false;
    this.namespaceList = [];
    this.clusterService.getClusterNamespaces(clusterId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.namespaceList = !!data.data?.Cluster?.namespaces ? namespacediffset(data.data.Cluster.namespaces) : [];
      this.showNoNonDefaultNamespaceAlert = this.namespaceList.length <= 0
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
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

  showCreateConfigmap(): void {
    this.getAllHealthyClusterIds();
    this.createConfigmapOpen = true;
  }

  doCreateConfigmapFinish(): void {
    const dataMap: JSON = JSON.parse(JSON.stringify({}));
    if (!!this.dataForm.value.cmData) {
      for (const each of this.dataForm.value.cmData) {
        dataMap[each.key] = each.value;
      }
    }
    const cmParam: CreateConfigMap = {
      name: this.basicComfigmapForm.value.name,
      namespace: this.basicComfigmapForm.value.namespace,
      data: dataMap,
      cluster: this.basicComfigmapForm.value.cluster
    };
    this.volumeService.createConfigMap(cmParam).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.CONFIGMAP.CREATE-CM-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.CreateConfigMap){
        this.toast.show(`${data.data.CreateConfigMap.name} `+this.translate.instant('MESSAGE.VOLUME.CONFIGMAP.CREATE-CM-MSG'), this.translate.instant('MESSAGE.VOLUME.CONFIGMAP.CREATE-CM-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.CONFIGMAP.CREATE-CM-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doCreateConfigmapReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doCreateConfigmapCancel(): void {
    this.doCreateConfigmapReset()
  }

  private doCreateConfigmapReset(): void {
    this.wizardnewconfigmap.reset();
    this.basicComfigmapForm.reset();
    while (this.dataList.length !== 0) {
      this.dataList.removeAt(0);
    }
    this.clusterList = [];
    this.namespaceList = [];
    this.createConfigmapOpen = false;
  }

  deleteData(index: number): void {
    this.dataList.removeAt(index);
  }

  addKeyValueData(): void {
    this.dataList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  showAddFileData(): void {
    this.showAddFile = true;
  }


  addFileToData(): void {
    const fileUpload = this.fileUpload.nativeElement;
    const file = fileUpload.files[0];
    if (file) {
      var reader = new FileReader();
      reader.readAsText(file, "UTF-8");
      reader.onload = (e) => {
        const fileContext = reader.result;
        this.dataList.push(new FormGroup({
          key: new FormControl(file.name, Validators.required),
          value: new FormControl(fileContext, Validators.required)
        }));
      };
    }
    this.fileUpload.nativeElement.value = '';
    this.filename = null;
    this.showAddFile = false;
  }

  cancelAddFile(): void {
    this.fileUpload.nativeElement.value = null;
    this.filename = null;
    this.showAddFile = false;
  }

  getFileName(): void {
    const fileUpload = this.fileUpload.nativeElement;
    const file = fileUpload.files[0];
    this.filename = file ? file.name : '';
  }

  showUpdateConfigmap(cm: ConfigMap): void {
    this.selectedConfigmap = cm;
    if (!!cm.data) {
      const datas = cm.data;
      for (const key of Object.keys(datas)) {
        const value = datas[key];
        this.editDataList.push(new FormGroup({
          key: new FormControl(key),
          value: new FormControl(value)
        }));
      }
    }
    this.updateConfigmapOpen = true;
  }

  deleteEditData(index: number): void {
    this.editDataList.removeAt(index);
  }

  addEditData(): void {
    this.editDataList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  showAddEditFileData(): void {
    this.showAddEditFile = true;
  }

  addEditFileToData(): void {
    const fileUpload = this.fileEditUpload.nativeElement;
    const file = fileUpload.files[0];
    if (file) {
      var reader = new FileReader();
      reader.readAsText(file, "UTF-8");
      reader.onload = (e) => {
        const fileContext = reader.result;
        this.editDataList.push(new FormGroup({
          key: new FormControl(file.name, Validators.required),
          value: new FormControl(fileContext, Validators.required)
        }));
      };
    }
    this.fileEditUpload.nativeElement.value = '';
    this.editFilename = null;
    this.showAddEditFile = false;
  }

  cancelAddEditFile(): void {
    this.fileEditUpload.nativeElement.value = null;
    this.editFilename = null;
    this.showAddEditFile = false;
  }

  getEditFileName(): void {
    const fileUpload = this.fileEditUpload.nativeElement;
    const file = fileUpload.files[0];
    this.editFilename = file ? file.name : '';
  }

  doUpdateConfigmapFinish(): void {
    const dataMap: JSON = JSON.parse(JSON.stringify({}));
    if (!!this.editDataForm.value.cmData) {
      for (const each of this.editDataForm.value.cmData) {
        dataMap[each.key] = each.value;
      }
    }
    const cmParam: UpdateConfigMap = {
      name: this.selectedConfigmap.name,
      namespace: this.selectedConfigmap.namespace,
      data: dataMap,
      cluster: this.selectedConfigmap.cluster.id,
    };
    this.volumeService.updateConfigMap(cmParam).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.CONFIGMAP.UPDATE-CM-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.UpdateConfigMap){
        this.toast.show(`${data.data.UpdateConfigMap.name} `+this.translate.instant('MESSAGE.VOLUME.CONFIGMAP.UPDATE-CM-MSG'), this.translate.instant('MESSAGE.VOLUME.CONFIGMAP.UPDATE-CM-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.CONFIGMAP.UPDATE-CM-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doUpdateConfigmapReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doUpdateConfigmapCancel(): void {
    this.doUpdateConfigmapReset()
  }

  private doUpdateConfigmapReset(): void {
    while (this.editDataList.length !== 0) {
      this.editDataList.removeAt(0);
    }
    this.updateConfigmapOpen = false;
  }

  showRemoveDialog(cm: ConfigMap): void {
    this.selectedConfigmap = cm;
    this.selectedConfigmapName = cm.name;
    this.removeDialog = true;
  }


  deleteConfigmapByName(cm: ConfigMap): void {
    this.removeDialog = false;
    this.volumeService.deleteConfigMap(cm.name, cm.cluster.id, cm.namespace).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.CONFIGMAP.DELETE-CM-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.DeleteConfigMap){
          this.toast.show(`${data.data.DeleteConfigMap.name} `+this.translate.instant('MESSAGE.VOLUME.CONFIGMAP.DELETE-CM-MSG'), this.translate.instant('MESSAGE.VOLUME.CONFIGMAP.DELETE-CM-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.CONFIGMAP.DELETE-CM-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

}
