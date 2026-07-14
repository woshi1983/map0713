import { Component, Input, OnInit, ViewChild } from '@angular/core';
import { BinaryBuild, BinaryImage, DeleteConfirmationAction, DeleteObjectEvent, handleObject } from '../../../../classes';
import { AuthService } from '../../../../services/auth.service';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from '@ngx-translate/core';
import { BuildService } from '../../../../services/build.service';
import { OrganizationService } from '../../../../services/organization.service';
import { DeleteObjectComponent } from '../../../shared-components/delete-object/delete-object.component';
import { CONFIRMATIONSTATE, DELETEOBJECTBUTTON, DELETEOBJECTTYPE, ALERTTYPE } from '../../../../utils/static-entries/shared.const';

@Component({
  selector: 'app-binary-build',
  templateUrl: './binary-build.component.html',
  styleUrls: ['./binary-build.component.css']
})
export class BinaryBuildComponent implements OnInit {
  // admin and user role can handle cluster resources
  get canWriteResources (): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER_RESOURCE);
  }

  @Input() image: BinaryImage;

  constructor (
    private buildService: BuildService,
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
  allBuilds: BinaryBuild[] = [];
  noHistory: boolean;
  loading = false;

  deleteBuildName: string;
  showBuildLogs = false;
  logsContainer: string;

  ngOnInit (): void {
    this.refresh();
  }

  refresh (): void {
    this.loading = true;
    this.buildService.getBinaryBuildsByImage(this.image.name, this.image.cluster.id).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-BUILD.GET-BINARY-BUILD-HISTORY-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.BinaryBuilds) {
        this.allBuilds = data.data.BinaryBuilds;
        this.noHistory = false;
      } else {
        this.allBuilds = [];
        this.noHistory = true;
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-BUILD.GET-BINARY-BUILD-HISTORY-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.loading = false;
    });
  }

  startBuild (): void {
    this.buildService.createBinaryBuild(this.image.name, this.image.cluster.id).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.STARTERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-BUILD.START-BUILD-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.CreateBinaryBuild) {
        this.toast.show(`${data.data.CreateBinaryBuild.name} ` + this.translate.instant('MESSAGE.BUILDS.BINARY-BUILD.START-BUILD-MSG'), this.translate.instant('MESSAGE.BUILDS.BINARY-BUILD.START-BUILD-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-BUILD.START-BUILD-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.refresh();
    })
  }

  showBuildLogging (name: string): void {
    this.showBuildLogs = true;
    this.logsContainer = name;
  }

  showDeleteBuild (name: string): void {
    this.deleteBuildName = name;
    const deleteObjectEvent: DeleteObjectEvent = {
      title: this.translate.instant('BUILD.BINARYIMAGE-DETAILS.BUILD-DETAILS.DELETEBUILDJOB-TITLE'),
      message: this.translate.instant('BUILD.BINARYIMAGE-DETAILS.BUILD-DETAILS.DELETEBUILDJOB-DESC'),
      formInputLabel: '',
      alerts: [{
        message: this.translate.instant('BUILD.BINARYIMAGE-DETAILS.BUILD-DETAILS.DELETEBUILDJOB-ALERT'),
        type: ALERTTYPE.INFO
      }],
      objectName: name,
      objectType: DELETEOBJECTTYPE.BINARYBUILD,
      button: DELETEOBJECTBUTTON.YES_NO
    };
    this.deleteDialogView.open(deleteObjectEvent);
  }

  deleteBinaryBuild (event: DeleteConfirmationAction): void {
    if (event.objectType === DELETEOBJECTTYPE.BINARYBUILD && event.state === CONFIRMATIONSTATE.CONFIRMED) {
      this.buildService.deleteBinaryBuild(this.deleteBuildName, this.image.cluster.id).subscribe(data => {
        if (data.errors) {
          for (const each of data.errors) {
            if (each.message) {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-BUILD.DELETE-BUILD-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            }
          }
        }
        if (data.data?.DeleteBinaryBuild) {
          this.toast.show(`${data.data.DeleteBinaryBuild.name} ` + this.translate.instant('MESSAGE.BUILDS.BINARY-BUILD.DELETE-BUILD-MSG'), this.translate.instant('MESSAGE.BUILDS.BINARY-BUILD.DELETE-BUILD-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
        }
      }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-BUILD.DELETE-BUILD-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
        this.refresh();
      });
    }
  }
}
