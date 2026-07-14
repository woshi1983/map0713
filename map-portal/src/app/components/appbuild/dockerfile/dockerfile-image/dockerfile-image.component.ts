import { Component, ViewChild } from '@angular/core';
import { AuthService } from '../../../../services/auth.service';
import { ClustersService } from '../../../../services/clusters.service';
import { BuildService } from '../../../../services/build.service';
import { RepositorysService } from '../../../../services/repositorys.service';
import { ActivatedRoute } from '@angular/router';
import {
  Cluster,
  CreateDockerfileImage, DeleteConfirmationAction, DeleteObjectEvent,
  DockerfileImage,
  handleObject,
  UpdateDockerfileImage
} from '../../../../classes';
import { ClrDatagridPagination, ClrWizard } from '@clr/angular';
import { CustomValidatorDirective, nameValidator } from '../../../../utils/validators/custom-validator.directive';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from '@ngx-translate/core';
import { ApplicationsService } from '../../../../services/applications.service';
import { OrganizationService } from '../../../../services/organization.service';
import { DeleteObjectComponent } from '../../../shared-components/delete-object/delete-object.component';
import { customTrackBy } from '../../../../utils/operator/customTrackBy';
import { CONFIRMATIONSTATE, DELETEOBJECTBUTTON, DELETEOBJECTTYPE } from '../../../../utils/static-entries/shared.const';

@Component({
  selector: 'app-dockerfile-image',
  templateUrl: './dockerfile-image.component.html',
  styleUrls: ['./dockerfile-image.component.css']
})
export class DockerfileImageComponent {
  // admin and user role can handle cluster resources
  get canWriteResources (): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER_RESOURCE);
  }

  constructor (
    private clusterService: ClustersService,
    private applicationService: ApplicationsService,
    private buildService: BuildService,
    public authService: AuthService,
    private registryService: RepositorysService,
    private route: ActivatedRoute,
    private toast: ToastrService,
    private translate: TranslateService,
    private orgService: OrganizationService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  @ViewChild('deleteObjectComponent')
  deleteDialogView: DeleteObjectComponent;

  orgId: string;
  allImages: DockerfileImage[] = [];
  selectedImage: DockerfileImage;
  allClusters: Cluster[] = [];

  createDockerfileImageOpen = false;
  editDockerfileImageOpen = false;

  pushableRegistrytList = [];
  outputRegistryEndpoint: string;
  outputProjectList = [];
  noOutputProjectAlert = false;

  loading = true;
  showSearch = true;
  searchName: string;

  customTrackBy = customTrackBy
  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizarddockerfileimage') wizardnewdockerfileimage: ClrWizard;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  readonly packagePattern = CustomValidatorDirective.packagePattern;
  readonly gitUrlPattern = CustomValidatorDirective.gitUrlPattern;
  dockerfileImageSourceForm = new FormGroup({
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

  refresh (): void {
    this.loading = true;
    const filters = [{ key: 'organization', value: [this.orgId] }];
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.buildService.getDockerfileImages(filters).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.GET-DOCKERFILE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.allImages = data.data?.DockerfileImages ? data.data.DockerfileImages : [];
      this.selectedImage = null;
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.GET-DOCKERFILE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.loading = false;
    });
  }

  getAllRegistry (clusterId: string): void {
    this.pushableRegistrytList = null;
    let builderRegistrytList = [];
    for (const cluster of this.allClusters) {
      if (cluster.id === clusterId) {
        builderRegistrytList = cluster.imageRegistries.edge.nodes;
      }
    }
    // check pushable
    const registryList = [];
    for (const registry of builderRegistrytList) {
      if (registry.imageRegistry.authUser) {
        registryList.push(registry);
      }
    }
    this.pushableRegistrytList = registryList;
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
        this.outputRegistryEndpoint = '';
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

  showCreateImage (): void {
    this.clusterService.getClusterAndRegistry(this.orgId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.allClusters = data.data?.Clusters?.edge?.nodes ? data.data.Clusters.edge.nodes : [];
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }))
    this.createDockerfileImageOpen = true;
  }

  private doDockerfileImageReset (): void {
    this.wizardnewdockerfileimage.reset();
    this.dockerfileImageSourceForm.reset();
    this.dockerfileImageSourceForm.patchValue({
      sourceType: 'git',
      gitBranch: 'master',
      gitSubdir: '/',
      dockerfileName: 'Dockerfile'
    });
    this.allClusters = [];
    this.wizardeditdockerfileimage.reset();
    this.editDockerfileImageSourceForm.reset();
    this.createDockerfileImageOpen = false;
    this.editDockerfileImageOpen = false;
    this.noOutputProjectAlert = false;
    this.pushableRegistrytList = [];
    this.outputProjectList = [];
  }

  doDockerfileImageFinish (): void {
    const name = this.dockerfileImageSourceForm.value.name;
    const outputImageProject = this.dockerfileImageSourceForm.value.appImageProject;
    const outputRegistryURL = new URL(this.outputRegistryEndpoint);
    const outputImage = `${outputRegistryURL.host}/${outputImageProject}/${name}`;
    const annoations: JSON = JSON.parse(JSON.stringify({}));
    annoations['map.binary-image.output.registryId'] = this.dockerfileImageSourceForm.value.appImageRegistry;
    const param: CreateDockerfileImage = {
      name: name,
      annotations: annoations,
      spec: {
        source: this.dockerfileImageSourceForm.value.sourceType,
        distImage: outputImage,
        branch: this.dockerfileImageSourceForm.value.gitBranch ? this.dockerfileImageSourceForm.value.gitBranch : 'master',
        subDir: this.dockerfileImageSourceForm.value.gitSubdir ? this.dockerfileImageSourceForm.value.gitSubdir : '/',
        fileName: this.dockerfileImageSourceForm.value.dockerfileName ? this.dockerfileImageSourceForm.value.dockerfileName : 'Dockerfile',
        url: this.dockerfileImageSourceForm.value.dockerfileURL
      },
      cluster: this.dockerfileImageSourceForm.value.cluster
    }
    this.buildService.createDockerfileImage(param).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.CREATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.CREATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.CreateDockerfileImage) {
        this.toast.show(`${data.data.CreateDockerfileImage.name} ` + this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.CREATE-IMAGE-MSG'), this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.CREATE-IMAGE-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.CREATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.doDockerfileImageReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doDockerfileImageCancel (): void {
    this.doDockerfileImageReset();
  }

  showUpdateDockerfileImage (selectedImage: DockerfileImage): void {
    this.selectedImage = selectedImage;
    const outputImageRegistry = selectedImage.annotations['map.binary-image.output.registryId'];
    const outputImageList = selectedImage.spec.distImage.split('/');
    const outputImageProject = outputImageList[1];
    this.editDockerfileImageSourceForm.setValue({
      name: selectedImage.name,
      cluster: selectedImage.cluster.id,
      sourceType: selectedImage.spec.source,
      appImageRegistry: outputImageRegistry,
      appImageProject: outputImageProject,
      dockerfileURL: selectedImage.spec.url,
      gitBranch: selectedImage.spec.branch,
      gitSubdir: selectedImage.spec.subDir,
      dockerfileName: selectedImage.spec.fileName
    });
    this.getAllRegistryForUpdate(selectedImage.cluster.id, outputImageRegistry);
    this.editDockerfileImageOpen = true;
  }

  doUpdateDockerfileImageFinish (): void {
    const name = this.editDockerfileImageSourceForm.value.name;
    const outputImageProject = this.editDockerfileImageSourceForm.value.appImageProject;
    const outputRegistryURL = new URL(this.outputRegistryEndpoint);
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
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doUpdateDockerfileImageCancel (): void {
    this.doDockerfileImageReset();
  }

  showDeleteImage (selectedImage: DockerfileImage): void {
    this.selectedImage = selectedImage;
    const deleteObjectEvent: DeleteObjectEvent = {
      title: this.translate.instant('BUILD.DOCKERFILEIMAGE.ACTION.DELETEIMAGE'),
      message: this.translate.instant('BUILD.DOCKERFILEIMAGE.ACTION.DELETEIMAGE-DESC'),
      formInputLabel: this.translate.instant('SHARED.DELETE.DELETEDESC'),
      alerts: [],
      objectName: selectedImage.name,
      objectType: DELETEOBJECTTYPE.DOCKERFILEIMAGE,
      button: DELETEOBJECTBUTTON.CONFIRM_CANCEL
    };
    this.deleteDialogView.open(deleteObjectEvent);
  }

  deleteDockerfileImage (event: DeleteConfirmationAction): void {
    if (event.objectType === DELETEOBJECTTYPE.DOCKERFILEIMAGE && event.state === CONFIRMATIONSTATE.CONFIRMED) {
      this.buildService.deleteDockerfileImage(this.selectedImage?.name, this.selectedImage?.cluster?.id).subscribe(data => {
        if (data.errors) {
          for (const each of data.errors) {
            if (each.message) {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.DELETE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            }
          }
        }
        if (data.data?.DeleteDockerfileImage) {
          this.toast.show(`${data.data.DeleteDockerfileImage.name} ` + this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.DELETE-IMAGE-MSG'), this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.DELETE-IMAGE-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
        }
      }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.DOCKERFILE-IMAGE.DELETE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
        if (this.pagination.currentPage === 1) {
          this.refresh();
        } else {
          this.pagination.currentPage = 1;
        }
      });
    }
  }
}
