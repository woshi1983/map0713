import { Component, OnInit, ViewChild } from '@angular/core';
import { AuthService } from 'app/services/auth.service';
import { ToastContainerDirective, ToastrService } from 'ngx-toastr';
import { handleObject, ModalEvent } from '../../../classes';
import { AboutDialogComponent } from '../about-dialog/about-dialog.component';
import { ChangePasswordComponent } from '../change-password/change-password.component';
import { UserinfoSettingComponent } from '../userinfo-setting/userinfo-setting.component';
import { SubscribeMapdLogsComponent } from '../subscribe-mapd-logs/subscribe-mapd-logs.component';
import { OrganizationService } from '../../../services/organization.service';
import { MessagesService } from '../../../services/message.service';
import { TranslateService } from '@ngx-translate/core';
import { FEATUREGATES, ALERTTYPE } from '../../../utils/static-entries/shared.const';

@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']
})
export class HomeComponent implements OnInit {
  featureGates = FEATUREGATES
  collapsed = false;
  orgId: string;

  get hideClusterProvider (): boolean {
    return !this.authService.can(this.orgId, handleObject.CLUSTER_PROVIDER);
  }

  get hideUser (): boolean {
    return !this.authService.can(this.orgId, handleObject.USER);
  }

  get hideAuditlog (): boolean {
    return !this.authService.can(this.orgId, handleObject.AUDITLOG);
  }

  get hideOverview (): boolean {
    return !this.authService.can(this.orgId, handleObject.OVERVIEW);
  }

  @ViewChild(ToastContainerDirective, { static: true }) toastContainer: ToastContainerDirective;

  @ViewChild(AboutDialogComponent)
  aboutDialogView: AboutDialogComponent;

  @ViewChild(ChangePasswordComponent)
  changePwdView: ChangePasswordComponent;

  @ViewChild(UserinfoSettingComponent)
  accountSettingView: UserinfoSettingComponent;

  @ViewChild(SubscribeMapdLogsComponent)
  mapdLogsView: SubscribeMapdLogsComponent;

  constructor (
    public authService: AuthService,
    private toast: ToastrService,
    private orgService: OrganizationService,
    private messageService: MessagesService,
    private translate: TranslateService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
    this.checkOrgQuota();
  }

  ngOnInit (): void {
    this.toast.overlayContainer = this.toastContainer;
  }

  // check org quota for alert
  checkOrgQuota (): void {
    const alertFlagList = [0, 0, 0];
    const alertMessageList = ['', '', ''];
    let alertType = ALERTTYPE.INFO;
    let alertMessage: string;
    this.orgService.getOrgQuotaById(this.orgId).subscribe(data => {
      if (data?.data?.Organization?.quota) {
        const cpuPercent = Number((data.data.Organization.quota.used.cpus / data.data.Organization.quota.total.cpus)) * 100;
        const cpuUsagePercent = data.data.Organization.quota.total.cpus !== 0 ? Number(cpuPercent.toFixed(2)) : 0;
        const memPercent = Number((data.data.Organization.quota.used.memoryMi / data.data.Organization.quota.total.memoryMi)) * 100;
        const memoryUsagePercent = data.data.Organization.quota.total.memoryMi !== 0 ? Number(memPercent.toFixed(2)) : 0;
        const storagePercent = Number((data.data.Organization.quota.used.storageGi / data.data.Organization.quota.total.storageGi)) * 100;
        const storageUsagePercent = data.data.Organization.quota.total.storageGi !== 0 ? Number(storagePercent.toFixed(2)) : 0;
        alertFlagList[0] = cpuUsagePercent >= 85 ? cpuUsagePercent >= 95 ? 2 : 1 : 0;
        alertFlagList[1] = memoryUsagePercent >= 75 ? memoryUsagePercent >= 85 ? 2 : 1 : 0;
        alertFlagList[2] = storageUsagePercent >= 80 ? storageUsagePercent >= 90 ? 2 : 1 : 0;
        if (alertFlagList[0] + alertFlagList[1] + alertFlagList[2] > 0) {
          alertType = alertFlagList.find(x => x === 2) ? ALERTTYPE.DANGER : ALERTTYPE.WARNING;
          for (let i = 0; i < 3; i++) {
            if (alertFlagList[i] === 0) {
              continue;
            }
            switch (i) {
              case 0:
                alertMessageList[i] = alertFlagList[0] === 1 ? this.translate.instant('MESSAGE.ORG.CPUQUOTA-MIN') : this.translate.instant('MESSAGE.ORG.CPUQUOTA-MAX');
                break;
              case 1:
                alertMessageList[i] = alertFlagList[1] === 1 ? this.translate.instant('MESSAGE.ORG.MEMQUOTA-MIN') : this.translate.instant('MESSAGE.ORG.MEMQUOTA-MAX');
                break;
              case 2:
                alertMessageList[i] = alertFlagList[2] === 1 ? this.translate.instant('MESSAGE.ORG.STORAGEQUOTA-MIN') : this.translate.instant('MESSAGE.ORG.STORAGEQUOTA-MAX');
                break;
            }
          }
          alertMessage = this.translate.instant('MESSAGE.ORG.PROJECT') + alertMessageList.filter(Boolean).join(', ') + '。';
          this.messageService.sendMessage(alertMessage, alertType);
        }
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.ORG.GET-ORG-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }))
  }

  // Open modal dialog
  openModal (event: ModalEvent): void {
    switch (event.modalName) {
      case 'USER_PROFILE':
        this.accountSettingView.open();
        break;
      case 'CHANGE_PWD':
        this.changePwdView.open();
        break;
      case 'ABOUT':
        this.aboutDialogView.open();
        break;
      case 'MAPD_LOGS':
        this.mapdLogsView.open();
        break;
      default:
        break;
    }
  }
}
