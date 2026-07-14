import { Component, Input, OnInit } from '@angular/core';
import { AuthService } from "../../../../services/auth.service";
import { ToastrService } from "ngx-toastr";
import { TranslateService } from "@ngx-translate/core";
import { OrganizationService } from "../../../../services/organization.service";
import { QuotaResult } from "../../../../classes";

@Component({
  selector: 'app-org-quota',
  templateUrl: './quota.component.html',
  styleUrls: ['./quota.component.css']
})
export class QuotaComponent implements OnInit {

  // show quota meet alert in home page

  @Input() orgId;
  constructor(
    public authService: AuthService,
    private orgService: OrganizationService,
    private toast: ToastrService,
    private translate: TranslateService
  ) { }

  orgQuota: QuotaResult = {
    cpu: { total: 0 }, memory: { total: 0 }, storage: { total: 0 }
  };
  loading = false;

  ngOnInit(): void {
    this.refresh();
  }

  refresh(): void {
    this.loading = true;
    this.orgService.getOrgQuotaById(this.orgId).subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.ORG.GET-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if(!!data.data?.Organization?.quota){
        if(data.data.Organization.quota.total.cpus !== 0){
          const cpuPercent = Number((data.data.Organization.quota.used.cpus / data.data.Organization.quota.total.cpus)) * 100;
          this.orgQuota.cpu = {
            total: data.data.Organization.quota.total.cpus,
            free: data.data.Organization.quota.total.cpus - data.data.Organization.quota.used.cpus,
            used: data.data.Organization.quota.used.cpus,
            percent: Math.round((cpuPercent + Number.EPSILON) * 100) / 100
          }
        }
        if(data.data.Organization.quota.total.memoryMi !== 0){
          const memPercent = Number((data.data.Organization.quota.used.memoryMi / data.data.Organization.quota.total.memoryMi)) * 100;
          this.orgQuota.memory= {
            total: data.data.Organization.quota.total.memoryMi / 1024,
            free: Math.round(((data.data.Organization.quota.total.memoryMi - data.data.Organization.quota.used.memoryMi) / 1024 + Number.EPSILON) * 1000 ) / 1000,
            used: Math.round((data.data.Organization.quota.used.memoryMi / 1024 + Number.EPSILON) * 1000) / 1000,
            percent: Math.round((memPercent + Number.EPSILON) * 100) / 100
          }
        }
        if(data.data.Organization.quota.total.storageGi !== 0){
          const storagePercent = Number((data.data.Organization.quota.used.storageGi / data.data.Organization.quota.total.storageGi)) * 100;
          this.orgQuota.storage = {
            total: data.data.Organization.quota.total.storageGi,
            free: data.data.Organization.quota.total.storageGi - data.data.Organization.quota.used.storageGi,
            used: data.data.Organization.quota.used.storageGi,
            percent: Math.round((storagePercent + Number.EPSILON) * 100) / 100
          }
        }
      } else {
        this.orgQuota = null;
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.ORG.GET-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false)
  }

}
