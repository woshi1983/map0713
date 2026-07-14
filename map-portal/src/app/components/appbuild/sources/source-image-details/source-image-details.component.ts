import { Component, OnInit, ViewChild } from '@angular/core';
import {
  DeleteConfirmationAction,
  DeleteObjectEvent,
  handleObject,
  SourceImage,
  UpdateSourceImage
} from '../../../../classes';
import { ActivatedRoute, Router } from '@angular/router';
import { AuthService } from '../../../../services/auth.service';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from '@ngx-translate/core';
import { BuildService } from '../../../../services/build.service';
import { FormArray, FormControl, FormGroup, Validators } from '@angular/forms';
import { ClrWizard } from '@clr/angular';
import { RepositorysService } from '../../../../services/repositorys.service';
import { ClustersService } from '../../../../services/clusters.service';
import { CustomValidatorDirective, nameValidator } from '../../../../utils/validators/custom-validator.directive';
import { OrganizationService } from '../../../../services/organization.service';
import { customTrackBy } from '../../../../utils/operator/customTrackBy';
import { DeleteObjectComponent } from '../../../shared-components/delete-object/delete-object.component';
import { CONFIRMATIONSTATE, DELETEOBJECTBUTTON, DELETEOBJECTTYPE } from '../../../../utils/static-entries/shared.const';

@Component({
  selector: 'app-source-image-details',
  templateUrl: './source-image-details.component.html',
  styleUrls: ['./source-image-details.component.css']
})
export class SourceImageDetailsComponent implements OnInit {
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
    private orgService: OrganizationService,
    private clusterService: ClustersService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  customTrackBy = customTrackBy;

  @ViewChild('deleteObjectComponent')
  deleteDialogView: DeleteObjectComponent;

  orgId: string;
  clusterId: string;
  imageName: string;
  image: SourceImage;
  registryId = '';
  source: string;

  editSourceImageOpen = false;

  pushableRegistrytList = [];
  outputRegistryEndpoint: string;
  outputProjectList = [];
  imageRegistrytList = [];
  imageRegistryEndpoint: string;
  imageProjectList = [];
  sourceImageList = [];
  sourceTagsList = [];
  noImageProjectAlert = false;
  noImageNameAlert = false;
  noImageTagAlert = false;
  noOutputProjectAlert = false;

  loading = false;

  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  readonly packagePattern = CustomValidatorDirective.packagePattern;
  readonly gitUrlPattern = CustomValidatorDirective.gitUrlPattern;
  @ViewChild('wizardeditsourceimage') wizardeditsourceimage: ClrWizard;
  editSourceImageForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(64), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    cluster: new FormControl('', [Validators.required]),
    sourceType: new FormControl('git', [Validators.required]),
    appImageRegistry: new FormControl('', [Validators.required]),
    appImageProject: new FormControl('', [Validators.required])
  });

  editGitEnvs = new FormArray([]);
  editGitSourceParamForm = new FormGroup({
    url: new FormControl('', [Validators.required, Validators.pattern(this.gitUrlPattern)]),
    gitBranch: new FormControl('master'),
    subdir: new FormControl(''),
    envs: this.editGitEnvs
  });

  editBlobEnvs = new FormArray([]);
  editBolbSourceParamForm = new FormGroup({
    url: new FormControl('', [Validators.required, Validators.pattern(this.packagePattern)]),
    subdir: new FormControl(''),
    envs: this.editBlobEnvs
  });

  editImageEnvs = new FormArray([]);
  editImageSourceParamForm = new FormGroup({
    sourceImageRegistry: new FormControl('', [Validators.required]),
    sourceImageProject: new FormControl('', [Validators.required]),
    sourceImageName: new FormControl('', [Validators.required]),
    sourceImageTag: new FormControl('', [Validators.required]),
    subdir: new FormControl(''),
    envs: this.editImageEnvs
  });

  ngOnInit (): void {
    this.refresh();
  }

  refresh (): void {
    this.loading = true;
    this.clusterId = this.route.snapshot.paramMap.get('id');
    this.imageName = this.route.snapshot.paramMap.get('name');
    this.image = null;
    this.buildService.getSourceImageDetails(this.clusterId, this.imageName).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.GET-SOURCE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.SourceImage) {
        this.image = data.data.SourceImage;
        this.source = this.image.spec.source.git != null ? 'git' : this.image.spec.source.blob != null ? 'blob' : 'registry';
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.GET-SOURCE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.loading = false;
    });
  }

  // TODO: need merge edit to shared module

  backToImages (): void {
    this.router.navigate(['/portal/appbuild/source-image']);
  }

  getAllOutputProjects (registryId: string, updated?: boolean): void {
    this.outputProjectList = null;
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
          this.noOutputProjectAlert = !this.outputProjectList.find(x => x === this.editSourceImageForm.value.appImageProject);
        }
      } else {
        this.noOutputProjectAlert = true;
        this.outputProjectList = [];
        this.outputRegistryEndpoint = '';
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-REGISTRY-PROJECT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
  }

  getAllTags (imageName: string, updated?: boolean): void {
    this.sourceTagsList = [];
    this.noImageNameAlert = false;
    for (const repo of this.sourceImageList) {
      if (repo.name === imageName) {
        this.sourceTagsList = repo.tags;
        if (updated) {
          if (!this.sourceTagsList.find(x => x === this.editImageSourceParamForm.value.sourceImageTag)) {
            this.noImageTagAlert = true;
          }
        }
      }
    }
  }

  changeNoTagsAlert (): void {
    this.noImageTagAlert = false;
  }

  changeNoOutputProjectAlert (): void {
    this.noOutputProjectAlert = false;
  }

  getAllRegistryForUpdate (clusterId: string, outputRegistryId: string): void {
    this.pushableRegistrytList = null;
    this.imageRegistrytList = null;
    this.clusterService.getRegistryOfCluster(clusterId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.Cluster?.imageRegistries?.edge?.nodes) {
        this.imageRegistrytList = data.data.Cluster.imageRegistries.edge.nodes;
        const registryList = [];
        for (const registry of this.imageRegistrytList) {
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

  getAllProjectsForUpdate (sourceRegistryId: string): void {
    this.imageProjectList = null;
    this.imageRegistryEndpoint = null;
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
        this.imageRegistryEndpoint = data.data.ImageRegistry.endpoint;
        this.imageProjectList = data.data.ImageRegistry.projects;
        if (!this.imageProjectList.find(x => x === this.editImageSourceParamForm.value.sourceImageProject)) {
          this.noImageProjectAlert = true;
        } else {
          this.getAllImagesForUpdate(sourceRegistryId, this.editImageSourceParamForm.value.sourceImageProject);
        }
      } else {
        this.noImageProjectAlert = true;
        this.imageProjectList = [];
        this.imageRegistryEndpoint = '';
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-REGISTRY-PROJECT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
  }

  getAllImagesForUpdate (sourceRegistryId: string, projectName: string): void {
    this.sourceImageList = null;
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
        this.sourceImageList = data.data.ImageRepos;
        if (!this.sourceImageList.find(x => x.name === this.editImageSourceParamForm.value.sourceImageName)) {
          this.noImageNameAlert = true;
        } else {
          this.getAllTags(this.editImageSourceParamForm.value.sourceImageName, true);
        }
      } else {
        this.noImageNameAlert = true;
        this.sourceImageList = [];
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-PROJECT-IMAGE-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
  }

  editGitNewEnv (): void {
    this.editGitEnvs.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  editDeleteGitEnv (index: number): void {
    this.editGitEnvs.removeAt(index);
  }

  editBlobNewEnv (): void {
    this.editBlobEnvs.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  editDeleteBlobEnv (index: number): void {
    this.editBlobEnvs.removeAt(index);
  }

  editRegistryNewEnv (): void {
    this.editImageEnvs.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  editDeleteRegistryEnv (index: number): void {
    this.editImageEnvs.removeAt(index);
  }

  showUpdateSourceImage (): void {
    const outputImageRegistry = this.image.annotations['map.binary-image.output.registryId'];
    const outputImageList = this.image.spec.distImage.split('/');
    const outputImageProject = outputImageList[1];
    this.editSourceImageForm.patchValue({
      name: this.imageName,
      appImageRegistry: outputImageRegistry,
      appImageProject: outputImageProject,
      cluster: this.clusterId
    });
    if (this.image.spec.source.git != null) {
      this.getAllRegistryForUpdate(this.image.cluster.id, outputImageRegistry);
      this.editSourceImageForm.patchValue({
        sourceType: 'git'
      });
      this.editGitSourceParamForm.patchValue({
        url: this.image.spec.source.git.url,
        gitBranch: this.image.spec.source.git.revision,
        subdir: !!this.image.spec.source.subPath ? this.image.spec.source.subPath : ''
      });
      if (this.image.spec.env) {
        const env = this.image.spec.env;
        for (const key of Object.keys(env)) {
          this.editGitEnvs.push(new FormGroup({
            key: new FormControl(key),
            value: new FormControl(env[key])
          }));
        }
      }
    } else if (this.image.spec.source.blob != null) {
      this.getAllRegistryForUpdate(this.image.cluster.id, outputImageRegistry);
      this.editSourceImageForm.patchValue({
        sourceType: 'blob'
      });
      this.editBolbSourceParamForm.patchValue({
        url: this.image.spec.source.blob.url,
        subdir: !!this.image.spec.source.subPath ? this.image.spec.source.subPath : ''
      });
      if (this.image.spec.env) {
        const env = this.image.spec.env;
        for (const key of Object.keys(env)) {
          this.editBlobEnvs.push(new FormGroup({
            key: new FormControl(key),
            value: new FormControl(env[key])
          }));
        }
      }
    } else {
      const sourceImageRegistry = this.image.annotations['map.binary-image.source.registryId'];
      this.editSourceImageForm.patchValue({
        sourceType: 'registry'
      });
      const imageList = this.image.spec.source.registry.image.split('/', 3);
      const sourceImageProject = imageList[1];
      const sourceImage = imageList[2];
      const sourceImageList = sourceImage.split(':', 2);
      const sourceImageName = sourceImageList[0];
      const sourceImageTag = sourceImageList[1];
      this.editImageSourceParamForm.patchValue({
        sourceImageRegistry: sourceImageRegistry,
        sourceImageProject: sourceImageProject,
        sourceImageName: sourceImageName,
        sourceImageTag: sourceImageTag,
        subdir: !!this.image.spec.source.subPath ? this.image.spec.source.subPath : ''
      });
      if (this.image.spec.env) {
        const env = this.image.spec.env;
        for (const key of Object.keys(env)) {
          this.editImageEnvs.push(new FormGroup({
            key: new FormControl(key),
            value: new FormControl(env[key])
          }));
        }
      }
      this.getAllRegistryForUpdate(this.image.cluster.id, outputImageRegistry);
      this.getAllProjectsForUpdate(sourceImageRegistry);
    }
    this.editSourceImageOpen = true;
  }

  doUpdateSourceImageFinish (): void {
    const name = this.editSourceImageForm.value.name;
    const cluster = this.editSourceImageForm.value.cluster;
    const outputImageProject = this.editSourceImageForm.value.appImageProject;
    const outputRegistryURL = new URL(this.outputRegistryEndpoint);
    const outputImage = `${outputRegistryURL.host}/${outputImageProject}/${name}`;
    const sourceType = this.editSourceImageForm.value.sourceType;
    const annoations: JSON = JSON.parse(JSON.stringify({}));
    annoations['map.binary-image.output.registryId'] = this.editSourceImageForm.value.appImageRegistry;
    switch (sourceType) {
      case 'git':
        const gEnvsMap: JSON = JSON.parse(JSON.stringify({}));
        if (this.editGitSourceParamForm.value.envs) {
          for (const each of this.editGitSourceParamForm.value.envs) {
            gEnvsMap[each.key] = each.value;
          }
        }
        const gBuildParam: UpdateSourceImage = {
          name: name,
          annotations: annoations,
          spec: {
            source: {
              git: {
                url: this.editGitSourceParamForm.value.url,
                revision: this.editGitSourceParamForm.value.gitBranch ? this.editGitSourceParamForm.value.gitBranch : 'master'
              },
              subPath: this.editGitSourceParamForm.value.subdir ? this.editGitSourceParamForm.value.subdir : null
            },
            env: gEnvsMap,
            distImage: outputImage
          },
          cluster: cluster
        };
        this.buildService.updateSourceImage(gBuildParam).subscribe(data => {
          if (data.errors) {
            for (const each of data.errors) {
              if (each.message) {
                this.toast.show(this.translate.instant('MESSAGE.SHARED.UPDATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.UPDATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
              }
            }
          }
          if (data.data?.UpdateSourceImage) {
            this.toast.show(`${data.data.UpdateSourceImage.name} ` + this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.UPDATE-IMAGE-MSG'), this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.UPDATE-IMAGE-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
          }
        }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.UPDATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
          this.doEditSourceImageReset();
          this.refresh();
        });
        break;
      case 'blob':
        const bEnvsMap: JSON = JSON.parse(JSON.stringify({}));
        if (this.editBolbSourceParamForm.value.envs) {
          for (const each of this.editBolbSourceParamForm.value.envs) {
            bEnvsMap[each.key] = each.value;
          }
        }
        const bBuildParam: UpdateSourceImage = {
          name: name,
          annotations: annoations,
          spec: {
            source: {
              blob: {
                url: this.editBolbSourceParamForm.value.url
              },
              subPath: this.editBolbSourceParamForm.value.subdir ? this.editBolbSourceParamForm.value.subdir : null
            },
            env: bEnvsMap,
            distImage: outputImage
          },
          cluster: cluster
        };
        this.buildService.updateSourceImage(bBuildParam).subscribe(data => {
          if (data.errors) {
            for (const each of data.errors) {
              if (each.message) {
                this.toast.show(this.translate.instant('MESSAGE.SHARED.UPDATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.UPDATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
              }
            }
          }
          if (data.data?.UpdateSourceImage) {
            this.toast.show(`${data.data.UpdateSourceImage.name} ` + this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.UPDATE-IMAGE-MSG'), this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.UPDATE-IMAGE-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
          }
        }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.UPDATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
          this.doEditSourceImageReset();
          this.refresh();
        });
        break;
      case 'registry':
        annoations['map.binary-image.source.registryId'] = this.editImageSourceParamForm.value.sourceImageRegistry;
        const imageRegistryURL = new URL(this.imageRegistryEndpoint);
        const sourceImageProject = this.editImageSourceParamForm.value.sourceImageProject;
        const sourceImageName = this.editImageSourceParamForm.value.sourceImageName;
        const sourceImageTag = this.editImageSourceParamForm.value.sourceImageTag;
        const image = `${imageRegistryURL.host}/${sourceImageProject}/${sourceImageName}:${sourceImageTag}`;
        const rEnvsMap: JSON = JSON.parse(JSON.stringify({}));
        if (this.editImageSourceParamForm.value.envs) {
          for (const each of this.editImageSourceParamForm.value.envs) {
            rEnvsMap[each.key] = each.value;
          }
        }
        const rBuildParam: UpdateSourceImage = {
          name: name,
          annotations: annoations,
          spec: {
            source: {
              registry: {
                image: image
              },
              subPath: this.editImageSourceParamForm.value.subdir ? this.editImageSourceParamForm.value.subdir : null
            },
            env: rEnvsMap,
            distImage: outputImage
          },
          cluster: cluster
        };
        this.buildService.updateSourceImage(rBuildParam).subscribe(data => {
          if (data.errors) {
            for (const each of data.errors) {
              if (each.message) {
                this.toast.show(this.translate.instant('MESSAGE.SHARED.UPDATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.UPDATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
              }
            }
          }
          if (data.data?.UpdateSourceImage) {
            this.toast.show(`${data.data.UpdateSourceImage.name} ` + this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.UPDATE-IMAGE-MSG'), this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.UPDATE-IMAGE-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
          }
        }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.UPDATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
          this.doEditSourceImageReset();
          this.refresh();
        });
        break;
    }
  }

  private doEditSourceImageReset (): void {
    this.editSourceImageOpen = false;
    this.editSourceImageForm.reset();
    this.wizardeditsourceimage.reset();
    this.editImageSourceParamForm.reset();
    this.editBolbSourceParamForm.reset();
    this.editGitSourceParamForm.reset();
    while (this.editGitEnvs.length !== 0) {
      this.editGitEnvs.removeAt(0);
    }
    while (this.editBlobEnvs.length !== 0) {
      this.editBlobEnvs.removeAt(0);
    }
    while (this.editImageEnvs.length !== 0) {
      this.editImageEnvs.removeAt(0);
    }
    this.imageRegistrytList = [];
    this.imageProjectList = [];
    this.registryId = '';
    this.sourceImageList = [];
    this.sourceTagsList = [];
    this.noImageProjectAlert = false;
    this.noImageNameAlert = false;
    this.noImageTagAlert = false;
    this.noOutputProjectAlert = false;
    this.pushableRegistrytList = [];
    this.outputRegistryEndpoint = '';
    this.outputProjectList = [];
    this.imageRegistryEndpoint = '';
  }

  doUpdateSourceImageCancel (): void {
    this.doEditSourceImageReset();
  }

  showDeleteImage (): void {
    const deleteObjectEvent: DeleteObjectEvent = {
      title: this.translate.instant('BUILD.SOURCEIMAGE.ACTION.DELETEIMAGE'),
      message: this.translate.instant('BUILD.SOURCEIMAGE.ACTION.DELETEIMAGE-DESC'),
      formInputLabel: this.translate.instant('SHARED.DELETE.DELETEDESC'),
      alerts: [],
      objectName: this.imageName,
      objectType: DELETEOBJECTTYPE.SOURCEIMAGEDETAILS,
      button: DELETEOBJECTBUTTON.CONFIRM_CANCEL
    };
    this.deleteDialogView.open(deleteObjectEvent);
  }

  deleteSourceImage (event: DeleteConfirmationAction): void {
    if (event.objectType === DELETEOBJECTTYPE.SOURCEIMAGEDETAILS && event.state === CONFIRMATIONSTATE.CONFIRMED) {
      this.buildService.deleteSourceImage(this.imageName, this.clusterId).subscribe(data => {
        if (data.errors) {
          for (const each of data.errors) {
            if (each.message) {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.DELETE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            }
          }
        }
        if (data.data?.DeleteSourceImage) {
          this.toast.show(`${data.data.DeleteSourceImage.name} ` + this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.DELETE-IMAGE-MSG'), this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.DELETE-IMAGE-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
          this.router.navigate(['/portal/appbuild/source-image']);
        }
      }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.DELETE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
    }
  }
}
