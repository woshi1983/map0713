import { Component, OnInit, Input, ViewChild } from '@angular/core';
import {
  DeleteConfirmationAction, DeleteObjectEvent,
  handleObject,
  SourceBuild,
  SourceImage
} from '../../../../classes';
import { BuildService } from '../../../../services/build.service';
import { AuthService } from '../../../../services/auth.service';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from '@ngx-translate/core';
import { OrganizationService } from '../../../../services/organization.service';
import { DeleteObjectComponent } from '../../../shared-components/delete-object/delete-object.component';
import {
  ALERTTYPE,
  CONFIRMATIONSTATE,
  DELETEOBJECTBUTTON,
  DELETEOBJECTTYPE
} from '../../../../utils/static-entries/shared.const';

@Component({
  selector: 'app-source-build',
  templateUrl: './source-build.component.html',
  styleUrls: ['./source-build.component.css']
})
export class SourceBuildComponent implements OnInit {
  // admin and user role can handle cluster resources
  get canWriteResources (): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER_RESOURCE);
  }

  @Input() image: SourceImage;
  constructor (
    private buildService: BuildService,
    public authService: AuthService,
    private toast: ToastrService,
    private orgService: OrganizationService,
    private translate: TranslateService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  @ViewChild('deleteObjectComponent')
  deleteDialogView: DeleteObjectComponent;

  orgId: string;
  allBuilds: SourceBuild[] = [];
  noHistory: boolean;

  showSearchInput = false;
  loading = false;

  deleteBuildName: string;
  showBuildLogs = false;
  logsContainer: string;

  ngOnInit (): void {
    this.refresh();
  }

  refresh (): void {
    this.loading = true;
    this.allBuilds = [];
    this.buildService.getSourceBuildsByImage(this.image.name, this.image.cluster.id).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-BUILD.GET-SOURCE-BUILD-HISTORY-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.SourceBuilds) {
        this.allBuilds = data.data.SourceBuilds;
        this.noHistory = false;
      } else {
        this.allBuilds = [];
        this.noHistory = true;
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-BUILD.GET-SOURCE-BUILD-HISTORY-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.loading = false;
    });
  }

  startBuild (): void {
    this.buildService.createSourceBuild(this.image.name, this.image.cluster.id).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.STARTERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-BUILD.START-BUILD-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.CreateSourceBuild) {
        this.toast.show(`${data.data.CreateSourceBuild.name} ` + this.translate.instant('MESSAGE.BUILDS.SOURCE-BUILD.START-BUILD-MSG'), this.translate.instant('MESSAGE.BUILDS.SOURCE-BUILD.START-BUILD-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-BUILD.START-BUILD-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
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
      title: '删除源代码构建任务',
      message: '确认删除源代码构建任务',
      formInputLabel: '',
      alerts: [{
        message: '删除构建任务不会删除镜像仓库中构建输出的镜像。',
        type: ALERTTYPE.INFO
      }],
      objectName: name,
      objectType: DELETEOBJECTTYPE.SOURCEBUILD,
      button: DELETEOBJECTBUTTON.YES_NO
    };
    this.deleteDialogView.open(deleteObjectEvent);
  }

  deleteSourceBuild (event: DeleteConfirmationAction): void {
    if (event.objectType === DELETEOBJECTTYPE.SOURCEBUILD && event.state === CONFIRMATIONSTATE.CONFIRMED) {
      this.buildService.deleteSourceBuild(this.deleteBuildName, this.image?.cluster?.id).subscribe(data => {
        if (data.errors) {
          for (const each of data.errors) {
            if (each.message) {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-BUILD.DELETE-BUILD-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            }
          }
        }
        if (data.data?.DeleteSourceBuild) {
          this.toast.show(`${data.data.DeleteSourceBuild.name} ` + this.translate.instant('MESSAGE.BUILDS.SOURCE-BUILD.DELETE-BUILD-MSG'), this.translate.instant('MESSAGE.BUILDS.SOURCE-BUILD.DELETE-BUILD-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
        }
      }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-BUILD.DELETE-BUILD-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
        this.refresh();
      });
    }
  }
}
