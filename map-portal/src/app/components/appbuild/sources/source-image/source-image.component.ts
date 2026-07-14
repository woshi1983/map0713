import { Component, OnInit, ViewChild } from '@angular/core';
import { BuildService } from '../../../../services/build.service';
import { AuthService } from '../../../../services/auth.service';
import { ActivatedRoute } from '@angular/router';
import {
  SourceImage,
  CreateSourceImage,
  UpdateSourceImage,
  Cluster, handleObject, DeleteObjectEvent, DeleteConfirmationAction
} from '../../../../classes';
import { ClrDatagridPagination, ClrWizard } from '@clr/angular';
import { CustomValidatorDirective, nameValidator } from '../../../../utils/validators/custom-validator.directive';
import { FormArray, FormControl, FormGroup, Validators } from '@angular/forms';
import { ClustersService } from '../../../../services/clusters.service';
import { RepositorysService } from '../../../../services/repositorys.service';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from '@ngx-translate/core';
import { ApplicationsService } from '../../../../services/applications.service';
import { OrganizationService } from '../../../../services/organization.service';
import { DeleteObjectComponent } from '../../../shared-components/delete-object/delete-object.component';
import { customTrackBy } from '../../../../utils/operator/customTrackBy';
import { CONFIRMATIONSTATE, DELETEOBJECTBUTTON, DELETEOBJECTTYPE } from '../../../../utils/static-entries/shared.const';

@Component({
  selector: 'app-source-image',
  templateUrl: './source-image.component.html',
  styleUrls: ['./source-image.component.css']
})
export class SourceImageComponent implements OnInit {
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

  customTrackBy = customTrackBy

  @ViewChild('deleteObjectComponent')
  deleteDialogView: DeleteObjectComponent;

  orgId: string;
  allImages: SourceImage[] = [];
  selectedImage: SourceImage;

  allClusters: Cluster[] = [];
  noTbsEnabledClusterAlert = false;

  createSourceImageOpen = false;
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

  loading = true;
  showSearch = true;
  searchName: string;

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardsourceimage') wizardnewsourceimage: ClrWizard;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  readonly packagePattern = CustomValidatorDirective.packagePattern;
  readonly gitUrlPattern = CustomValidatorDirective.gitUrlPattern;
  basicSourceImageForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(64), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    cluster: new FormControl('', [Validators.required]),
    sourceType: new FormControl('git', [Validators.required]),
    appImageRegistry: new FormControl('', [Validators.required]),
    appImageProject: new FormControl('', [Validators.required])
  });

  envsList = new FormArray([]);
  gitSourceParamForm = new FormGroup({
    url: new FormControl('', [Validators.required, Validators.pattern(this.gitUrlPattern)]),
    gitBranch: new FormControl('master'),
    subdir: new FormControl(''),
    envs: this.envsList
  });

  bolbSourceParamForm = new FormGroup({
    url: new FormControl('', [Validators.required, Validators.pattern(this.packagePattern)]),
    subdir: new FormControl(''),
    envs: this.envsList
  });

  imageSourceParamForm = new FormGroup({
    sourceImageRegistry: new FormControl('', [Validators.required]),
    sourceImageProject: new FormControl('', [Validators.required]),
    sourceImageName: new FormControl('', [Validators.required]),
    sourceImageTag: new FormControl('', [Validators.required]),
    subdir: new FormControl(''),
    envs: this.envsList
  });

  @ViewChild('wizardeditsourceimage') wizardeditsourceimage: ClrWizard;
  editSourceImageForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(64), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    cluster: new FormControl('', [Validators.required]),
    sourceType: new FormControl('git', [Validators.required]),
    appImageRegistry: new FormControl('', [Validators.required]),
    appImageProject: new FormControl('', [Validators.required])
  });

  editEnvsList = new FormArray([]);
  editGitSourceParamForm = new FormGroup({
    url: new FormControl('', [Validators.required, Validators.pattern(this.gitUrlPattern)]),
    gitBranch: new FormControl('master'),
    subdir: new FormControl(''),
    envs: this.editEnvsList
  });

  editBolbSourceParamForm = new FormGroup({
    url: new FormControl('', [Validators.required, Validators.pattern(this.packagePattern)]),
    subdir: new FormControl(''),
    envs: this.editEnvsList
  });

  editImageSourceParamForm = new FormGroup({
    sourceImageRegistry: new FormControl('', [Validators.required]),
    sourceImageProject: new FormControl('', [Validators.required]),
    sourceImageName: new FormControl('', [Validators.required]),
    sourceImageTag: new FormControl('', [Validators.required]),
    subdir: new FormControl(''),
    envs: this.editEnvsList
  });

  ngOnInit (): void {
    this.clusterService.getExtensionEnabledClusterId(this.orgId, 'tbs').subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.noTbsEnabledClusterAlert = data.data?.Clusters?.totalCount === 0;
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }))
  }

  refresh (): void {
    this.loading = true;
    const filters = [{ key: 'organization', value: [this.orgId] }];
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.buildService.getSourceImages(filters).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.GET-SOURCE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.allImages = data.data?.SourceImages ? data.data.SourceImages : [];
      this.selectedImage = null;
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.GET-SOURCE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.loading = false;
    });
  }

  getAllRegistry (clusterId: string): void {
    this.pushableRegistrytList = null;
    this.imageRegistrytList = null;
    for (const cluster of this.allClusters) {
      if (cluster.id === clusterId) {
        this.imageRegistrytList = cluster.imageRegistries.edge.nodes;
      }
    }
    // check pushable
    const registryList = [];
    for (const registry of this.imageRegistrytList) {
      if (registry.imageRegistry.authUser) {
        registryList.push(registry);
      }
    }
    this.pushableRegistrytList = registryList;
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

  getAllProjects (registryId: string): void {
    this.imageProjectList = null;
    this.imageRegistryEndpoint = null;
    this.registryService.getRegistryProject(registryId).subscribe(data => {
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
      } else {
        this.imageProjectList = [];
        this.imageRegistryEndpoint = '';
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-REGISTRY-PROJECT-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
  }

  getAllImages (projectName: string): void {
    this.sourceImageList = null;
    this.registryService.getAllImagesByProject(this.imageSourceParamForm.value.sourceImageRegistry, projectName).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-PROJECT-IMAGE-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.ImageRepos) {
        this.sourceImageList = data.data.ImageRepos;
      } else {
        this.noImageNameAlert = true;
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.BINARY-IMAGE.GET-PROJECT-IMAGE-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
  }

  getAllTags (imageName: string, updated?: boolean): void {
    this.sourceTagsList = null;
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

  changeNoTagsAlert (): void {
    this.noImageTagAlert = false;
  }

  changeNoOutputProjectAlert (): void {
    this.noOutputProjectAlert = false;
  }

  resetOtherTypeForm (type: string): void {
    switch (type) {
      case 'git':
        this.bolbSourceParamForm.reset();
        this.imageSourceParamForm.reset();
        break;
      case 'blob':
        this.gitSourceParamForm.reset();
        this.gitSourceParamForm.patchValue({
          gitBranch: 'master'
        });
        this.imageSourceParamForm.reset();
        break;
      case 'registry':
        this.gitSourceParamForm.reset();
        this.gitSourceParamForm.patchValue({
          gitBranch: 'master'
        });
        this.bolbSourceParamForm.reset();
        break;
      default:
        console.error('does not support type ' + type)
    }
    while (this.envsList.length !== 0) {
      this.envsList.removeAt(0);
    }
  }

  addNewEnv (): void {
    this.envsList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  deleteEnv (index: number): void {
    this.envsList.removeAt(index);
  }

  newEditEnv (): void {
    this.editEnvsList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  deleteEditEnv (index: number): void {
    this.editEnvsList.removeAt(index);
  }

  showCreateSourceImage (): void {
    this.clusterService.getTbsEnabledClusterId(this.orgId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.allClusters = data.data?.Clusters?.edge?.nodes ? data.data.Clusters.edge.nodes : [];
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }))
    this.createSourceImageOpen = true;
  }

  private doSourceImageReset (): void {
    this.wizardnewsourceimage.reset();
    this.basicSourceImageForm.reset();
    this.basicSourceImageForm.patchValue({
      sourceType: 'git'
    });
    this.gitSourceParamForm.reset();
    this.gitSourceParamForm.patchValue({
      gitBranch: 'master'
    });
    this.bolbSourceParamForm.reset();
    this.imageSourceParamForm.reset();
    while (this.envsList.length !== 0) {
      this.envsList.removeAt(0);
    }
    this.createSourceImageOpen = false;
    this.imageRegistrytList = [];
    this.imageProjectList = [];
    this.allClusters = [];
    this.sourceImageList = [];
    this.sourceTagsList = [];
    this.pushableRegistrytList = [];
    this.outputProjectList = [];
    this.outputRegistryEndpoint = '';
    this.imageRegistryEndpoint = '';
  }

  doSourceImageFinish (): void {
    const name = this.basicSourceImageForm.value.name;
    const outputImageProject = this.basicSourceImageForm.value.appImageProject;
    const outputRegistryURL = new URL(this.outputRegistryEndpoint);
    const outputImage = `${outputRegistryURL.host}/${outputImageProject}/${name}`;
    const annoations: JSON = JSON.parse(JSON.stringify({}));
    annoations['map.binary-image.output.registryId'] = this.basicSourceImageForm.value.appImageRegistry;
    const EnvsMap: JSON = JSON.parse(JSON.stringify({}));
    if (this.envsList.length > 0) {
      for (const each of this.envsList.controls) {
        EnvsMap[each.value.key] = each.value.value;
      }
    }
    const BuildParam: CreateSourceImage = {
      name: name,
      annotations: annoations,
      spec: {
        source: {},
        env: EnvsMap,
        distImage: outputImage
      },
      cluster: this.basicSourceImageForm.value.cluster
    };
    const sourceType = this.basicSourceImageForm.value.sourceType;
    switch (sourceType) {
      case 'git': {
        BuildParam.spec.source = {
          git: {
            url: this.gitSourceParamForm.value.url,
            revision: this.gitSourceParamForm.value.gitBranch ? this.gitSourceParamForm.value.gitBranch : 'master'
          },
          subPath: this.gitSourceParamForm.value.subdir ? this.gitSourceParamForm.value.subdir : null
        };
        break;
      }
      case 'blob': {
        BuildParam.spec.source = {
          blob: {
            url: this.bolbSourceParamForm.value.url
          },
          subPath: this.bolbSourceParamForm.value.subdir ? this.bolbSourceParamForm.value.subdir : null
        };
        break;
      }
      case 'registry': {
        annoations['map.binary-image.source.registryId'] = this.imageSourceParamForm.value.sourceImageRegistry;
        const imageRegistryURL = new URL(this.imageRegistryEndpoint);
        const sourceImageProject = this.imageSourceParamForm.value.sourceImageProject;
        const sourceImageName = this.imageSourceParamForm.value.sourceImageName;
        const sourceImageTag = this.imageSourceParamForm.value.sourceImageTag;
        const image = `${imageRegistryURL.host}/${sourceImageProject}/${sourceImageName}:${sourceImageTag}`;
        BuildParam.spec.source = {
          registry: {
            image: image
          },
          subPath: this.imageSourceParamForm.value.subdir ? this.imageSourceParamForm.value.subdir : null
        };
        break;
      }
      default:
        console.log('not support type.');
        break;
    }
    this.buildService.createSourceImage(BuildParam).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.CREATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.CREATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.CreateSourceImage) {
        this.toast.show(`${data.data.CreateSourceImage.name} ` + this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.CREATE-IMAGE-MSG'), this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.CREATE-IMAGE-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
        // source Image build can't start auto because of TBS init.
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.CREATE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.doSourceImageReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doSourceImageCancel (): void {
    this.doSourceImageReset();
  }

  showUpdateSourceImage (image: SourceImage): void {
    this.selectedImage = image;
    const outputImageRegistry = image.annotations['map.binary-image.output.registryId'];
    const outputImageList = image.spec.distImage.split('/');
    const outputImageProject = outputImageList[1];
    this.editSourceImageForm.patchValue({
      name: image.name,
      appImageRegistry: outputImageRegistry,
      appImageProject: outputImageProject,
      cluster: image.cluster.id
    });
    if (image.spec.env) {
      const env = image.spec.env;
      for (const key of Object.keys(env)) {
        this.editEnvsList.push(new FormGroup({
          key: new FormControl(key),
          value: new FormControl(env[key])
        }));
      }
    }
    if (image.spec.source.git != null) {
      this.getAllRegistryForUpdate(image.cluster.id, outputImageRegistry);
      this.editSourceImageForm.patchValue({
        sourceType: 'git'
      });
      this.editGitSourceParamForm.patchValue({
        url: image.spec.source.git.url,
        gitBranch: image.spec.source.git.revision,
        subdir: !!image.spec.source.subPath ? image.spec.source.subPath : ''
      });
    } else if (image.spec.source.blob != null) {
      this.getAllRegistryForUpdate(image.cluster.id, outputImageRegistry);
      this.editSourceImageForm.patchValue({
        sourceType: 'blob'
      });
      this.editBolbSourceParamForm.patchValue({
        url: image.spec.source.blob.url,
        subdir: !!image.spec.source.subPath ? image.spec.source.subPath : ''
      });
    } else {
      const sourceImageRegistry = image.annotations['map.binary-image.source.registryId'];
      this.getAllRegistryForUpdate(image.cluster.id, outputImageRegistry);
      this.getAllProjectsForUpdate(sourceImageRegistry);
      this.editSourceImageForm.patchValue({
        sourceType: 'registry'
      });
      const imageList = image.spec.source.registry.image.split('/', 3);
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
        subdir: !!image.spec.source.subPath ? image.spec.source.subPath : ''
      });
    }
    this.editSourceImageOpen = true;
  }

  doUpdateSourceImageFinish (): void {
    const name = this.editSourceImageForm.value.name;
    const outputImageProject = this.editSourceImageForm.value.appImageProject;
    const outputRegistryURL = new URL(this.outputRegistryEndpoint);
    const outputImage = `${outputRegistryURL.host}/${outputImageProject}/${name}`;
    const annoations: JSON = JSON.parse(JSON.stringify({}));
    annoations['map.binary-image.output.registryId'] = this.editSourceImageForm.value.appImageRegistry;
    const EnvsMap: JSON = JSON.parse(JSON.stringify({}));
    if (this.editEnvsList.length > 0) {
      for (const each of this.editEnvsList.controls) {
        EnvsMap[each.value.key] = each.value.value;
      }
    }
    const BuildParam: UpdateSourceImage = {
      name: name,
      annotations: annoations,
      spec: {
        source: {},
        env: EnvsMap,
        distImage: outputImage
      },
      cluster: this.editSourceImageForm.value.cluster
    };
    const sourceType = this.editSourceImageForm.value.sourceType;
    switch (sourceType) {
      case 'git': {
        BuildParam.spec.source = {
          git: {
            url: this.editGitSourceParamForm.value.url,
            revision: this.editGitSourceParamForm.value.gitBranch ? this.editGitSourceParamForm.value.gitBranch : 'master'
          },
          subPath: this.editGitSourceParamForm.value.subdir ? this.editGitSourceParamForm.value.subdir : null
        };
        break;
      }
      case 'blob': {
        BuildParam.spec.source = {
          blob: {
            url: this.editBolbSourceParamForm.value.url
          },
          subPath: this.editBolbSourceParamForm.value.subdir ? this.editBolbSourceParamForm.value.subdir : null
        };
        break;
      }
      case 'registry': {
        annoations['map.binary-image.source.registryId'] = this.editImageSourceParamForm.value.sourceImageRegistry;
        const imageRegistryURL = new URL(this.imageRegistryEndpoint);
        const sourceImageProject = this.editImageSourceParamForm.value.sourceImageProject;
        const sourceImageName = this.editImageSourceParamForm.value.sourceImageName;
        const sourceImageTag = this.editImageSourceParamForm.value.sourceImageTag;
        const image = `${imageRegistryURL.host}/${sourceImageProject}/${sourceImageName}:${sourceImageTag}`;
        BuildParam.spec.source = {
          registry: {
            image: image
          },
          subPath: this.editImageSourceParamForm.value.subdir ? this.editImageSourceParamForm.value.subdir : null
        };
        break;
      }
      default:
        console.log('not support type.');
    }
    this.buildService.updateSourceImage(BuildParam).subscribe(data => {
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
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  private doEditSourceImageReset (): void {
    this.editSourceImageOpen = false;
    this.editSourceImageForm.reset();
    this.wizardeditsourceimage.reset();
    this.editImageSourceParamForm.reset();
    this.editBolbSourceParamForm.reset();
    this.editGitSourceParamForm.reset();
    while (this.editEnvsList.length !== 0) {
      this.editEnvsList.removeAt(0);
    }
    this.imageRegistrytList = [];
    this.imageProjectList = [];
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

  showDeleteSourceImage (selectedImage: SourceImage): void {
    this.selectedImage = selectedImage;
    const deleteObjectEvent: DeleteObjectEvent = {
      title: this.translate.instant('BUILD.SOURCEIMAGE.ACTION.DELETEIMAGE'),
      message: this.translate.instant('BUILD.SOURCEIMAGE.ACTION.DELETEIMAGE-DESC'),
      formInputLabel: this.translate.instant('SHARED.DELETE.DELETEDESC'),
      alerts: [],
      objectName: selectedImage.name,
      objectType: DELETEOBJECTTYPE.SOURCEIMAGE,
      button: DELETEOBJECTBUTTON.CONFIRM_CANCEL
    };
    this.deleteDialogView.open(deleteObjectEvent);
  }

  deleteSourceImage (event: DeleteConfirmationAction): void {
    if (event.objectType === DELETEOBJECTTYPE.BINARYIMAGE && event.state === CONFIRMATIONSTATE.CONFIRMED) {
      this.buildService.deleteSourceImage(this.selectedImage?.name, this.selectedImage?.cluster?.id).subscribe(data => {
        if (data.errors) {
          for (const each of data.errors) {
            if (each.message) {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.DELETE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            }
          }
        }
        if (data.data?.DeleteSourceImage) {
          this.toast.show(`${data.data.DeleteSourceImage.name} ` + this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.DELETE-IMAGE-MSG'), this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.DELETE-IMAGE-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
        }
      }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.BUILDS.SOURCE-IMAGE.DELETE-IMAGE-ERROR'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
        if (this.pagination.currentPage === 1) {
          this.refresh();
        } else {
          this.pagination.currentPage = 1;
        }
      });
    }
  }
}
