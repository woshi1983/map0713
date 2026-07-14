import {Component, OnInit, ViewChild} from '@angular/core';
import {AuthService} from "../../../services/auth.service";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {Organization, handleObject, CreateOrganization, UpdateOrganization} from "../../../classes";
import {ClrDatagridPagination, ClrDatagridStateInterface, ClrWizard, ClrDatagridSortOrder} from "@clr/angular";
import {OrganizationService} from "../../../services/organization.service";
import {CustomValidatorDirective, nameValidator} from "../../../utils/validators/custom-validator.directive";
import {FormControl, FormGroup, Validators} from "@angular/forms";
import {Router} from "@angular/router";
import { CookieService } from 'ngx-cookie';
import {EventbusService} from "../../../services/eventbus.service";

@Component({
  selector: 'app-organization',
  templateUrl: './organization.component.html',
  styleUrls: ['./organization.component.css']
})
export class OrganizationComponent implements OnInit {

  //only superadmin can handle org create/update/delete
  // remove org quota update for org admin
  get canWrite(): boolean {
    return this.authService.can('NULL', handleObject.ORGANIZATION);
  }

  constructor(
    private cookie: CookieService,
    public authService: AuthService,
    private orgService: OrganizationService,
    private toast: ToastrService,
    private translate: TranslateService,
    private router: Router,
    private eventbusService: EventbusService
  ) {}

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  orgList: Organization[] = [];
  selected: Organization;
  selectedOrgId: string;
  selectedOrgName: string;
  lastOrg: boolean = true;

  loading = false;
  total = 0;
  showSearch = true;
  searchName: string;

  showCreateOrgOpen = false;
  showUpdateOrgOpen = false;
  removeDialog = false;

  @ViewChild('wizardorg') wizardneworg: ClrWizard;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  basicOrgForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl('')
  });
  resourceQuotaForm = new FormGroup({
    cpu: new FormControl(0, Validators.min(0)),
    memory: new FormControl(0, Validators.min(0)),
    storage: new FormControl(0, Validators.min(0))
  })
  @ViewChild('wizardupdateorg') wizardupdateorg: ClrWizard;
  updateOrgForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl('')
  });
  updateResourceQuotaForm = new FormGroup({
    cpu: new FormControl(0, Validators.min(0)),
    memory: new FormControl(0, Validators.min(0)),
    storage: new FormControl(0, Validators.min(0))
  })
  deleteOrgForm = new FormGroup({
    confirm: new FormControl('', [Validators.required]),
  });

  ngOnInit(): void {
  }

  refresh(state?: ClrDatagridStateInterface): void {
    this.loading = true;
    const filters = [];
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.orgService.getAllOrgs(!!state?.page?.current ? state.page.current - 1 : 0, !!state?.page?.size ? state.page.size : Number(this.pagination.pageSize), filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.ORG.GET-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.orgList = !!data.data?.Organizations?.edge?.nodes ? data.data.Organizations.edge.nodes : [];
      this.selected = null;
      this.total = !!data.data?.Organizations?.totalCount ? data.data.Organizations.totalCount : 0;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.ORG.GET-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {this.loading = false;});
  }

  showCreateOrg(): void {
    this.showCreateOrgOpen = true;
  }

  doCreateOrgReset(): void {
    this.wizardneworg.reset();
    this.basicOrgForm.reset();
    this.resourceQuotaForm.reset();
    this.resourceQuotaForm.setValue({
      cpu: 0,
      memory: 0,
      storage: 0
    })
    this.showCreateOrgOpen = false;
  }

  doCreateOrgCancel(): void {
    this.doCreateOrgReset();
  }

  doCreateOrgFinish(): void {
    const param: CreateOrganization = {
      name: this.basicOrgForm.value.name,
      description: this.basicOrgForm.value.description,
      quota: {
        cpus: this.resourceQuotaForm.value.cpu,
        memoryMi: this.resourceQuotaForm.value.memory * 1024,
        storageGi: this.resourceQuotaForm.value.storage
      }
    }
    this.orgService.createOrg(param).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.ORG.CREATE-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.CreateOrganization){
        this.toast.show(`${data.data.CreateOrganization.name} `+this.translate.instant('MESSAGE.ORG.CREATE-ORG-MSG'), this.translate.instant('MESSAGE.ORG.CREATE-ORG-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.ORG.CREATE-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doCreateOrgReset();
      this.eventbusService.refreshUserOrgEmitter.emit(true);
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  showUpdateOrg(org: Organization): void {
     this.selectedOrgId = org.id;
     this.selectedOrgName = org.name;
     this.updateOrgForm.setValue({
       name: org.name,
       description: org.description
     })
    this.updateResourceQuotaForm.setValue({
      cpu: org.quota.total.cpus,
      memory: org.quota.total.memoryMi / 1024,
      storage: org.quota.total.storageGi,
    })
     this.showUpdateOrgOpen = true;
  }

  doUpdateOrgReset(): void {
    this.wizardupdateorg.reset();
    this.updateOrgForm.reset();
    this.updateResourceQuotaForm.reset();
    this.showUpdateOrgOpen = false;
    this.selectedOrgId = null;
    this.selectedOrgName = null;
  }

  doUpdateOrgCancel(): void {
    this.doUpdateOrgReset();
  }

  doUpdateOrgFinish(): void {
    const param: UpdateOrganization = {
      id: this.selectedOrgId,
      name: this.updateOrgForm.value.name,
      description: this.updateOrgForm.value.description,
      quota: {
        cpus: this.updateResourceQuotaForm.value.cpu,
        memoryMi: this.updateResourceQuotaForm.value.memory * 1024,
        storageGi: this.updateResourceQuotaForm.value.storage
      }
    }
    this.orgService.updateOrg(param).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.ORG.UPDATE-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.UpdateOrganization){
        // check org name changed?
        if(this.selectedOrgName != this.updateOrgForm.value.name){
          //check current selected org is changed?
          const currentOrgId = this.orgService.getSelectedOrgId;
          if(currentOrgId === this.selectedOrgId){
              // set org in cookie and refresh current page
              this.orgService.setSelectOrg(this.selectedOrgId, this.updateOrgForm.value.name);
          }
          // refresh org list in navigate
          this.eventbusService.refreshUserOrgEmitter.emit(true);
        }
        this.toast.show(`${data.data.UpdateOrganization.name} `+this.translate.instant('MESSAGE.ORG.UPDATE-ORG-MSG'), this.translate.instant('MESSAGE.ORG.UPDATE-ORG-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.ORG.UPDATE-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doUpdateOrgReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  // check last org
  getOrgCount(): void {
    this.orgService.getAllOrgCounts().subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.ORG.GET-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.lastOrg = data.data.Organizations.totalCount <= 1;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.ORG.GET-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
  }

  showDeleteOrg(org: Organization): void {
    this.getOrgCount();
    this.selectedOrgName = org.name;
    this.removeDialog = true;
  }

  // only superadmin can delete org
  deleteOrgById(org: Organization): void {
    this.removeDialog = false;
    this.orgService.deleteOrg(org.id).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             if(each.message === 'object is busy'){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+this.translate.instant('MESSAGE.SHARED.DELETE-ORG-ALERT'), this.translate.instant('MESSAGE.ORG.DELETE-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             } else {
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.ORG.DELETE-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
           }
        }
      }
      if (!!data.data?.DeleteOrganization){
        this.toast.show(`${data.data.DeleteOrganization.name} `+this.translate.instant('MESSAGE.ORG.DELETE-ORG-MSG'), this.translate.instant('MESSAGE.ORG.DELETE-ORG-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
        // delete current org
        const currentOrgId = this.orgService.getSelectedOrgId;
        if(currentOrgId === data.data.DeleteOrganization.id){
          const deleteIndex = this.orgList.findIndex(x => x.id === data.data.DeleteOrganization.id);
          if(deleteIndex > -1){
            this.orgList.splice(deleteIndex, 1);
          }
          let sortedOrgList: Organization[];
          sortedOrgList = this.orgList.sort((n1, n2) => {
                  if(n1.name > n2.name){return 1;}
                  if(n1.name < n2.name){return -1;}
                  return 0
            });
          this.orgService.setSelectOrg(sortedOrgList[0].id, sortedOrgList[0].name);
          window.setTimeout(function(){window.location.reload()}, 1000);
        }
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.ORG.DELETE-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
        this.eventbusService.refreshUserOrgEmitter.emit(true);
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

}
