import { Component, OnInit, ViewChild } from '@angular/core';
import {
  BinaryImage,
  DeleteConfirmationAction,
  DeleteObjectEvent,
  handleObject,
  UpdateBinaryImage
} from '../../../../classes';
import { ActivatedRoute, Router } from '@angular/router';
import { AuthService } from '../../../../services/auth.service';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from '@ngx-translate/core';
import { BuildService } from '../../../../services/build.service';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { ClrWizard } from '@clr/angular';
import { RepositorysService } from '../../../../services/repositorys.service';
import { ClustersService } from '../../../../services/clusters.service';
import { CustomValidatorDirective, nameValidator } from '../../../../utils/validators/custom-validator.directive';
import { OrganizationService } from '../../../../services/organization.service';
import { DeleteObjectComponent } from '../../../shared-components/delete-object/delete-object.component';
import { CONFIRMATIONSTATE, DELETEOBJECTBUTTON, DELETEOBJECTTYPE } from '../../../../utils/static-entries/shared.const';
import { customTrackBy } from '../../../../utils/operator/customTrackBy';

@Component({
  selector: 'app-build-image-details',
  templateUrl: './binary-image-details.component.html',
  styleUrls: ['./binary-image-details.component.css']
})
export class BinaryImageDetailsComponent implements OnInit {
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
    private registryService: RepositorysService,
    private clusterService: ClustersService,
    private orgService: OrganizationService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  customTrackBy = customTrackBy

  @ViewChild('deleteObjectComponent')
  deleteDialogView: DeleteObjectComponent;

  orgId: string;
  clusterId: string;
  imageName: string;
  image: BinaryImage;

  editBinaryImageOpen = false;

  builderRegistrytList = [];
  builderRegistryEndpoint: string;
  builderProjectList = [];
  builderImageList = [];
  builderTagsList = [];
  pushableRegistrytList = [];
  outputRegistryEndpoint: string;
  outputProjectList = [];
  imageProjectName = '';
  noImageProjectAlert = false;
  noOutputProjectAlert = false;
  noImageNameAlert = false;
  noImageTagAlert = false;

  loading = false;

  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  readonly packagePattern = CustomValidatorDirective.packagePattern;
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

  ngOnInit (): void {
    this.refresh();
  }

  refresh (): void {
    this.loading = true;
    this.clusterId = this.route.snapshot.paramMap.get('id');
    this.imageName = this.route.snapshot.paramMap.get('name');
    this.image = null;
    this.buildService.getBianryImageDetails(this.clusterId, this.imageName).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-BINARY-IMAGE-ERROR'), {
              toastClass: 'alert-danger',
              closeButton: true,
              extendedTimeOut: 0,
              timeOut: 0
            })
          }
        }
      }
      if (data.data?.BinaryImage) {
        this.image = data.data.BinaryImage;
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-BINARY-IMAGE-ERROR'), {
      toastClass: 'alert-danger',
      closeButton: true,
      extendedTimeOut: 0,
      timeOut: 0
    }), () => {
      this.loading = false;
    });
  }

  backToImages (): void {
    this.router.navigate(['/portal/appbuild/binary-image']);
  }

  getAllRegistryForUpdate (clusterId: string, sourceRegistryId: string, outputRegistryId: string): void {
    this.builderRegistrytList = null;
    this.pushableRegistrytList = null;
    this.clusterService.getRegistryOfCluster(clusterId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {
              toastClass: 'alert-danger',
              closeButton: true,
              extendedTimeOut: 0,
              timeOut: 0
            })
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
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {
      toastClass: 'alert-danger',
      closeButton: true,
      extendedTimeOut: 0,
      timeOut: 0
    }))
  }

  getAllBuilderProjectsForUpdate (sourceRegistryId: string): void {
    this.builderProjectList = null;
    this.builderRegistryEndpoint = null;
    this.noImageProjectAlert = false;
    this.registryService.getRegistryProject(sourceRegistryId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-REGISTRY-PROJECT-FAILED'), {
              toastClass: 'alert-danger',
              closeButton: true,
              extendedTimeOut: 0,
              timeOut: 0
            })
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
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-REGISTRY-PROJECT-FAILED'), {
      toastClass: 'alert-danger',
      closeButton: true,
      extendedTimeOut: 0,
      timeOut: 0
    }));
  }

  getAllImagesForUpdate (sourceRegistryId: string, projectName: string): void {
    this.builderImageList = null;
    this.noImageProjectAlert = false;
    this.registryService.getAllImagesByProject(sourceRegistryId, projectName).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-PROJECT-IMAGE-FAILED'), {
              toastClass: 'alert-danger',
              closeButton: true,
              extendedTimeOut: 0,
              timeOut: 0
            })
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
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-PROJECT-IMAGE-FAILED'), {
      toastClass: 'alert-danger',
      closeButton: true,
      extendedTimeOut: 0,
      timeOut: 0
    }));
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

  getAllOutputProjectsForUpdate (registryId: string): void {
    this.outputProjectList = null;
    this.outputRegistryEndpoint = null;
    this.noOutputProjectAlert = false;
    this.registryService.getRegistryProject(registryId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-REGISTRY-PROJECT-FAILED'), {
              toastClass: 'alert-danger',
              closeButton: true,
              extendedTimeOut: 0,
              timeOut: 0
            })
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
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-REGISTRY-PROJECT-FAILED'), {
      toastClass: 'alert-danger',
      closeButton: true,
      extendedTimeOut: 0,
      timeOut: 0
    }));
  }

  changeNoTagsAlert (): void {
    this.noImageTagAlert = false;
  }

  changeNoOutputProjectAlert (): void {
    this.noOutputProjectAlert = false;
  }

  showUpdateImage (): void {
    const selectedBuilderRegistry = this.image.annotations['map.binary-image.source.registryId'];
    const outputImageRegistry = this.image.annotations['map.binary-image.output.registryId'];
    const baseImageList = this.image.spec.baseImage.split('/');
    const selectedBuilderProject = baseImageList[1];
    const builderImage = baseImageList[2];
    const selectedBuilderImage = builderImage.split(':');
    const builderImageName = selectedBuilderImage[0];
    const builderTagName = selectedBuilderImage[1];
    const outputImageList = this.image.spec.distImage.split('/');
    const outputImageProject = outputImageList[1];
    this.editBinaryImageSourceForm.setValue({
      name: this.imageName,
      cluster: this.clusterId,
      sourceURL: this.image.spec.url,
      builderRegistry: selectedBuilderRegistry,
      builderProject: selectedBuilderProject,
      builderImage: builderImageName,
      builderImageTag: builderTagName,
      appImageRegistry: outputImageRegistry,
      appImageProject: outputImageProject,
      deployDir: this.image.spec.binaryDir,
      processId: this.image.spec.processUID
    });
    this.getAllRegistryForUpdate(this.image.cluster.id, selectedBuilderRegistry, outputImageRegistry);
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
            this.toast.show(this.translate.instant('MESSAGE.SHARED.UPDATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.UPDATE-IMAGE-ERROR'), {
              toastClass: 'alert-danger',
              closeButton: true,
              extendedTimeOut: 0,
              timeOut: 0
            })
          }
        }
      }
      if (data.data?.UpdateBinaryImage) {
        this.toast.show(`${data.data.UpdateBinaryImage.name} ` + this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.UPDATE-IMAGE-MSG'), this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.UPDATE-IMAGE-SUCCESS'), {
          toastClass: 'alert-success',
          closeButton: true,
          extendedTimeOut: 10000,
          timeOut: 10000
        })
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.UPDATE-IMAGE-ERROR'), {
      toastClass: 'alert-danger',
      closeButton: true,
      extendedTimeOut: 0,
      timeOut: 0
    }), () => {
      this.doBinaryImageReset();
      this.refresh();
    });
  }

  doUpdateBinaryImageCancel (): void {
    this.doBinaryImageReset();
  }

  private doBinaryImageReset (): void {
    this.wizardeditbinaryimage.reset();
    this.editBinaryImageSourceForm.reset();
    this.editBinaryImageOpen = false;
    this.builderRegistrytList = [];
    this.builderProjectList = [];
    this.builderImageList = [];
    this.builderTagsList = [];
    this.pushableRegistrytList = [];
    this.outputProjectList = [];
    this.imageProjectName = '';
    this.noOutputProjectAlert = false;
    this.noImageProjectAlert = false;
    this.noImageNameAlert = false;
    this.noImageTagAlert = false;
  }

  showDeleteBinaryImage (): void {
    const deleteObjectEvent: DeleteObjectEvent = {
      title: this.translate.instant('BUILD.BINARYIMAGE.ACTION.DELETEIMAGE'),
      message: this.translate.instant('BUILD.BINARYIMAGE.ACTION.DELETEIMAGE-DESC'),
      formInputLabel: this.translate.instant('SHARED.DELETE.DELETEDESC'),
      alerts: [],
      objectName: this.imageName,
      objectType: DELETEOBJECTTYPE.BINARYIMAGEDETAILS,
      button: DELETEOBJECTBUTTON.CONFIRM_CANCEL
    };
    this.deleteDialogView.open(deleteObjectEvent);
  }

  deleteBinaryImage (event: DeleteConfirmationAction): void {
    if (event.objectType === DELETEOBJECTTYPE.BINARYIMAGEDETAILS && event.state === CONFIRMATIONSTATE.CONFIRMED) {
      this.buildService.deleteBinaryImage(this.imageName, this.clusterId).subscribe(data => {
        if (data.errors) {
          for (const each of data.errors) {
            if (each.message) {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.DELETE-IMAGE-ERROR'), {
                toastClass: 'alert-danger',
                closeButton: true,
                extendedTimeOut: 0,
                timeOut: 0
              })
            }
          }
        }
        if (data.data?.DeleteBinaryImage) {
          this.toast.show(`${data.data.DeleteBinaryImage.name} ` + this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.DELETE-IMAGE-MSG'), this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.DELETE-IMAGE-SUCCESS'), {
            toastClass: 'alert-success',
            closeButton: true,
            extendedTimeOut: 10000,
            timeOut: 10000
          });
          this.router.navigate(['/portal/appbuild/binary-image']);
        }
      }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.DELETE-IMAGE-ERROR'), {
        toastClass: 'alert-danger',
        closeButton: true,
        extendedTimeOut: 0,
        timeOut: 0
      }));
    }
  }
}
