import {Component, OnInit, ViewChild, Input} from '@angular/core';
import {
  Cluster,
  Backup,
  BackupStorageLocation,
  Namespace,
  CreateBackup,
  Restore,
  ClusterBackupSchedule,
  CreateClusterBackupSchedule,
  UpdateClusterBackupSchedule, handleObject
} from '../../../../classes';
import { AuthService} from '../../../../services/auth.service';
import {ClusterBackupService} from "../../../../services/clusterbackup.service";
import {BackupLocationService} from "../../../../services/backuplocation.service";
import {ClrDatagridPagination, ClrDatagridStateInterface, ClrWizard} from '@clr/angular';
import {FormArray, FormControl, FormGroup, Validators} from "@angular/forms";
import {ClustersService} from "../../../../services/clusters.service";
import {ClusterRestoreService} from "../../../../services/clusterrestore.service";
import {CustomValidatorDirective, nameValidator, cronValidator} from "../../../../utils/validators/custom-validator.directive";
import {ActivatedRoute} from "@angular/router";
import {namespacediffset, timeTransit} from "../../../../utils/operator";
import { ToastrService } from 'ngx-toastr';
import {TranslateService} from "@ngx-translate/core";
import {InstanceService} from "../../../../services/instance.service";
import {BackupScheduleService} from "../../../../services/backup-schedule.service";
import {OrganizationService} from "../../../../services/organization.service";
import {FEATUREGATES} from "../../../../utils/static-entries/shared.const";

@Component({
  selector: 'app-cluster-backups',
  templateUrl: './backups-restores.component.html',
  styleUrls: ['./backups-restores.component.css']
})
export class BackupsRestoresComponent implements OnInit {

  featureGates = FEATUREGATES

  //admin and user role can handle cluster resources
  get canWriteResources(): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER_RESOURCE);
  }

  @Input() clusterId: string;

  constructor(
    public authService: AuthService,
    private clusterBackupService: ClusterBackupService,
    private clusterService: ClustersService,
    private clusterRestoreService: ClusterRestoreService,
    private backupLocationService: BackupLocationService,
    private route: ActivatedRoute,
    private toast: ToastrService,
    private translate: TranslateService,
    private instanceService: InstanceService,
    private orgService: OrganizationService,
    private backupScheduleService: BackupScheduleService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  orgId: string;
  allClusterBackups: Backup[] = [];
  selectedBackup: Backup;
  selectedBackupName: string;
  allBackupsScheduler: ClusterBackupSchedule[] = [];
  selectedBackupScheduler: ClusterBackupSchedule;
  selectedSchedulerName: string;
  allClusterRestores: Restore[] = [];
  selectedRestore: Restore;

  clusterList: Cluster[] = [];
  allBackupLocation: BackupStorageLocation[] = [];
  clusterBackupLocationList: BackupStorageLocation[] = [];
  namespaceList: Namespace[] = [];
  componentList: {namespace: string, component: Component}[] = [];
  uniqComponentList: Component[] = [];

  showBackupSearch = false;
  searchBackupName: string;
  showBackupSchedulerSearch = false;
  searchBackupSchedulerName: string;
  showRestoreSearch = false;
  searchRestoreName: string;
  removeBackupDialog = false;
  removeSchedulerDialog = false;
  createBackupOpen = false;
  showBackupDialog = false;
  showUpdateSchedulerDialog = false;
  showBackupSchedulerDialog = false;
  createRestoreOpen = false;

  backupLoading = true;
  backupSchedulerLoading = true;
  scheduleTotal = 0;
  restoreLoading = true;
  loadingNamespace = true;

  disableRemove = false;


  @ViewChild('backupPagination') backupPagination: ClrDatagridPagination;
  @ViewChild('schedulerPagination') schedulerPagination: ClrDatagridPagination;
  @ViewChild('restorePagination') restorePagination: ClrDatagridPagination;
  @ViewChild('wizardbackup') wizardnewbackup: ClrWizard;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  basicBackupsForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(20), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    backupLocation: new FormControl('', [Validators.required]),
    backupMode: new FormControl('ONETIME'),
    backupCron: new FormControl('')
  });
  backupParamsForm = new FormGroup({
    includeNamespaces: new FormControl(''),
    enableVolumeBackup: new FormControl(true),
    includeComponent: new FormControl(''),
    backupTTL: new FormControl(30)
  });
  @ViewChild('wizardrestore') wizardnewrestore: ClrWizard;
  restoreNsMapping = new FormArray([]);
  restoreParamsForm = new FormGroup({
    includeNamespaces: new FormControl(''),
    nsMapping: this.restoreNsMapping
  });
  deleteBackupForm = new FormGroup({
    confirm: new FormControl('', [Validators.required]),
  });
  updateBackupSchedulerForm = new FormGroup({
    backupCron: new FormControl('', [Validators.required, cronValidator()])
  })
  deleteSchedulerForm = new FormGroup({
    confirm: new FormControl('', [Validators.required]),
  });


  ngOnInit(): void {
  }

  refreshBackup(): void {
    this.backupLoading = true;
    const filters = [];
    filters.push({ key: 'cluster', value: [this.clusterId] });
    if(!!this.searchBackupName){
      filters.push({key: 'name', value: [this.searchBackupName]});
    }
    this.clusterBackupService.getClusterBackups(filters).subscribe(data => {
      if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.GET-BACKUP-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
      }
      this.allClusterBackups = !!data.data?.Backups ? data.data.Backups : [];
      this.selectedBackup = null;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.GET-BACKUP-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.backupLoading = false);
  }

  refreshBackupScheduler(state?: ClrDatagridStateInterface): void {
    this.backupSchedulerLoading = true;
    const filters = [];
    filters.push({ key: 'cluster', value: [this.clusterId] });
    if(!!this.searchBackupSchedulerName){
      filters.push({key: 'name', value: [this.searchBackupSchedulerName]});
    }
    this.backupScheduleService.getClusterBackupSchedule(!!state?.page?.current ? state.page.current - 1 : 0, !!state?.page?.size ? state.page.size : Number(this.schedulerPagination.pageSize), filters).subscribe(data => {
      if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.GET-SCHEDULE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
      }
      this.allBackupsScheduler = !!data.data?.ClusterBackupSchedules?.edge?.nodes ? data.data.ClusterBackupSchedules.edge.nodes : [];
      this.selectedBackupScheduler = null;
      this.scheduleTotal = !!data.data?.ClusterBackupSchedules?.totalCount ? data.data.ClusterBackupSchedules.totalCount : 0;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.GET-SCHEDULE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.backupSchedulerLoading = false);
  }

  refreshRestore(): void {
    this.restoreLoading = true;
    const filters = [];
    filters.push({ key: 'cluster', value: [this.clusterId] });
    if(!!this.searchRestoreName){
      filters.push({key: 'name', value: [this.searchRestoreName]});
    }
    this.clusterRestoreService.getClusterRestores(filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.GET-RESTORE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.allClusterRestores = !!data.data?.Restores ? data.data.Restores : [];
      this.selectedRestore = null;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.GET-RESTORE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.restoreLoading = false);
  }

  getClusterBackupLocation(): void {
    this.clusterBackupLocationList = [];
    this.backupLocationService.getAvailableBackupLocatons(this.clusterId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.LOCATION.GET-BAKCUPLOCATION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.BackupStorageLocations) {
        for (const each of data.data.BackupStorageLocations){
          this.allBackupLocation = data.data.BackupStorageLocations;
          if (each.status === 'AVAILABLE' && each.spec.accessMode === 'READ_WRITE'){
             this.clusterBackupLocationList.push(each);
          }
        }
        this.namespaceList = namespacediffset(data.data.BackupStorageLocations[0].cluster.namespaces);
        this.getAllComponentByNs();
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.LOCATION.GET-BAKCUPLOCATION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  showCreateBackup(){
    this.getClusterBackupLocation();
    this.createBackupOpen = true;
  }

  getAllComponentByNs(): void {
    let allNamespace: string[] = [];
    for (const each of this.namespaceList){
      allNamespace.push(each.name);
    }
    if(allNamespace.length > 0) {
      const filters = [{key: 'cluster', value: [this.clusterId]}, {key: 'namespace', value: allNamespace}];
      this.instanceService.getInstancesByNamespace(filters).subscribe(data => {
        if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.INSTANCE.GET-INSTANCE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
        }
        if(!!data.data?.Instances?.edge?.nodes){
          for(const instance of data.data.Instances.edge.nodes){
            if(instance.status.state === 'SUCCESS'){
              this.componentList.push({namespace: instance.namespace, component: instance.component})
            }
          }
        }
        // return all component by default
        this.getUniqComponentList({model: []});
      })
    }
  }

  getUniqComponentList($event: any): void {
    const nsList = $event.model;
    this.uniqComponentList = [];
    //fixed: https://forum.ionicframework.com/t/ionic-5-angular-ion-select-does-not-show-the-selected-values/186633/9
    setTimeout(() => {
      this.backupParamsForm.patchValue({
        includeComponent: ''
      })
    }, 0)
    if(this.componentList.length > 0){
      if(!!nsList && nsList.length > 0){
        // filter by ns
        let nsComponentList = [];
        for(const ns of nsList){
          const nsComponent = this.componentList.filter(x => x.namespace === ns)
          nsComponentList = nsComponentList.concat(nsComponent);
        }
        // remove duplicate
        for(const item of nsComponentList){
          if(!this.uniqComponentList.find(x => x['id'] === item.component['id'])){
            this.uniqComponentList.push(item.component)
          }
        }
      } else {
        // return all ns component
        for(const item of this.componentList){
          if(!this.uniqComponentList.find(x => x['id'] === item.component['id'])){
            this.uniqComponentList.push(item.component)
          }
        }
      }
    }
  }

  public doCreateBackupFinish(): void {
    let selectorNamespaceList = [];
    if(!!this.backupParamsForm.value.includeNamespaces){
      selectorNamespaceList = this.backupParamsForm.value.includeNamespaces;
    } else {
      for (const each of this.namespaceList){
        selectorNamespaceList.push(each.name);
      }
    }
    let selectorComponentList = [];
    if(!!this.backupParamsForm.value.includeComponent){
      for(const component of this.backupParamsForm.value.includeComponent)
      selectorComponentList.push(component.id);
    }
    switch (this.basicBackupsForm.value.backupMode) {
      case 'ONETIME':
        const backupInput: CreateBackup = {
          name: this.basicBackupsForm.value.name,
          spec: {
            useRestic: this.backupParamsForm.value.enableVolumeBackup,
            includedNamespaces: selectorNamespaceList,
            excludedNamespaces: [],
            excludedResources: ['packages.data.packaging.carvel.dev', 'packagemetadatas.data.packaging.carvel.dev', 'events'],
            ttl: `PT${this.backupParamsForm.value.backupTTL * 24}H0M0S`
          },
          location: this.basicBackupsForm.value.backupLocation,
          cluster: this.clusterId
        };
        if(selectorComponentList.length > 0){
          backupInput.spec.components = selectorComponentList
        }
        this.clusterBackupService.createClusterBackup(backupInput).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.CREATE-BACKUP-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data?.data?.CreateBackup){
            this.toast.show(`${data?.data.CreateBackup.name} `+this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.CREATE-BACKUP-MSG'), this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.CREATE-BACKUP-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.CREATE-BACKUP-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          this.doClusterBackupReset();
          if (this.backupPagination.currentPage === 1) {
            this.refreshBackup();
          } else {
            this.backupPagination.currentPage = 1;
          }
        });
        break;
      case 'SCHEDULER':
        const scheduleInput: CreateClusterBackupSchedule = {
          name: this.basicBackupsForm.value.name,
          spec: {
            useRestic: this.backupParamsForm.value.enableVolumeBackup,
            includedNamespaces: selectorNamespaceList,
            excludedNamespaces: [],
            ttl: `PT${this.backupParamsForm.value.backupTTL * 24}H0M0S`
          },
          schedule: this.basicBackupsForm.value.backupCron,
          location: this.basicBackupsForm.value.backupLocation,
          cluster: this.clusterId
        };
        if(selectorComponentList.length > 0){
          scheduleInput.spec.components = selectorComponentList
        }
        this.backupScheduleService.createClusterBackupSchedule(scheduleInput).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.CREATE-SCHEDULE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data?.data?.CreateClusterBackupSchedule){
            this.toast.show(`${data?.data.CreateClusterBackupSchedule.name} `+this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.CREATE-SCHEDULE-MSG'), this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.CREATE-SCHEDULE-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.CREATE-SCHEDULE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          this.doClusterBackupReset();
          if (this.schedulerPagination.currentPage === 1) {
            this.refreshBackupScheduler();
          } else {
            this.schedulerPagination.currentPage = 1;
          }
        });
        break;
      default:
        break;
    }

  }

  public doCreateBackupCancel(): void {
    this.doClusterBackupReset();
  }

  private doClusterBackupReset(): void {
    this.wizardnewbackup.reset();
    this.basicBackupsForm.reset();
    this.basicBackupsForm.patchValue({
      backupMode: 'ONETIME'
    });
    this.backupParamsForm.reset();
    this.backupParamsForm.patchValue({
      enableVolumeBackup: true,
      backupTTL: 30
    });
    this.createBackupOpen = false;
    this.clusterList = [];
    this.namespaceList = [];
    this.componentList = [];
    this.uniqComponentList = [];
    this.setSchedulerVaildtor('ONETIME');
  }

  showBackupInfo(backup: Backup): void {
    this.selectedBackup = backup;
    this.showBackupDialog = true;
  }

  showRemoveBackupDialog(backup: Backup){
    this.disableRemove = false;
    this.selectedBackupName = backup.name;
    if (backup?.location?.spec?.accessMode === 'READ_ONLY'){
       this.disableRemove = true;
    }
    this.removeBackupDialog = true;
  }

  deleteBackupById(backup: Backup){
    this.removeBackupDialog = false;
    this.clusterBackupService.deleteClusterBackup(backup.name, backup.cluster.id).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.DELETE-BACKUP-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data?.data?.DeleteBackup) {
          this.toast.show(`${data?.data.DeleteBackup.name} `+this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.DELETE-BACKUP-MSG'), this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.DELETE-BACKUP-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.DELETE-BACKUP-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.backupPagination.currentPage === 1) {
        this.refreshBackup();
        this.refreshRestore();
      } else {
        this.backupPagination.currentPage = 1;
      }
    });
  }

  setSchedulerVaildtor(mode: string): void {
    if(mode === 'SCHEDULER'){
      this.basicBackupsForm.controls['backupCron'].setValidators([Validators.required, cronValidator()]);
      this.basicBackupsForm.controls['backupCron'].updateValueAndValidity();
    } else {
      this.basicBackupsForm.controls['backupCron'].clearValidators();
      this.basicBackupsForm.controls['backupCron'].updateValueAndValidity();
    }
  }

  showBackupScheduler(scheduler: ClusterBackupSchedule): void {
    this.selectedBackupScheduler = scheduler;
    this.showBackupSchedulerDialog = true;
  }

  showEditBackupScheduler(scheduler: ClusterBackupSchedule): void {
    this.showUpdateSchedulerDialog = true;
    this.updateBackupSchedulerForm.setValue({
      backupCron: scheduler.schedule
    })
  }

  doUpdateBackupScheduler(scheduler: ClusterBackupSchedule): void {
    this.showUpdateSchedulerDialog = false;
    const input: UpdateClusterBackupSchedule = {
      id: scheduler.id,
      schedule: this.updateBackupSchedulerForm.value.backupCron
    }
    this.backupScheduleService.updateClusterBackupSchedule(input).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.UPDATE-SCHEDULE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data?.data?.UpdateClusterBackupSchedule) {
          this.toast.show(`${data?.data.UpdateClusterBackupSchedule.name} `+this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.UPDATE-SCHEDULE-MSG'), this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.UPDATE-SCHEDULE-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.UPDATE-SCHEDULE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.schedulerPagination.currentPage === 1) {
        this.refreshBackupScheduler();
      } else {
        this.schedulerPagination.currentPage = 1;
      }
    });
  }

  showDeleteScheduler(scheduler: ClusterBackupSchedule): void {
    this.selectedSchedulerName = scheduler.name;
    this.removeSchedulerDialog = true;
  }

  deleteSchedulerById(scheduler: ClusterBackupSchedule){
    this.removeSchedulerDialog = false;
    this.backupScheduleService.deleteClusterBackupSchedule(scheduler.id).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.DELETE-SCHEDULE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data?.data?.DeleteClusterBackupSchedule) {
          this.toast.show(`${data?.data.DeleteClusterBackupSchedule.name} `+this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.DELETE-SCHEDULE-MSG'), this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.DELETE-SCHEDULE-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.DELETE-SCHEDULE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.schedulerPagination.currentPage === 1) {
        this.refreshBackupScheduler();
      } else {
        this.schedulerPagination.currentPage = 1;
      }
    });
  }

  showCreateRestore(backup: Backup){
    this.selectedBackupName = backup.name;
    const backupClusterId = backup.cluster.id;
    this.createRestoreOpen = true;
    this.getBackupNamespace(backupClusterId, backup.name);
  }

  getBackupNamespace(cluster: string, backupName: string): void {
    this.loadingNamespace = true;
    this.namespaceList = [];
    this.clusterBackupService.getBackupNamespaces(cluster, backupName).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.GET-BACKUP-NAMESPACE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.namespaceList = !!data.data?.Backup ? data.data.Backup.spec.includedNamespaces : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.GET-BACKUP-NAMESPACE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.loadingNamespace = false;
    });
  }

  public doCreateRestoreFinish(backup: Backup): void {
    let currentDate=new Date();
    const name = backup.name + '-r' + timeTransit(currentDate);
    let nsMappingMap: JSON = JSON.parse(JSON.stringify({}));
    if (!!this.restoreParamsForm.value.nsMapping) {
      for (const k of this.restoreParamsForm.value.nsMapping) {
        nsMappingMap[k.oldNs] = k.newNs;
      }
    }
    const selectorNamespaceList = !!this.restoreParamsForm.value.namespaces ? this.restoreParamsForm.value.namespaces : this.namespaceList;

    this.clusterRestoreService.createClusterRestore(name, backup.cluster.id, backup.name, selectorNamespaceList, nsMappingMap).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.CREATE-RESTORE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data?.data?.CreateRestore){
        this.toast.show(`${data?.data.CreateRestore.name} `+this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.CREATE-RESTORE-MSG'), this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.CREATE-RESTORE-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.BACKUP.CREATE-RESTORE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doClusterRestoreReset();
      if (this.restorePagination.currentPage === 1) {
        this.refreshRestore();
      } else {
        this.restorePagination.currentPage = 1;
      }
    });
  }

  public doCreateRestoreCancel(): void {
    this.doClusterRestoreReset();
  }

  private doClusterRestoreReset(): void {
    this.wizardnewrestore.reset();
    this.restoreParamsForm.reset();
    while (this.restoreNsMapping.length !== 0) {
      this.restoreNsMapping.removeAt(0);
    }
    this.createRestoreOpen = false;
    this.namespaceList = [];
  }

  addNsMapping(): void {
    this.restoreNsMapping.push(new FormGroup({
      oldNs: new FormControl('', Validators.required),
      newNs: new FormControl('', Validators.required)
    }));
  }

  deleteNsMapping(index: number): void{
    this.restoreNsMapping.removeAt(index);
  }

}
