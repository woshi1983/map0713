import {Component, OnInit, ViewChild} from '@angular/core';
import {
  Cluster,
  PersistentVolumeClaim,
  Namespace,
  CreatePersistentVolumeClaim, UpdatePersistentVolumeClaim, handleObject, StorageClass
} from "../../../../classes";
import { AuthService} from "../../../../services/auth.service";
import {ClustersService} from "../../../../services/clusters.service";
import {VolumeService} from "../../../../services/volume.service";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {ClrDatagridPagination, ClrWizard} from "@clr/angular";
import {CustomValidatorDirective, nameValidator} from "../../../../utils/validators/custom-validator.directive";
import {FormControl, FormGroup, Validators} from "@angular/forms";
import {OrganizationService} from "../../../../services/organization.service";
import {namespacediffset} from "../../../../utils/operator";

@Component({
  selector: 'app-pvc',
  templateUrl: './pvc.component.html',
  styleUrls: ['./pvc.component.css']
})
export class PvcComponent implements OnInit {

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
  clusterList: Cluster[];
  allPVCs: PersistentVolumeClaim[] = [];
  selectedPVC: PersistentVolumeClaim;
  selectedPVCName: string;
  allStorageClass: StorageClass[] = [];

  namespaceList: Namespace[] = [];
  showNoNonDefaultNamespaceAlert = false;
  allowVolumeExpansion = false;

  showSearch = true;
  searchName: string;
  loading = false;

  createPVCOpen = false;
  showUpdatePvc = false;
  currentVolumeSize: number;
  removeDialog = false;

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardpvc') wizardnewpvc: ClrWizard;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  basicPVCForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    cluster: new FormControl('', Validators.required),
    namespace: new FormControl('', Validators.required),
    storageClass: new FormControl('', Validators.required),
    accessMode: new FormControl('', Validators.required),
    request: new FormControl('', Validators.required),
    limit: new FormControl(''),
    volumeMode: new FormControl('Filesystem')
  });
  updatePVCForm = new FormGroup({
    request: new FormControl('', Validators.required)
  });
  deletePVCForm = new FormGroup({
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
    this.volumeService.getAllPVCs(filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.PVC.GET-PVC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.allPVCs = !!data.data?.PersistentVolumeClaims ? data.data.PersistentVolumeClaims : [];
      this.selectedPVC = null;
      }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.PVC.GET-PVC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false);
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

  getAllStorageClasses(clusterId: string): void {
    this.allStorageClass = null;
    this.volumeService.getAllStorageClassesByCluster(clusterId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.SC.GET-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.allStorageClass = !!data.data?.StorageClasses ? data.data.StorageClasses : [];
      }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SC.GET-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
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
     this.showNoNonDefaultNamespaceAlert = this.namespaceList.length <= 0;
     this.getAllStorageClasses(clusterId);
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  checkCanEditable(labels: JSON): boolean {
    if(!!labels){
      for (const key of Object.keys(labels)) {
        if(key === 'map.tanzu.vmware.com/component'){
          return false;
        }
      }
    }
    return true;
  }

  showCreatePVC(): void {
    this.getAllHealthyClusterIds();
    this.createPVCOpen = true;
  }

  doCreatePVCFinish(): void {
    const pvcParam: CreatePersistentVolumeClaim = {
      name: this.basicPVCForm.value.name,
      namespace: this.basicPVCForm.value.namespace,
      spec: {
        accessModes: [this.basicPVCForm.value.accessMode],
        resources: {
          requests: {
            storage: this.basicPVCForm.value.request + 'Gi'
          },
          limits: {
            storage: !!this.basicPVCForm.value.limit ? this.basicPVCForm.value.limit + 'Gi' : null
          }
        },
        storageClass: this.basicPVCForm.value.storageClass,
        volumeMode: this.basicPVCForm.value.volumeMode
      },
      cluster: this.basicPVCForm.value.cluster,
    };
    this.volumeService.createPVC(pvcParam).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.PVC.CREATE-PVC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.CreatePersistentVolumeClaim){
        this.toast.show(`${data.data.CreatePersistentVolumeClaim.name} `+this.translate.instant('MESSAGE.VOLUME.PVC.CREATE-PVC-MSG'), this.translate.instant('MESSAGE.VOLUME.PVC.CREATE-PVC-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.PVC.CREATE-PVC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doCreatePVCReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doCreatePVCCancel(): void {
    this.doCreatePVCReset()
  }

  private doCreatePVCReset(): void {
    this.wizardnewpvc.reset();
    this.basicPVCForm.reset();
    this.basicPVCForm.patchValue({
      volumeMode: "Filesystem"
    });
    this.createPVCOpen = false;
    this.clusterList = [];
    this.namespaceList = [];
    this.allStorageClass = [];
  }

  // as pvc's storageclass is string
  AllowVolumeExpansion(pvc: PersistentVolumeClaim): void {
    this.allowVolumeExpansion = false;
    this.volumeService.getAllStorageClassesByCluster(pvc.cluster.id).subscribe(data => {
      if(!!data.data?.StorageClasses){
        for(let sc of data.data.StorageClasses){
          if(sc.name === pvc.spec.storageClass){
            this.allowVolumeExpansion = sc.allowVolumeExpansion;
          }
        }
      }
      this.showUpdatePvc = true;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SC.GET-SC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  showUpdatePvcVolume(pvc: PersistentVolumeClaim): void {
    this.AllowVolumeExpansion(pvc);
    this.selectedPVC = pvc;
    this.selectedPVCName = pvc.name;
    this.currentVolumeSize = Number(pvc.spec.resources.requests.storage.split('Gi')[0]);
    this.updatePVCForm.patchValue({
      request: Number(pvc.spec.resources.requests.storage.split('Gi')[0])
    })
  }

  doUpdatePvcVolumeReset(): void {
    this.updatePVCForm.reset();
    this.currentVolumeSize = 0;
    this.selectedPVC = null;
    this.selectedPVCName = '';
  }

  doUpdatePvcVolume(pvc: PersistentVolumeClaim): void {
    this.showUpdatePvc = false;
    const pvcParam: UpdatePersistentVolumeClaim = {
      name: pvc.name,
      namespace: pvc.namespace,
      spec: {
        resources: {
          requests: {
            storage: this.updatePVCForm.value.request + 'Gi'
          }
        }
      },
      cluster: pvc.cluster.id,
    };
    this.volumeService.updatePVC(pvcParam).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.PVC.UPDATE-PVC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.UpdatePersistentVolumeClaim){
        this.toast.show(`${data.data.UpdatePersistentVolumeClaim.name} `+this.translate.instant('MESSAGE.VOLUME.PVC.UPDATE-PVC-MSG'), this.translate.instant('MESSAGE.VOLUME.PVC.UPDATE-PVC-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.PVC.UPDATE-PVC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doUpdatePvcVolumeReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  showRemoveDialog(pvc: PersistentVolumeClaim): void {
    this.selectedPVC = pvc;
    this.selectedPVCName = pvc.name;
    this.removeDialog = true;
  }

  deletePVCByName(pvc: PersistentVolumeClaim): void {
    this.removeDialog = false;
    this.volumeService.deletePVC(pvc.name, pvc.cluster.id, pvc.namespace).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.PVC.DELETE-PVC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.DeletePersistentVolumeClaim){
          this.toast.show(`${data.data.DeletePersistentVolumeClaim.name} `+this.translate.instant('MESSAGE.VOLUME.PVC.DELETE-PVC-MSG'), this.translate.instant('MESSAGE.VOLUME.PVC.DELETE-PVC-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.PVC.DELETE-PVC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

}
