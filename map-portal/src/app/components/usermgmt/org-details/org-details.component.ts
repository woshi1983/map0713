import {Component, OnInit, ViewChild} from '@angular/core';
import {AuthService} from "../../../services/auth.service";
import {ActivatedRoute, Router} from "@angular/router";
import {OrganizationService} from "../../../services/organization.service";
import {handleObject, Organization, UpdateOrganization} from "../../../classes";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {ClrWizard} from "@clr/angular";
import {FormControl, FormGroup, Validators} from "@angular/forms";
import {CustomValidatorDirective, nameValidator} from "../../../utils/validators/custom-validator.directive";
import { CookieService } from 'ngx-cookie';
import {EventbusService} from "../../../services/eventbus.service";

@Component({
  selector: 'app-org-details',
  templateUrl: './org-details.component.html',
  styleUrls: ['./org-details.component.css']
})
export class OrgDetailsComponent implements OnInit {

  get canWrite(): boolean {
    return this.authService.can('NULL', handleObject.ORGANIZATION);
  }

  constructor(
    private cookie: CookieService,
    private route: ActivatedRoute,
    private router: Router,
    public authService: AuthService,
    private orgService: OrganizationService,
    private toast: ToastrService,
    private translate: TranslateService,
    private eventbusService: EventbusService
  ) { }

  orgList: Organization[] = [];
  org: Organization;
  orgId: string;
  lastOrg: boolean;
  loading = false;

  showUpdateOrgOpen = false;
  removeDialog = false;

  tabLinkStatusList = {
    'memberActive': true,
    'quotaActive': false
  }

  @ViewChild('wizardupdateorg') wizardupdateorg: ClrWizard;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
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
    this.refresh();
  }

  refresh(): void {
    this.loading = true;
    this.orgId = this.route.snapshot.paramMap.get('id');
    this.orgService.getOrgById(this.orgId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.ORG.GET-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.Organization) {
        this.org = data.data.Organization;
      } else {
        this.org = null;
      }
      this.getOrgCount();
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.ORG.GET-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.loading = false;
    });
  }

  backToOrg(): void {
    this.router.navigate(['portal/usermgmt/org']);
  }

  showUpdateOrg(): void {
     this.updateOrgForm.setValue({
       name: this.org.name,
       description: this.org.description
     })
    this.updateResourceQuotaForm.setValue({
      cpu: this.org.quota.total.cpus,
      memory: this.org.quota.total.memoryMi / 1024,
      storage: this.org.quota.total.storageGi,
    })
     this.showUpdateOrgOpen = true;
  }

  doUpdateOrgReset(): void {
    this.wizardupdateorg.reset();
    this.updateOrgForm.reset();
    this.updateResourceQuotaForm.reset();
    this.showUpdateOrgOpen = false;
  }

  doUpdateOrgCancel(): void {
    this.doUpdateOrgReset();
  }

  doUpdateOrgFinish(): void {
    const param: UpdateOrganization = {
      id: this.orgId,
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
        if(this.org.name != this.updateOrgForm.value.name){
          //check current selected org is changed?
          const currentOrgId = this.orgService.getSelectedOrgId;
          if(currentOrgId === this.orgId){
              // set org in cookie and refresh current page
              this.orgService.setSelectOrg(this.orgId, this.updateOrgForm.value.name);
          }
          // refresh org list in navigate
          this.eventbusService.refreshUserOrgEmitter.emit(true);
        }
        this.toast.show(`${data.data.UpdateOrganization.name} `+this.translate.instant('MESSAGE.ORG.UPDATE-ORG-MSG'), this.translate.instant('MESSAGE.ORG.UPDATE-ORG-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.ORG.UPDATE-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doUpdateOrgReset();
      this.refresh();
    });
  }

  showDeleteOrg(): void {
    this.orgService.getAllOrgNames().subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.ORG.GET-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.Organizations?.edge?.nodes) {
        this.orgList = data.data.Organizations.edge.nodes;
        this.removeDialog = true;
      } else {
        this.orgList = [];
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.ORG.GET-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
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

  // only superadmin can delete org
  deleteOrgById(): void {
    this.removeDialog = false;
    this.orgService.deleteOrg(this.orgId).subscribe(data => {
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
      this.eventbusService.refreshUserOrgEmitter.emit(true);
    });
  }



}
