import { Component, ViewChild } from '@angular/core';
import {
  CreateLoadBalancer, DeleteConfirmationAction,
  DeleteObjectEvent,
  LoadBalancer,
  LoadBalancerType,
  UpdateLoadBalancer
} from '../../../classes';
import { LoadbalancerService } from '../../../services/loadbalancer.service';
import { ClrDatagridPagination, ClrDatagridStateInterface, ClrWizard } from '@clr/angular';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { AuthService } from '../../../services/auth.service';
import { CustomValidatorDirective, nameValidator } from '../../../utils/validators/custom-validator.directive';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from '@ngx-translate/core';
import { OrganizationService } from '../../../services/organization.service';
import {
  DELETEOBJECTBUTTON,
  DELETEOBJECTTYPE,
  CONFIRMATIONSTATE,
  FEATUREGATES
} from '../../../utils/static-entries/shared.const';
import { DeleteObjectComponent } from '../../shared-components/delete-object/delete-object.component';
import { CheckStatus } from '../../../utils/static-entries/shared.classes';
import { checkCerts, certCheckResult } from "../../../utils/operator/check-certs";

@Component({
  selector: 'app-lb',
  templateUrl: './loadbalancer.component.html',
  styleUrls: ['./loadbalancer.component.css']
})
export class LoadbalancerComponent {
  featureGates = FEATUREGATES

  constructor (
    private lbService: LoadbalancerService,
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
  lbs: LoadBalancer[] = [];
  selectedLB: LoadBalancer;
  noAVILBAlert = false;
  total = 0;
  showSearch = true;
  searchName: string;
  loading = false;

  addGlbOpen = false;
  editGlbOpen = false;

  checkStatus: CheckStatus = {
    checked: false,
    checkSuccess: false
  };

  certResult: certCheckResult[] = [];

  @ViewChild('wizardglb') wizardnewglb: ClrWizard;
  readonly LoadBalancerType = LoadBalancerType;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  readonly endpointPattern = CustomValidatorDirective.endpointPattern;
  basicGlbForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    type: new FormControl(LoadBalancerType.AVI, Validators.required),
    endpoint: new FormControl('', [Validators.required, Validators.pattern(this.endpointPattern)]),
    caCert: new FormControl(''),
    skipTLS: new FormControl(false),
    user: new FormControl(''),
    password: new FormControl(''),
    seGroup: new FormControl('Default-Group'),
    routeAviNetwork: new FormControl('')
  });

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardeditglb') wizardeditglb: ClrWizard;
  editGlbForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    type: new FormControl('', Validators.required),
    endpoint: new FormControl('', [Validators.required, Validators.pattern(this.endpointPattern)]),
    caCert: new FormControl(''),
    skipTLS: new FormControl(false),
    user: new FormControl(''),
    password: new FormControl(''),
    seGroup: new FormControl('Default-Group'),
    routeAviNetwork: new FormControl('')
  });

  refresh (state?: ClrDatagridStateInterface): void {
    this.loading = true;
    const filters = [{ key: 'organization', value: [this.orgId] }];
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.lbService.getAllLB(state?.page?.current ? state.page.current - 1 : 0, state?.page?.size ? state.page.size : Number(this.pagination.pageSize), filters).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.LOADBALANCER.GET-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.LoadBalancers?.edge?.nodes) {
        this.lbs = data.data.LoadBalancers.edge.nodes;
        const typeList = [];
        for (const lb of this.lbs) {
          typeList.push(lb.type)
        }
        this.noAVILBAlert = !typeList.includes(LoadBalancerType.AVI);
      } else {
        this.lbs = [];
        this.noAVILBAlert = true;
      }
      this.selectedLB = null;
      this.total = data?.data?.LoadBalancers?.totalCount ? data.data.LoadBalancers.totalCount : 0;
      this.certResult = checkCerts(this.lbs);
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.LOADBALANCER.GET-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => { this.loading = false });
  }

  showAddGlb (): void {
    this.addGlbOpen = true;
  }

  resetCheckStatus (): void {
    this.checkStatus.checkSuccess = false;
    this.checkStatus.checked = false;
  }

  showEditGLb (selectedGlb: LoadBalancer): void {
    this.selectedLB = selectedGlb;
    this.editGlbForm.setValue({
      name: selectedGlb.name,
      description: selectedGlb.description,
      type: selectedGlb.type,
      endpoint: selectedGlb.endpoint,
      caCert: selectedGlb.caCert,
      skipTLS: selectedGlb.insecureSkipTLSVerify,
      user: selectedGlb?.authUser,
      password: '',
      routeAviNetwork: selectedGlb.spec?.preferredNetwork ? selectedGlb.spec.preferredNetwork : '',
      seGroup: selectedGlb.spec?.seGroup ? selectedGlb.spec?.seGroup : 'Default-Group'
    });
    this.editGlbOpen = true;
  }

  checkCreateGlbAuth (): void {
    const input: CreateLoadBalancer = {
      name: this.basicGlbForm.value.name,
      type: LoadBalancerType.AVI,
      endpoint: this.basicGlbForm.value.endpoint,
      caCert: !this.basicGlbForm.value.skipTLS ? this.basicGlbForm.value.caCert : null,
      insecureSkipTLSVerify: this.basicGlbForm.value.skipTLS,
      authUser: this.basicGlbForm.value.user,
      authPassword: this.basicGlbForm.value.password,
      organization: this.orgId,
      validate: true,
      dryrun: true
    };
    this.lbService.validateCreateLB(input).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.CHECKERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.LOADBALANCER.CHECK-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.checkStatus.checkSuccess = !!data.data;
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.LOADBALANCER.CHECK-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.checkStatus.checked = true;
    });
  }

  checkUpdateGlbAuth (): void {
    const input: UpdateLoadBalancer = {
      id: this.selectedLB?.id,
      name: this.selectedLB?.name,
      endpoint: this.editGlbForm.value.endpoint,
      caCert: !this.editGlbForm.value.skipTLS ? this.editGlbForm.value.caCert : null,
      insecureSkipTLSVerify: this.editGlbForm.value.skipTLS,
      authUser: this.editGlbForm.value.user,
      authPassword: this.editGlbForm.value.password,
      validate: true,
      dryrun: true
    };
    this.lbService.validateUpdateLB(input).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.CHECKERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.LOADBALANCER.CHECK-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.checkStatus.checkSuccess = !!data.data;
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.LOADBALANCER.CHECK-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.checkStatus.checked = true;
    });
  }

  private doGlbReset (): void {
    this.wizardnewglb.reset();
    this.basicGlbForm.reset();
    this.basicGlbForm.patchValue({
      type: LoadBalancerType.AVI,
      skipTLS: false,
      seGroup: 'Default-Group'
    })
    this.editGlbForm.reset();
    this.addGlbOpen = false;
    this.resetCheckStatus();
  }

  doGlbFinish (): void {
    const param: CreateLoadBalancer = {
      description: this.basicGlbForm.value.description,
      name: this.basicGlbForm.value.name,
      type: this.basicGlbForm.value.type,
      endpoint: this.basicGlbForm.value.endpoint,
      caCert: !this.basicGlbForm.value.skipTLS ? this.basicGlbForm.value.caCert : null,
      insecureSkipTLSVerify: this.basicGlbForm.value.skipTLS,
      authUser: this.basicGlbForm.value.user,
      authPassword: this.basicGlbForm.value.password,
      spec: {
        avi: {
          preferredNetwork: this.basicGlbForm.value.routeAviNetwork,
          seGroup: this.basicGlbForm.value.seGroup
        }
      },
      organization: this.orgId
    };
    this.lbService.createLB(param).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            if (each.message === 'object not found') {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.ADDERROR') + ' 设置的路由网络名称在AVI中不存在！', this.translate.instant('MESSAGE.LOADBALANCER.ADD-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            } else {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.ADDERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.LOADBALANCER.ADD-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            }
          }
        }
      }
      if (data.data?.CreateLoadBalancer) {
        this.toast.show(`${data.data.CreateLoadBalancer.name} ` + this.translate.instant('MESSAGE.LOADBALANCER.ADD-LB-MSG'), this.translate.instant('MESSAGE.LOADBALANCER.ADD-LB-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.LOADBALANCER.ADD-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.doGlbReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doGlbCancel (): void {
    this.doGlbReset();
  }

  doEditGlbFinish (): void {
    const param: UpdateLoadBalancer = {
      id: this.selectedLB?.id,
      description: this.editGlbForm.value.description,
      name: this.editGlbForm.value.name,
      endpoint: this.editGlbForm.value.endpoint,
      caCert: !this.editGlbForm.value.skipTLS ? this.editGlbForm.value.caCert : null,
      insecureSkipTLSVerify: this.editGlbForm.value.skipTLS,
      authUser: this.editGlbForm.value.user,
      spec: {
        avi: {
          preferredNetwork: this.editGlbForm.value.routeAviNetwork
        }
      }
    };
    if (this.editGlbForm.value.password) {
      param.authPassword = this.editGlbForm.value.password
    }
    this.lbService.updateLB(param).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            if (each.message === 'object not found') {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.ADDERROR') + ' 设置的路由网络名称在AVI中不存在！', this.translate.instant('MESSAGE.LOADBALANCER.ADD-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            } else {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.ADDERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.LOADBALANCER.ADD-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            }
          }
        }
      }
      if (data.data?.UpdateLoadBalancer) {
        this.toast.show(`${data.data.UpdateLoadBalancer.name} ` + this.translate.instant('MESSAGE.LOADBALANCER.UPDATE-LB-MSG'), this.translate.instant('MESSAGE.LOADBALANCER.UPDATE-LB-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.LOADBALANCER.UPDATE-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.doGlbReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doEditGlbCancel (): void {
    this.doGlbReset();
  }

  showDeleteLB (selectedLB: LoadBalancer): void {
    this.selectedLB = selectedLB;
    const deleteObjectEvent: DeleteObjectEvent = {
      title: this.translate.instant('LBMGMT.ACTION.DELETELB'),
      message: this.translate.instant('LBMGMT.ACTION.DELETEPAGEDESC'),
      formInputLabel: this.translate.instant('SHARED.DELETE.DELETEDESC'),
      alerts: [],
      objectName: selectedLB.name,
      objectType: DELETEOBJECTTYPE.LB,
      button: DELETEOBJECTBUTTON.CONFIRM_CANCEL
    };
    this.deleteDialogView.open(deleteObjectEvent);
  }

  deleteLB (event: DeleteConfirmationAction): void {
    if (event.objectType === DELETEOBJECTTYPE.LB && event.state === CONFIRMATIONSTATE.CONFIRMED) {
      this.lbService.deleteLB(this.selectedLB?.id).subscribe(data => {
        if (data.errors) {
          for (const each of data.errors) {
            if (each.message) {
              if (each.message) {
                if (each.message === 'object is busy') {
                  this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ' 请先解除集群供应商的关联后再执行此操作！', this.translate.instant('MESSAGE.LOADBALANCER.DELETE-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
                } else {
                  this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.LOADBALANCER.DELETE-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
                }
              }
            }
          }
        }
        if (data.data?.DeleteLoadBalancer) {
          this.toast.show(`${data.data.DeleteLoadBalancer.name} ` + this.translate.instant('MESSAGE.LOADBALANCER.DELETE-LB-MSG'), this.translate.instant('MESSAGE.LOADBALANCER.DELETE-LB-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
        }
      }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.LOADBALANCER.DELETE-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
        if (this.pagination.currentPage === 1) {
          this.refresh();
        } else {
          this.pagination.currentPage = 1;
        }
      });
    }
  }

  closeNoLBAlert (): void {
    this.noAVILBAlert = false;
  }
}
