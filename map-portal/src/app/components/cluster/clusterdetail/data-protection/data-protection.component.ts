import {Component, OnInit, ViewChild, Input} from '@angular/core';
import { AuthService} from '../../../../services/auth.service';
import {ClrDatagridPagination, ClrWizard} from '@clr/angular';
import {FormControl, FormGroup, Validators} from '@angular/forms';
import {
  BackupStorageLocation, BackupStorageLocationAccessMode,
  CreateBackupStorageLocation,
  handleObject
} from "../../../../classes";
import { BackupLocationService } from "../../../../services/backuplocation.service";
import {ActivatedRoute} from "@angular/router";
import {ClustersService} from "../../../../services/clusters.service";
import {CustomValidatorDirective, nameValidator} from '../../../../utils/validators/custom-validator.directive';
import { ToastrService } from 'ngx-toastr';
import {TranslateService} from "@ngx-translate/core";
import {OrganizationService} from "../../../../services/organization.service";

@Component({
  selector: 'app-cluster-data-protection',
  templateUrl: './data-protection.component.html',
  styleUrls: ['./data-protection.component.css']
})
export class DataProtectionComponent implements OnInit {

  //admin and user role can handle cluster resources
  get canWriteResources(): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER_RESOURCE);
  }

  @Input() clusterId: string;
  @Input() clusterStatus: string;

  constructor(
    public authService: AuthService,
    private backupLocationService: BackupLocationService,
    private route: ActivatedRoute,
    private clusterService: ClustersService,
    private toast: ToastrService,
    private orgService: OrganizationService,
    private translate: TranslateService) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  orgId: string;
  allBackupLocation: BackupStorageLocation[] = [];
  selectedBackupLocation: BackupStorageLocation;
  selectedBackupLocationName: string;
  backupLocationOpen = false;

  showLocationSearch = false;
  searchName: string;
  showBackupLocationInfoDialog = false;
  removeDialog = false;
  showConnectAlert = false;
  checkConnectSuccess = false;

  backupLoading = true;
  noBackupLocationAlert = true;

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardbackuplocation') wizardnewbackuplocation: ClrWizard;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  readonly endpointPattern = CustomValidatorDirective.endpointPattern;
  basicBackupLocationForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    backupProvider: new FormControl('minio', [Validators.required]),
  });
  minioBackupLocationForm = new FormGroup({
    endpoint: new FormControl('', [Validators.required, Validators.pattern(this.endpointPattern)]),
    cacert: new FormControl(''),
    accessKey: new FormControl('', Validators.required),
    secretKey: new FormControl('', Validators.required),
    bucket: new FormControl('', Validators.required),
    prefix: new FormControl(''),
    accessMode: new FormControl('READ_WRITE'),
  });
  awsBackupLocationForm = new FormGroup({
    accessKey: new FormControl('', Validators.required),
    secretKey: new FormControl('', Validators.required),
    region: new FormControl('', Validators.required),
    bucket: new FormControl('', Validators.required),
    accessMode: new FormControl('READ_WRITE'),
  });
  deleteBackupLocationForm = new FormGroup({
    confirm: new FormControl('', [Validators.required]),
  });

  ngOnInit(): void {
  }

  refresh(): void{
    this.backupLoading = true;
    const filters = [];
    filters.push({ key: 'cluster', value: [this.clusterId] });
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.backupLocationService.getBackupLocationsOfCluster(filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.LOCATION.GET-BAKCUPLOCATION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.BackupStorageLocations) {
        this.allBackupLocation = data.data.BackupStorageLocations;
        this.selectedBackupLocation = null;
        this.noBackupLocationAlert = false;
      } else {
        this.allBackupLocation = [];
        this.selectedBackupLocation = null;
        this.noBackupLocationAlert = true;
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.LOCATION.GET-BAKCUPLOCATION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.backupLoading = false);
  }

  showCreateBackupLocation(): void {
    this.backupLocationOpen = true;
  }

  checkObjectStorageAuth(): void {
    const input: CreateBackupStorageLocation = {
      name: this.basicBackupLocationForm.value.name,
      spec: {
        provider: 'aws',
        accessMode: BackupStorageLocationAccessMode.READ_ONLY,
        profile: {
          awsAccessKeyID: this.minioBackupLocationForm.value.accessKey,
          awsSecretAccessKey: this.minioBackupLocationForm.value.secretKey
        },
        config: {
          region: 'minio',
          s3ForcePathStyle: true,
          s3Url: this.minioBackupLocationForm.value.endpoint
        },
        objectStorage: {
          bucket: this.minioBackupLocationForm.value.bucket,
          prefix: this.minioBackupLocationForm.value.prefix,
          caCert: this.minioBackupLocationForm.value.cacert
        }
      },
      cluster: this.clusterId,
      validate: true,
      dryrun: true
    };
    this.backupLocationService.validateMinio(input).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.CHECKERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.LOCATION.CHECK-AUTH-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.checkConnectSuccess = !!data.data;
      this.showConnectAlert = true;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.LOCATION.CHECK-AUTH-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  private doBackupLocationReset(): void {
    this.wizardnewbackuplocation.reset();
    this.basicBackupLocationForm.reset();
    this.basicBackupLocationForm.patchValue({
      backupProvider: 'minio'
    });
    this.minioBackupLocationForm.reset();
    this.minioBackupLocationForm.patchValue({
      accessMode: 'READ_WRITE'
    });
    this.awsBackupLocationForm.reset();
    this.awsBackupLocationForm.patchValue({
      accessMode: 'READ_WRITE'
    });
    this.backupLocationOpen = false;
  }

  doBackupLocationFinish(): void {
    const name = this.basicBackupLocationForm.value.name;
    const backupProvider = this.basicBackupLocationForm.value.backupProvider;
    switch (backupProvider) {
      case 'minio':
        const input: CreateBackupStorageLocation = {
          name: name,
          spec: {
            provider: 'aws',
            config: {
              region: 'minio',
              s3ForcePathStyle: true,
              s3Url: this.minioBackupLocationForm.value.endpoint
            },
            objectStorage: {
              bucket: this.minioBackupLocationForm.value.bucket,
              prefix: this.minioBackupLocationForm.value.prefix,
              caCert: this.minioBackupLocationForm.value.cacert
            },
            accessMode: this.minioBackupLocationForm.value.accessMode,
            profile: {
              awsAccessKeyID: this.minioBackupLocationForm.value.accessKey,
              awsSecretAccessKey: this.minioBackupLocationForm.value.secretKey
            }
          },
          cluster: this.clusterId,
          dryrun: false,
          validate: false
        };

        this.backupLocationService.createBackupLocation(input).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.LOCATION.CREATE-LOCATION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data?.data?.CreateBackupStorageLocation) {
            this.toast.show(`${data?.data.CreateBackupStorageLocation.name} `+this.translate.instant('MESSAGE.CLUSTER-DETAILD.LOCATION.CREATE-LOCATION-MSG'), this.translate.instant('MESSAGE.CLUSTER-DETAILD.LOCATION.CREATE-LOCATION-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.LOCATION.CREATE-LOCATION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          this.doBackupLocationReset();
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
        });
        break;
      case 'aws':
       // do aws backup location
       break;
      case 'gcp':
        // do gcp backup location
        break;
      case 'azure':
        // do azure backup location
        break;
      default:
        console.error('Backup Location provider type not support.');
  }
  }

  doBackupLocationCancel(): void {
    this.doBackupLocationReset();
  }

  showBackupLocationDetails(backuplocation: BackupStorageLocation): void {
    this.selectedBackupLocation = backuplocation;
    this.showBackupLocationInfoDialog = true;
  }


  showDeleteBackupLocation(backuplocation: BackupStorageLocation): void {
    this.selectedBackupLocationName = backuplocation.name;
    this.removeDialog = true;
  }

  deleteBackupLocationById(name: string): void {
    this.removeDialog = false;
    this.backupLocationService.deleteBackupLocation(name, this.clusterId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.LOCATION.DELETE-LOCATION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data?.data?.DeleteBackupStorageLocation) {
          this.toast.show(`${data?.data.DeleteBackupStorageLocation.name} `+this.translate.instant('MESSAGE.CLUSTER-DETAILD.LOCATION.DELETE-LOCATION-MSG'), this.translate.instant('MESSAGE.CLUSTER-DETAILD.LOCATION.DELETE-LOCATION-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.LOCATION.DELETE-LOCATION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  closeNoBackupLocationAlert(): void{
    this.noBackupLocationAlert = false;
  }
}
