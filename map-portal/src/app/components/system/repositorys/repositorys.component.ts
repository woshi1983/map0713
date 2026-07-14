import { Component, ViewChild } from '@angular/core';
import { AuthService } from '../../../services/auth.service';
import { RepositorysService } from '../../../services/repositorys.service';
import {
  CreateImageRegistry,
  ImageRegistry,
  UpdateImageRegistry,
  ImageRegistryType,
  DeleteObjectEvent, DeleteConfirmationAction
} from '../../../classes';
import { ClrDatagridPagination, ClrDatagridStateInterface, ClrWizard } from '@clr/angular';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { CustomValidatorDirective, nameValidator, CAFormatValidator } from '../../../utils/validators/custom-validator.directive';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from '@ngx-translate/core';
import { OrganizationService } from '../../../services/organization.service';
import { CheckStatus } from '../../../utils/static-entries/shared.classes';
import { DELETEOBJECTBUTTON, DELETEOBJECTTYPE, CONFIRMATIONSTATE } from '../../../utils/static-entries/shared.const';
import { DeleteObjectComponent } from '../../shared-components/delete-object/delete-object.component';
import {environment} from "../../../../environments/environment";
import { checkCerts, certCheckResult } from "../../../utils/operator/check-certs";

@Component({
  selector: 'app-repositorys',
  templateUrl: './repositorys.component.html',
  styleUrls: ['./repositorys.component.css']
})
export class RepositorysComponent {

  appBuild = environment.features.hasOwnProperty('appBuild') ? environment.features.appBuild : true;

  constructor (
    private repositoryService: RepositorysService,
    public authService: AuthService,
    private toast: ToastrService,
    private translate: TranslateService,
    private orgService: OrganizationService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  @ViewChild('deleteObjectComponent')
  deleteDialogView: DeleteObjectComponent;

  orgId: string;
  repositorys: ImageRegistry[] = [];
  selectedRepo: ImageRegistry;
  showSearch = true;
  searchName: string;
  total = 0;
  loading = true;

  addRepoOpen = false;
  editRepoOpen = false;

  checkStatus: CheckStatus = {
    checked: false,
    checkSuccess: false
  };

  certResult: certCheckResult[] = [];

  @ViewChild('wizardrepo') wizardnewrepo: ClrWizard;
  readonly ImageRegistryType = ImageRegistryType;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  readonly endpointPattern = CustomValidatorDirective.endpointPattern;
  basicRepoForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    type: new FormControl(ImageRegistryType.HARBOR, Validators.required),
    endpoint: new FormControl('', [Validators.required, Validators.pattern(this.endpointPattern)]),
    caCert: new FormControl('', CAFormatValidator()),
    pushable: new FormControl(true),
    user: new FormControl(''),
    password: new FormControl('')
  });

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardeditrepo') wizardeditrepo: ClrWizard;
  editRepoForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    type: new FormControl('', Validators.required),
    endpoint: new FormControl('', [Validators.required, Validators.pattern(this.endpointPattern)]),
    caCert: new FormControl('', CAFormatValidator()),
    pushable: new FormControl(''),
    user: new FormControl(''),
    password: new FormControl('')
  });

  refresh (state?: ClrDatagridStateInterface): void {
    this.loading = true;
    const filters = [{ key: 'organization', value: [this.orgId] }];
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.repositoryService.getImageRegistrys(state?.page?.current ? state.page.current - 1 : 0, state?.page?.size ? state.page.size : Number(this.pagination.pageSize), filters).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.REPOSITORY.GET-REPO-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.repositorys = data.data?.ImageRegistries?.edge?.nodes ? data.data.ImageRegistries.edge.nodes : [];
      this.selectedRepo = null;
      this.total = data.data?.ImageRegistries?.totalCount ? data.data.ImageRegistries.totalCount : 0;
      this.certResult = checkCerts(this.repositorys);
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.REPOSITORY.GET-REPO-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => { this.loading = false });
  }

  showAddRepo (): void {
    this.addRepoOpen = true;
  }

  showEditRepo (selectedRepo: ImageRegistry): void {
    this.selectedRepo = selectedRepo;
    this.editRepoForm.setValue({
      name: selectedRepo.name,
      description: selectedRepo.description,
      endpoint: selectedRepo.endpoint,
      caCert: selectedRepo.caCert,
      type: selectedRepo.type,
      user: '',
      password: '',
      pushable: false
    });
    if (selectedRepo.authUser) {
      this.editRepoForm.patchValue({
        pushable: true,
        user: selectedRepo.authUser
      })
    }
    this.editRepoOpen = true;
  }

  setAuthInfoVaildtor (enabled: boolean): void {
    if (enabled) {
      this.basicRepoForm.controls['user'].setValidators(Validators.required);
      this.basicRepoForm.controls['user'].updateValueAndValidity();
      this.basicRepoForm.controls['password'].setValidators(Validators.required);
      this.basicRepoForm.controls['password'].updateValueAndValidity();
    } else {
      this.basicRepoForm.controls['user'].clearValidators();
      this.basicRepoForm.controls['user'].updateValueAndValidity();
      this.basicRepoForm.controls['password'].clearValidators();
      this.basicRepoForm.controls['password'].updateValueAndValidity();
    }
  }

  checkCreateRepoAuth (): void {
    const input: CreateImageRegistry = {
      name: this.basicRepoForm.value.name,
      endpoint: this.basicRepoForm.value.endpoint,
      authUser: this.basicRepoForm.value.user,
      authPassword: this.basicRepoForm.value.password,
      type: 'HARBOR',
      caCert: this.basicRepoForm.value.caCert,
      organization: this.orgId,
      validate: true,
      dryrun: true
    };
    this.repositoryService.validateCreateImageRegistry(input).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.CHECKERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.REPOSITORY.CHECK-REPO-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.checkStatus.checkSuccess = !!data.data;
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.REPOSITORY.CHECK-REPO-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.checkStatus.checked = true;
    });
  }

  resetCheckStatus (): void {
    this.checkStatus.checkSuccess = false;
    this.checkStatus.checked = false;
  }

  setUpdateAuthInfoVaildtor (enabled: boolean): void {
    if (enabled) {
      this.editRepoForm.controls['user'].setValidators(Validators.required);
      this.editRepoForm.controls['user'].updateValueAndValidity();
    } else {
      this.editRepoForm.controls['user'].clearValidators();
      this.editRepoForm.controls['user'].updateValueAndValidity();
    }
  }

  checkUpdateRepoAuth (): void {
    const input: UpdateImageRegistry = {
      id: this.selectedRepo?.id,
      name: this.selectedRepo?.name,
      endpoint: this.editRepoForm.value.endpoint,
      authUser: this.editRepoForm.value.user,
      authPassword: this.editRepoForm.value.password,
      validate: true,
      dryrun: true
    };
    this.repositoryService.validateUpdateImageRegistry(input).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.CHECKERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.REPOSITORY.CHECK-REPO-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.checkStatus.checkSuccess = !!data.data;
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.REPOSITORY.CHECK-REPO-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.checkStatus.checked = true
    });
  }

  private doRepoReset (): void {
    this.wizardnewrepo.reset();
    this.basicRepoForm.reset();
    this.basicRepoForm.patchValue({
      type: ImageRegistryType.HARBOR,
      pushable: true
    });
    this.wizardeditrepo.reset();
    this.editRepoForm.reset();
    this.addRepoOpen = false;
    this.checkStatus = {
      checked: false,
      checkSuccess: false
    }
  }

  doRepoFinish (): void {
    const param: CreateImageRegistry = {
      name: this.basicRepoForm.value.name,
      description: this.basicRepoForm.value.description,
      type: this.basicRepoForm.value.type,
      endpoint: this.basicRepoForm.value.endpoint,
      caCert: this.basicRepoForm.value.caCert ? this.basicRepoForm.value.caCert : null,
      organization: this.orgId
    };
    if (this.basicRepoForm.value.pushable) {
      param.authUser = this.basicRepoForm.value.user;
      param.authPassword = this.basicRepoForm.value.password;
    }
    this.repositoryService.createImageRegistry(param).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.ADDERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.REPOSITORY.ADD-REPO-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.CreateImageRegistry) {
        this.toast.show(`${data.data.CreateImageRegistry.name} ` + this.translate.instant('MESSAGE.REPOSITORY.ADD-REPO-MSG'), this.translate.instant('MESSAGE.REPOSITORY.ADD-REPO-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.REPOSITORY.ADD-REPO-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.doRepoReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doRepoCancel (): void {
    this.doRepoReset();
  }

  doEditRepoFinish (): void {
    const param: UpdateImageRegistry = {
      id: this.selectedRepo?.id,
      name: this.editRepoForm.value.name,
      description: this.editRepoForm.value.description,
      endpoint: this.editRepoForm.value.endpoint,
      caCert: this.editRepoForm.value.caCert ? this.editRepoForm.value.caCert : null
    };
    // pushable change to unpushable
    if (this.editRepoForm.value.pushable) {
      // password updated
      if (this.editRepoForm.value.password) {
        param.authUser = this.editRepoForm.value.user;
        param.authPassword = this.editRepoForm.value.password;
      }
    } else {
      // unpushable, remove user and password
      param.authUser = '';
      param.authPassword = '';
    }
    this.repositoryService.updateImageRegistry(param).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.UPDATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.REPOSITORY.UPDATE-REPO-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.UpdateImageRegistry) {
        this.toast.show(`${data.data.UpdateImageRegistry.name} ` + this.translate.instant('MESSAGE.REPOSITORY.UPDATE-REPO-MSG'), this.translate.instant('MESSAGE.REPOSITORY.UPDATE-REPO-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.REPOSITORY.UPDATE-REPO-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.doRepoReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doEditRepoCancel (): void {
    this.doRepoReset();
  }

  showDeleteRepo (selectedRepo: ImageRegistry): void {
    this.selectedRepo = selectedRepo;
    const deleteObjectEvent: DeleteObjectEvent = {
      title: this.translate.instant('REGISTRYMGMT.ACTION.DELETEREGISTRY'),
      message: this.translate.instant('REGISTRYMGMT.ACTION.DELETEPAGEDESC'),
      formInputLabel: this.translate.instant('SHARED.DELETE.DELETEDESC'),
      alerts: [],
      objectName: selectedRepo.name,
      objectType: DELETEOBJECTTYPE.REPO,
      button: DELETEOBJECTBUTTON.CONFIRM_CANCEL
    };
    this.deleteDialogView.open(deleteObjectEvent);
  }

  deleteRepo (event: DeleteConfirmationAction): void {
    if (event.objectType === DELETEOBJECTTYPE.REPO && event.state === CONFIRMATIONSTATE.CONFIRMED) {
      this.repositoryService.deleteImageRegistry(this.selectedRepo?.id).subscribe(data => {
        if (data.errors) {
          for (const each of data.errors) {
            if (each.message) {
              if (each.message === 'object is busy') {
                this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + this.translate.instant('MESSAGE.REPOSITORY.FORCE-DELETE-REPO-MSG'), this.translate.instant('MESSAGE.REPOSITORY.DELETE-REPO-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
              } else {
                this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.REPOSITORY.DELETE-REPO-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
              }
            }
          }
        }
        if (data.data?.DeleteImageRegistry) {
          this.toast.show(`${data.data.DeleteImageRegistry.name} ` + this.translate.instant('MESSAGE.REPOSITORY.DELETE-REPO-MSG'), this.translate.instant('MESSAGE.REPOSITORY.DELETE-REPO-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
        }
      }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.REPOSITORY.DELETE-REPO-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
        if (this.pagination.currentPage === 1) {
          this.refresh();
        } else {
          this.pagination.currentPage = 1;
        }
      });
    }
  }
}
