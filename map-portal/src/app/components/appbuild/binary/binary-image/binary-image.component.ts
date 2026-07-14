import { Component, ViewChild } from '@angular/core';
import { BuildService } from '../../../../services/build.service';
import { AuthService } from '../../../../services/auth.service';
import { ClrDatagridPagination, ClrWizard } from '@clr/angular';
import {
  BinaryImage,
  Cluster,
  CreateBinaryImage,
  DeleteConfirmationAction,
  DeleteObjectEvent,
  handleObject,
  UpdateBinaryImage
} from '../../../../classes';
import { ClustersService } from '../../../../services/clusters.service';
import { CustomValidatorDirective, nameValidator } from '../../../../utils/validators/custom-validator.directive';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { RepositorysService } from '../../../../services/repositorys.service';
import { ActivatedRoute } from '@angular/router';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from '@ngx-translate/core';
import { OrganizationService } from '../../../../services/organization.service';
import { customTrackBy } from '../../../../utils/operator/customTrackBy';
import { DELETEOBJECTBUTTON, DELETEOBJECTTYPE, CONFIRMATIONSTATE } from '../../../../utils/static-entries/shared.const';
import { DeleteObjectComponent } from '../../../shared-components/delete-object/delete-object.component';

@Component({
  selector: 'app-binary-image',
  templateUrl: './binary-image.component.html',
  styleUrls: ['./binary-image.component.css']
})
export class BinaryImageComponent {
  // admin and user role can handle cluster resources
  get canWriteResources (): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER_RESOURCE);
  }

  constructor (
    private clusterService: ClustersService,
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
  allImages: BinaryImage[] = [];
  selectedImage: BinaryImage;

  allClusters: Cluster[] = [];

  createBinaryImageOpen = false;
  editBinaryImageOpen = false;

  builderRegistrytList = [];
  builderRegistryEndpoint: string;
  builderProjectList = [];
  builderImageList = [];
  builderTagsList = [];
  pushableRegistrytList = [];
  outputRegistryEndpoint: string;
  outputProjectList = [];
  noImageProjectAlert = false;
  noOutputProjectAlert = false;
  noImageNameAlert = false;
  noImageTagAlert = false;

  loading = true;
  showSearch = true;
  searchName: string;

  customTrackBy = customTrackBy

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardbinaryimage') wizardnewbinaryimage: ClrWizard;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  readonly packagePattern = CustomValidatorDirective.packagePattern;
  binaryImageSourceForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(64), Validators.pattern(this.restricNamePattern), nameValidator(['vmware-system-map-init'])]),
    cluster: new FormControl('', Validators.required),
    sourceURL: new FormControl('', [Validators.required, Validators.pattern(this.packagePattern)]),
    builderRegistry: new FormControl('', [Validators.required]),
    builderProject: new FormControl('', [Validators.required]),
    builderImage: new FormControl('', [Validators.required]),
    builderImageTag: new FormControl('', [Validators.required]),
    appImageRegistry: new FormControl('', [Validators.required]),
    appImageProject: new FormControl('', [Validators.required]),
    deployDir: new FormControl('/app', Validators.required),
    processId: new FormControl(0)
  });

  @ViewChild('wizardeditbinaryimage') wizardeditbinaryimage: ClrWizard;
  editBinaryImageSourceForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(64), Validators.pattern(this.restricNamePattern), nameValidator(['vmware-system-map-init'])]),
    cluster: new FormControl('', Validators.required),
    sourceURL: new FormControl('', [Validators.required, Validators.pattern(this.packagePattern)]),
    builderRegistry: new FormControl('', [Validators.required]),
    builderProject: new FormControl('', [Validators.required]),
    builderImage: new FormControl('', [Validators.required]),
    builderImageTag: new FormControl('', [Validators.required]),
    appImageRegistry: new FormControl('', [Validators.required]),
    appImageProject: new FormControl('', [Validators.required]),
    deployDir: new FormControl('/app', Validators.required),
    processId: new FormControl(0)
  });

  refresh (): void {
    this.loading = true;
    const filters = [{ key: 'organization', value: [this.orgId] }];
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.buildService.getBinaryImages(filters).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-BINARY-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.allImages = data.data?.BinaryImages ? data.data.BinaryImages : [];
      this.selectedImage = null;
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-BINARY-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.loading = false;
    });
  }

  getAllRegistry (clusterId: string): void {
    this.builderRegistrytList = null;
    this.pushableRegistrytList = null;
    for (const cluster of this.allClusters) {
      if (cluster.id === clusterId) {
        this.builderRegistrytList = cluster.imageRegistries.edge.nodes;
      }
    }
    // check pushable
    const registryList = [];
    for (const registry of this.builderRegistrytList) {
      if (registry.imageRegistry.authUser) {
        registryList.push(registry);
      }
    }
    this.pushableRegistrytList = registryList;
  }

  getAllBuilderProjects (registryId: string): void {
    this.builderProjectList = null;
    this.builderRegistryEndpoint = null;
    this.registryService.getRegistryProject(registryId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-REGISTRY-PROJECT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.ImageRegistry) {
        this.builderRegistryEndpoint = data.data.ImageRegistry.endpoint;
        this.builderProjectList = data.data.ImageRegistry.projects;
      } else {
        this.builderProjectList = [];
        this.builderRegistryEndpoint = '';
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-REGISTRY-PROJECT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
  }

  getAllImages (projectName: string): void {
    this.builderImageList = null;
    this.registryService.getAllImagesByProject(this.binaryImageSourceForm.value.builderRegistry, projectName).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-PROJECT-IMAGE-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.ImageRepos) {
        this.builderImageList = data.data.ImageRepos;
      } else {
        this.builderImageList = [];
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-PROJECT-IMAGE-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
  }

  getAllTags (imageName: string, updated?: boolean): void {
    this.builderTagsList = null;
    this.noImageNameAlert = false;
    for (const repo of this.builderImageList) {
      if (repo.name === imageName) {
        this.builderTagsList = repo.tags;
        if (updated) {
          if (!this.builderTagsList.find(x => x === this.editBinaryImageSourceForm.value.builderImageTag)) {
            this.noImageTagAlert = true;
          }
        }
      }
    }
  }

  getAllOutputProjects (registryId: string): void {
    this.outputProjectList = null;
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
      } else {
        this.outputProjectList = [];
        this.outputRegistryEndpoint = '';
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-REGISTRY-PROJECT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
  }

  getAllRegistryForUpdate (clusterId: string, sourceRegistryId: string, outputRegistryId: string): void {
    this.builderRegistrytList = null;
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
        this.builderRegistrytList = data.data.Cluster.imageRegistries.edge.nodes;
        // check pushable
        const registryList = [];
        for (const registry of this.builderRegistrytList) {
          if (registry.imageRegistry.authUser) {
            registryList.push(registry);
          }
        }
        this.pushableRegistrytList = registryList;
        this.getAllBuilderProjectsForUpdate(sourceRegistryId);
        this.getAllOutputProjectsForUpdate(outputRegistryId);
      } else {
        this.builderRegistrytList = [];
        this.pushableRegistrytList = [];
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }))
  }

  getAllBuilderProjectsForUpdate (sourceRegistryId: string): void {
    this.builderProjectList = null;
    this.builderRegistryEndpoint = null;
    this.noImageProjectAlert = false;
    this.registryService.getRegistryProject(sourceRegistryId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-REGISTRY-PROJECT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.ImageRegistry) {
        this.builderRegistryEndpoint = data.data.ImageRegistry.endpoint;
        this.builderProjectList = data.data.ImageRegistry.projects;
        if (!this.builderProjectList.find(x => x === this.editBinaryImageSourceForm.value.builderProject)) {
          this.noImageProjectAlert = true;
        } else {
          this.getAllImagesForUpdate(sourceRegistryId, this.editBinaryImageSourceForm.value.builderProject);
        }
      } else {
        this.noImageProjectAlert = true;
        this.builderProjectList = [];
        this.builderRegistryEndpoint = '';
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-REGISTRY-PROJECT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
  }

  getAllImagesForUpdate (sourceRegistryId: string, projectName: string): void {
    this.builderImageList = null;
    this.noImageProjectAlert = false;
    this.registryService.getAllImagesByProject(sourceRegistryId, projectName).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-PROJECT-IMAGE-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.ImageRepos) {
        this.builderImageList = data.data.ImageRepos;
        if (!this.builderImageList.find(x => x.name === this.editBinaryImageSourceForm.value.builderImage)) {
          this.noImageNameAlert = true;
        } else {
          this.getAllTags(this.editBinaryImageSourceForm.value.builderImage, true);
        }
      } else {
        this.noImageNameAlert = true;
        this.builderImageList = [];
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-PROJECT-IMAGE-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
  }

  getAllOutputProjectsForUpdate (registryId: string): void {
    this.outputProjectList = null;
    this.outputRegistryEndpoint = null;
    this.noOutputProjectAlert = false;
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
        this.noOutputProjectAlert = !this.outputProjectList.find(x => x === this.editBinaryImageSourceForm.value.appImageProject);
      } else {
        this.noOutputProjectAlert = true;
        this.outputProjectList = [];
        this.outputRegistryEndpoint = '';
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-REGISTRY-PROJECT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
  }

  changeNoTagsAlert (): void {
    this.noImageTagAlert = false;
  }

  changeNoOutputProjectAlert (): void {
    this.noOutputProjectAlert = false;
  }

  showCreateBinaryImage (): void {
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
    this.createBinaryImageOpen = true;
  }

  private doCreateBinaryImageReset (): void {
    this.wizardnewbinaryimage.reset();
    this.binaryImageSourceForm.reset();
    this.binaryImageSourceForm.patchValue({
      processId: '0',
      deployDir: '/app'
    });
    this.createBinaryImageOpen = false;
    this.builderRegistrytList = [];
    this.builderRegistryEndpoint = '';
    this.builderProjectList = [];
    this.builderImageList = [];
    this.builderTagsList = [];
    this.pushableRegistrytList = [];
    this.outputRegistryEndpoint = '';
    this.outputProjectList = [];
    this.allClusters = [];
  }

  doCreateBinaryImageFinish (): void {
    const name = this.binaryImageSourceForm.value.name;
    const builderRegistryURL = new URL(this.builderRegistryEndpoint);
    const bBuildProjectName = this.binaryImageSourceForm.value.builderProject;
    const bBuildImageName = this.binaryImageSourceForm.value.builderImage;
    const bBuildImageTag = this.binaryImageSourceForm.value.builderImageTag;
    const bBuildImage = `${builderRegistryURL.host}/${bBuildProjectName}/${bBuildImageName}:${bBuildImageTag}`;
    const outputRegistryURL = new URL(this.outputRegistryEndpoint);
    const outputImageProject = this.binaryImageSourceForm.value.appImageProject;
    const outputImage = `${outputRegistryURL.host}/${outputImageProject}/${name}`;
    const annoations: JSON = JSON.parse(JSON.stringify({}));
    annoations['map.binary-image.source.registryId'] = this.binaryImageSourceForm.value.builderRegistry;
    annoations['map.binary-image.output.registryId'] = this.binaryImageSourceForm.value.appImageRegistry;
    const param: CreateBinaryImage = {
      name: name,
      annotations: annoations,
      spec: {
        baseImage: bBuildImage,
        distImage: outputImage,
        binaryDir: this.binaryImageSourceForm.value.deployDir ? this.binaryImageSourceForm.value.deployDir : '/app',
        url: this.binaryImageSourceForm.value.sourceURL,
        processUID: this.binaryImageSourceForm.value.processId ? this.binaryImageSourceForm.value.processId : 0
      },
      cluster: this.binaryImageSourceForm.value.cluster
    }
    this.buildService.createBinaryImage(param).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.CREATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.CREATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.CreateBinaryImage) {
        this.toast.show(`${data.data.CreateBinaryImage.name} ` + this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.CREATE-IMAGE-MSG'), this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.CREATE-IMAGE-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.CREATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.doCreateBinaryImageReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doCreateBinaryImageCancel (): void {
    this.doCreateBinaryImageReset();
  }

  showUpdateBinaryImage (selectedImage: BinaryImage): void {
    this.selectedImage = selectedImage;
    const selectedBuilderRegistry = selectedImage.annotations['map.binary-image.source.registryId'];
    const outputImageRegistry = selectedImage.annotations['map.binary-image.output.registryId'];
    const baseImageList = selectedImage.spec.baseImage.split('/');
    const selectedBuilderProject = baseImageList[1];
    const builderImage = baseImageList[2];
    const selectedBuilderImage = builderImage.split(':');
    const builderImageName = selectedBuilderImage[0];
    const builderTagName = selectedBuilderImage[1];
    const outputImageList = selectedImage.spec.distImage.split('/');
    const outputImageProject = outputImageList[1];
    this.editBinaryImageSourceForm.setValue({
      name: selectedImage.name,
      cluster: selectedImage.cluster.id,
      sourceURL: selectedImage.spec.url,
      builderRegistry: selectedBuilderRegistry,
      builderProject: selectedBuilderProject,
      builderImage: builderImageName,
      builderImageTag: builderTagName,
      appImageRegistry: outputImageRegistry,
      appImageProject: outputImageProject,
      deployDir: selectedImage.spec.binaryDir,
      processId: selectedImage.spec.processUID
    });
    this.getAllRegistryForUpdate(selectedImage.cluster.id, selectedBuilderRegistry, outputImageRegistry);
    this.editBinaryImageOpen = true;
  }

  doUpdateBinaryImageFinish (): void {
    const name = this.editBinaryImageSourceForm.value.name;
    const builderRegistryURL = new URL(this.builderRegistryEndpoint);
    const bBuildProjectName = this.editBinaryImageSourceForm.value.builderProject;
    const bBuildImageName = this.editBinaryImageSourceForm.value.builderImage;
    const bBuildImageTag = this.editBinaryImageSourceForm.value.builderImageTag;
    const bBuildImage = `${builderRegistryURL.host}/${bBuildProjectName}/${bBuildImageName}:${bBuildImageTag}`;
    const outputRegistryURL = new URL(this.outputRegistryEndpoint);
    const outputImageProject = this.editBinaryImageSourceForm.value.appImageProject;
    const outputImage = `${outputRegistryURL.host}/${outputImageProject}/${name}`;
    const annoations: JSON = JSON.parse(JSON.stringify({}));
    annoations['map.binary-image.source.registryId'] = this.editBinaryImageSourceForm.value.builderRegistry;
    annoations['map.binary-image.output.registryId'] = this.editBinaryImageSourceForm.value.appImageRegistry;
    const param: UpdateBinaryImage = {
      name: name,
      annotations: annoations,
      spec: {
        baseImage: bBuildImage,
        distImage: outputImage,
        binaryDir: this.editBinaryImageSourceForm.value.deployDir ? this.editBinaryImageSourceForm.value.deployDir : '/app',
        url: this.editBinaryImageSourceForm.value.sourceURL,
        processUID: this.editBinaryImageSourceForm.value.processId ? this.editBinaryImageSourceForm.value.processId : 0
      },
      cluster: this.editBinaryImageSourceForm.value.cluster
    }
    this.buildService.updateBinaryImage(param).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.UPDATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.UPDATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.UpdateBinaryImage) {
        this.toast.show(`${data.data.UpdateBinaryImage.name} ` + this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.UPDATE-IMAGE-MSG'), this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.UPDATE-IMAGE-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.UPDATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.doUpdateBinaryImageReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doUpdateBinaryImageCancel (): void {
    this.doUpdateBinaryImageReset();
  }

  doUpdateBinaryImageReset (): void {
    this.wizardeditbinaryimage.reset();
    this.editBinaryImageSourceForm.reset();
    this.editBinaryImageOpen = false;
    this.builderRegistrytList = [];
    this.builderProjectList = [];
    this.builderImageList = [];
    this.builderTagsList = [];
    this.pushableRegistrytList = [];
    this.outputProjectList = [];
    this.noOutputProjectAlert = false;
    this.noImageProjectAlert = false;
    this.noImageNameAlert = false;
    this.noImageTagAlert = false;
  }

  showDeleteBinaryImage (selectedImage: BinaryImage): void {
    this.selectedImage = selectedImage;
    const deleteObjectEvent: DeleteObjectEvent = {
      title: this.translate.instant('BUILD.BINARYIMAGE.ACTION.DELETEIMAGE'),
      message: this.translate.instant('BUILD.BINARYIMAGE.ACTION.DELETEIMAGE-DESC'),
      formInputLabel: this.translate.instant('SHARED.DELETE.DELETEDESC'),
      alerts: [],
      objectName: selectedImage.name,
      objectType: DELETEOBJECTTYPE.BINARYIMAGE,
      button: DELETEOBJECTBUTTON.CONFIRM_CANCEL
    };
    this.deleteDialogView.open(deleteObjectEvent);
  }

  deleteBinaryImage (event: DeleteConfirmationAction): void {
    if (event.objectType === DELETEOBJECTTYPE.BINARYIMAGE && event.state === CONFIRMATIONSTATE.CONFIRMED) {
      this.buildService.deleteBinaryImage(this.selectedImage?.name, this.selectedImage?.cluster?.id).subscribe(data => {
        if (data.errors) {
          for (const each of data.errors) {
            if (each.message) {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.DELETE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            }
          }
        }
        if (data.data?.DeleteBinaryImage) {
          this.toast.show(`${data.data.DeleteBinaryImage.name} ` + this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.DELETE-IMAGE-MSG'), this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.DELETE-IMAGE-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
        }
      }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.DELETE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
        if (this.pagination.currentPage === 1) {
          this.refresh();
        } else {
          this.pagination.currentPage = 1;
        }
      });
    }
  }
}
