import { Component, OnInit, ViewChild } from '@angular/core';
import {
  DeleteConfirmationAction,
  DeleteObjectEvent,
  DockerfileImage,
  handleObject,
  UpdateDockerfileImage
} from '../../../../classes';
import { ActivatedRoute, Router } from '@angular/router';
import { AuthService } from '../../../../services/auth.service';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from '@ngx-translate/core';
import { BuildService } from '../../../../services/build.service';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { ClrWizard } from '@clr/angular';
import { ClustersService } from '../../../../services/clusters.service';
import { CustomValidatorDirective, nameValidator } from '../../../../utils/validators/custom-validator.directive';
import { OrganizationService } from '../../../../services/organization.service';
import { RepositorysService } from '../../../../services/repositorys.service';
import { customTrackBy } from '../../../../utils/operator/customTrackBy';
import { DeleteObjectComponent } from '../../../shared-components/delete-object/delete-object.component';
import { CONFIRMATIONSTATE, DELETEOBJECTBUTTON, DELETEOBJECTTYPE } from '../../../../utils/static-entries/shared.const';

@Component({
  selector: 'app-dockerfile-image-details',
  templateUrl: './dockerfile-image-details.component.html',
  styleUrls: ['./dockerfile-image-details.component.css']
})
export class DockerfileImageDetailsComponent implements OnInit {
  // admin and user role can handle cluster resources
  get canWriteResources (): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER_RESOURCE);
  }

  constructor (
    private route: ActivatedRoute,
    private router: Router,
    public authService: AuthService,
    private buildService: BuildService,
    private toast: ToastrService,
    private translate: TranslateService,
    private orgService: OrganizationService,
    private clusterService: ClustersService,
    private registryService: RepositorysService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  @ViewChild('deleteObjectComponent')
  deleteDialogView: DeleteObjectComponent;

  orgId: string;
  clusterId: string;
  imageName: string;
  image: DockerfileImage;

  editDockerfileImageOpen = false;

  pushableRegistrytList = [];
  outputRegistryEndpoint: string;
  outputProjectList = [];
  noOutputProjectAlert = false;
  loading = false;

  customTrackBy = customTrackBy

  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  readonly packagePattern = CustomValidatorDirective.packagePattern;
  readonly gitUrlPattern = CustomValidatorDirective.gitUrlPattern;
  @ViewChild('wizardeditdockerfileimage') wizardeditdockerfileimage: ClrWizard;
  editDockerfileImageSourceForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(64), Validators.pattern(this.restricNamePattern), nameValidator(['vmware-system-map-init'])]),
    cluster: new FormControl('git', [Validators.required]),
    sourceType: new FormControl('git', [Validators.required]),
    appImageRegistry: new FormControl('', [Validators.required]),
    appImageProject: new FormControl('', [Validators.required]),
    dockerfileURL: new FormControl('', [Validators.required, Validators.pattern(this.packagePattern)]),
    gitBranch: new FormControl('master'),
    gitSubdir: new FormControl('/'),
    dockerfileName: new FormControl('Dockerfile')
  });

  ngOnInit (): void {
    this.refresh();
  }

  refresh (): void {
    this.loading = true;
    this.clusterId = this.route.snapshot.paramMap.get('id');
    this.imageName = this.route.snapshot.paramMap.get('name');
    this.image = null;
    this.buildService.getDockerfileImageDetails(this.clusterId, this.imageName).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.GET-DOCKERFILE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.DockerfileImage) {
        this.image = data.data.DockerfileImage;
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.GET-DOCKERFILE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.loading = false;
    });
  }

  backToImages (): void {
    this.router.navigate(['/portal/appbuild/dockerfile-image']);
  }

  getAllOutputProjects (registryId: string, updated?: boolean): void {
    this.outputProjectList = [];
    this.noOutputProjectAlert = false;
    this.outputRegistryEndpoint = null;
    this.registryService.getRegistryProject(registryId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-REGISTRY-PROJECT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.ImageRegistry) {
        this.outputRegistryEndpoint = data.data.ImageRegistry.endpoint;
        this.outputProjectList = data.data.ImageRegistry.projects;
        if (updated) {
          this.noOutputProjectAlert = !this.outputProjectList.find(x => x === this.editDockerfileImageSourceForm.value.appImageProject);
        }
      } else {
        this.noOutputProjectAlert = true;
        this.outputProjectList = [];
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-REGISTRY-PROJECT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
  }

  getAllRegistryForUpdate (clusterId: string, outputRegistryId: string): void {
    this.pushableRegistrytList = null;
    this.clusterService.getRegistryOfCluster(clusterId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.Cluster?.imageRegistries?.edge?.nodes) {
        // check pushable
        const registryList = [];
        for (const registry of data.data.Cluster.imageRegistries.edge.nodes) {
          if (registry.imageRegistry.authUser) {
            registryList.push(registry);
          }
        }
        this.pushableRegistrytList = registryList;
        this.getAllOutputProjects(outputRegistryId, true);
      } else {
        this.pushableRegistrytList = [];
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }))
  }

  changeNoOutputProjectAlert (): void {
    this.noOutputProjectAlert = false;
  }

  showUpdateDockerfileImage (): void {
    const outputImageRegistry = this.image.annotations['map.binary-image.output.registryId'];
    const outputImageList = this.image.spec.distImage.split('/');
    const outputImageProject = outputImageList[1];
    this.editDockerfileImageSourceForm.setValue({
      name: this.imageName,
      cluster: this.clusterId,
      sourceType: this.image.spec.source,
      appImageRegistry: outputImageRegistry,
      appImageProject: outputImageProject,
      dockerfileURL: this.image.spec.url,
      gitBranch: this.image.spec.branch,
      gitSubdir: this.image.spec.subDir,
      dockerfileName: this.image.spec.fileName
    });
    this.getAllRegistryForUpdate(this.image.cluster.id, outputImageRegistry);
    this.editDockerfileImageOpen = true;
  }

  doUpdateDockerfileImageFinish (): void {
    const name = this.editDockerfileImageSourceForm.value.name;
    const outputRegistryURL = new URL(this.outputRegistryEndpoint);
    const outputImageProject = this.editDockerfileImageSourceForm.value.appImageProject;
    const outputImage = `${outputRegistryURL.host}/${outputImageProject}/${name}`;
    const annoations: JSON = JSON.parse(JSON.stringify({}));
    annoations['map.binary-image.output.registryId'] = this.editDockerfileImageSourceForm.value.appImageRegistry;
    const param: UpdateDockerfileImage = {
      name: name,
      annotations: annoations,
      spec: {
        source: this.editDockerfileImageSourceForm.value.sourceType,
        distImage: outputImage,
        branch: this.editDockerfileImageSourceForm.value.gitBranch ? this.editDockerfileImageSourceForm.value.gitBranch : 'master',
        subDir: this.editDockerfileImageSourceForm.value.gitSubdir ? this.editDockerfileImageSourceForm.value.gitSubdir : '/',
        fileName: this.editDockerfileImageSourceForm.value.dockerfileName ? this.editDockerfileImageSourceForm.value.dockerfileName : 'Dockerfile',
        url: this.editDockerfileImageSourceForm.value.dockerfileURL
      },
      cluster: this.editDockerfileImageSourceForm.value.cluster
    }
    this.buildService.updateDockerfileImage(param).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.UPDATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.UPDATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.UpdateDockerfileImage) {
        this.toast.show(`${data.data.UpdateDockerfileImage.name} ` + this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.UPDATE-IMAGE-MSG'), this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.UPDATE-IMAGE-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.UPDATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.doDockerfileImageReset();
      this.refresh();
    });
  }

  doUpdateDockerfileImageCancel (): void {
    this.doDockerfileImageReset();
  }

  private doDockerfileImageReset (): void {
    this.wizardeditdockerfileimage.reset();
    this.editDockerfileImageSourceForm.reset();
    this.editDockerfileImageOpen = false;
    this.noOutputProjectAlert = false;
  }

  showDeleteImage (): void {
    const deleteObjectEvent: DeleteObjectEvent = {
      title: this.translate.instant('BUILD.DOCKERFILEIMAGE.ACTION.EDITIMAGE'),
      message: this.translate.instant('BUILD.DOCKERFILEIMAGE.ACTION.DELETEIMAGE'),
      formInputLabel: this.translate.instant('SHARED.DELETE.DELETEDESC'),
      alerts: [],
      objectName: this.imageName,
      objectType: DELETEOBJECTTYPE.DOCKERFILEIMAGEDETAILS,
      button: DELETEOBJECTBUTTON.CONFIRM_CANCEL
    };
    this.deleteDialogView.open(deleteObjectEvent);
  }

  deleteDockerfileImage (event: DeleteConfirmationAction): void {
    if (event.objectType === DELETEOBJECTTYPE.DOCKERFILEIMAGEDETAILS && event.state === CONFIRMATIONSTATE.CONFIRMED) {
      this.buildService.deleteDockerfileImage(this.imageName, this.clusterId).subscribe(data => {
        if (data.errors) {
          for (const each of data.errors) {
            if (each.message) {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.DELETE-IMAGE-ERROR'), {
                toastClass: 'alert-danger',
                closeButton: true,
                extendedTimeOut: 0,
                timeOut: 0
              })
            }
          }
        }
        if (data.data?.DeleteDockerfileImage) {
          this.toast.show(`${data.data.DeleteDockerfileImage.name} ` + this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.DELETE-IMAGE-MSG'), this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.DELETE-IMAGE-SUCCESS'), {
            toastClass: 'alert-success',
            closeButton: true,
            extendedTimeOut: 10000,
            timeOut: 10000
          })
          this.router.navigate(['/portal/appbuild/dockerfile-image']);
        }
      }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.DELETE-IMAGE-ERROR'), {
        toastClass: 'alert-danger',
        closeButton: true,
        extendedTimeOut: 0,
        timeOut: 0
      }));
    }
  }
}
