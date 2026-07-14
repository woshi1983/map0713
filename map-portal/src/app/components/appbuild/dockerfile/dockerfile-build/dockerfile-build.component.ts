import { Component, OnInit, Input, ViewChild } from '@angular/core';
import {
  DeleteConfirmationAction,
  DeleteObjectEvent,
  DockerfileBuild,
  DockerfileImage,
  handleObject
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
  selector: 'app-dockerfile-build',
  templateUrl: './dockerfile-build.component.html',
  styleUrls: ['./dockerfile-build.component.css']
})
export class DockerfileBuildComponent implements OnInit {
  // admin and user role can handle cluster resources
  get canWriteResources (): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER_RESOURCE);
  }

  @Input() image: DockerfileImage;
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
  allBuilds: DockerfileBuild[] = [];
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
    this.buildService.getDockerfileBuildsByImage(this.image.name, this.image.cluster.id).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-BUILD.GET-BINARY-BUILD-HISTORY-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.DockerfileBuilds) {
        this.allBuilds = data.data.DockerfileBuilds;
        this.noHistory = false;
      } else {
        this.allBuilds = [];
        this.noHistory = true;
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-BUILD.GET-BINARY-BUILD-HISTORY-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.loading = false;
    });
  }

  startBuild (): void {
    this.buildService.createDockerfileBuild(this.image.name, this.image.cluster.id).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.STARTERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-BUILD.START-BUILD-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.CreateDockerfileBuild) {
        this.toast.show(`${data.data.CreateDockerfileBuild.name} ` + this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-BUILD.START-BUILD-MSG'), this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-BUILD.START-BUILD-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-BUILD.START-BUILD-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
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
      title: this.translate.instant('BUILD.DOCKERFILEIMAGE-DETAILS.DELETEBUILDJOB-TITLE'),
      message: this.translate.instant('BUILD.DOCKERFILEIMAGE-DETAILS.DELETEBUILDJOB-DESC'),
      formInputLabel: '',
      alerts: [{
        message: this.translate.instant('BUILD.BINARYIMAGE-DETAILS.BUILD-DETAILS.DELETEBUILDJOB-ALERT'),
        type: ALERTTYPE.INFO
      }],
      objectName: name,
      objectType: DELETEOBJECTTYPE.DOCKERFILEBUILD,
      button: DELETEOBJECTBUTTON.YES_NO
    };
    this.deleteDialogView.open(deleteObjectEvent);
  }

  deleteDockerfileBuild (event: DeleteConfirmationAction): void {
    if (event.objectType === DELETEOBJECTTYPE.BINARYBUILD && event.state === CONFIRMATIONSTATE.CONFIRMED) {
      this.buildService.deleteDockerfileBuild(this.deleteBuildName, this.image?.cluster?.id).subscribe(data => {
        if (data.errors) {
          for (const each of data.errors) {
            if (each.message) {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-BUILD.DELETE-BUILD-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            }
          }
        }
        if (data.data?.DeleteDockerfileBuild) {
          this.toast.show(`${data.data.DeleteDockerfileBuild.name} ` + this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-BUILD.DELETE-BUILD-MSG'), this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-BUILD.DELETE-BUILD-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
        }
      }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-BUILD.DELETE-BUILD-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
        this.refresh();
      });
    }
  }
}
