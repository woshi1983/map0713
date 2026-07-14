import { Component, ElementRef, ViewChild } from '@angular/core';
import { ClrDatagridPagination, ClrDatagridStateInterface, ClrWizard } from '@clr/angular';
import { FormGroup, FormControl, Validators, FormBuilder, FormArray } from '@angular/forms';
import { ProviderService } from '../../../services/provider.service';
import { AuthService } from '../../../services/auth.service';
import { LoadbalancerService } from '../../../services/loadbalancer.service';
import { RepositorysService } from '../../../services/repositorys.service';
import { CustomValidatorDirective, nameValidator } from '../../../utils/validators/custom-validator.directive';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from '@ngx-translate/core';
import {
  ClusterProvider,
  CreateTKGSClusterProvider,
  ImageRegistry,
  LoadBalancer,
  UpdateTKGSClusterProvider,
  CreateTKG2MClusterProvider,
  UpdateTKG2MClusterProvider,
  TKGMClusterProviderSpecNamespace,
  AddClusterProviderImageRegistry,
  RemoveClusterProviderImageRegistry, DeleteObjectEvent, DeleteConfirmationAction
} from '../../../classes/';
import { OrganizationService } from '../../../services/organization.service';
import { prividerNamespacediffset } from 'app/utils/operator/namespacediffset';
import { DELETEOBJECTBUTTON, DELETEOBJECTTYPE, CONFIRMATIONSTATE } from '../../../utils/static-entries/shared.const';
import { DeleteObjectComponent } from '../../shared-components/delete-object/delete-object.component';
import { CheckStatus } from '../../../utils/static-entries/shared.classes';
import { checkCerts, certCheckResult } from "../../../utils/operator/check-certs";


@Component({
  selector: 'app-clusterProvider',
  templateUrl: './provider.component.html',
  styleUrls: ['./provider.component.css']
})
export class ProviderComponent {
  constructor (
    public authService: AuthService,
    private providerService: ProviderService,
    private lbService: LoadbalancerService,
    private repositoryService: RepositorysService,
    private fb: FormBuilder,
    private toast: ToastrService,
    private translate: TranslateService,
    private orgService: OrganizationService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  @ViewChild('deleteObjectComponent')
  deleteDialogView: DeleteObjectComponent;

  orgId: string;
  providers: ClusterProvider[] = [];
  selectedProvider: ClusterProvider;
  showSearch = true;
  searchName: string;
  loading = true;
  total = 0;

  addProviderOpen = false;
  showProviderDetailsModel = false;
  editProviderOpen = false;

  showAddFile = false;
  updated = false;
  filename = '';
  unallowKeyExist = false;

  registryList: ImageRegistry[] = [];
  removeImageRegistries: RemoveClusterProviderImageRegistry[] = [];
  lbList: LoadBalancer[] = [];
  namespaceList: TKGMClusterProviderSpecNamespace[] = [];
  noSupervisorNSAlert = false;
  selectedNsUnexist = false;

  checkStatus: CheckStatus = {
    checked: false,
    checkSuccess: false
  };

  certResult: certCheckResult[] = [];

  @ViewChild('fileUpload', { static: false }) fileUpload: ElementRef;
  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardprovider') wizardnewprovider: ClrWizard;
  readonly endpointPattern = CustomValidatorDirective.endpointPattern;
  readonly CIDRPattern = CustomValidatorDirective.CIDRPattern;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  basicProviderForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(48), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    type: new FormControl('', Validators.required),
    registry: new FormControl('', Validators.required),
    defaultRegistry: new FormControl('', Validators.required),
    enableGlobalLB: new FormControl(false),
    loadBalancerId: new FormControl('')
  });

  tanzuKubernetesClusterParam = new FormGroup({
    endpoint: new FormControl('', [Validators.required, Validators.pattern(this.endpointPattern)]),
    caCert: new FormControl(''),
    skipTLS: new FormControl(false),
    username: new FormControl('', Validators.required),
    password: new FormControl('', Validators.required),
    namespace: new FormControl('', Validators.required)
  });

  envList = new FormArray([]);
  tkgmKubernetesClusterParam = new FormGroup({
    kubeConfig: new FormControl('', [Validators.required]),
    namespace: new FormControl('', Validators.required),
    newNamespace: new FormControl(''),
    vsphereNetwork: new FormControl(''),
    resourcePool: new FormControl(''),
    vmFolder: new FormControl(''),
    datastore: new FormControl(''),
    aviDataSeGroup: new FormControl(''),
    aviDataNetworkName: new FormControl(''),
    aviDataNetworkCIDR: new FormControl('', Validators.pattern(this.CIDRPattern)),
    defaultEnv: this.envList
  });

  @ViewChild('wizardeditprovider') wizardeditprovider: ClrWizard;
  editProviderForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    type: new FormControl('', Validators.required),
    registry: new FormControl('', Validators.required),
    defaultRegistry: new FormControl('', Validators.required),
    enableGlobalLB: new FormControl(false),
    loadBalancerId: new FormControl('')
  });

  editTanzuKubernetesClusterParam = new FormGroup({
    endpoint: new FormControl('', [Validators.required, Validators.pattern(this.endpointPattern)]),
    caCert: new FormControl(''),
    skipTLS: new FormControl(false),
    username: new FormControl('', Validators.required),
    password: new FormControl(''),
    namespace: new FormControl('', Validators.required)
  });

  editEnvList = new FormArray([]);
  editTkgmKubernetesClusterParam = new FormGroup({
    kubeConfig: new FormControl(''),
    namespace: new FormControl('', Validators.required),
    newNamespace: new FormControl(''),
    vsphereNetwork: new FormControl(''),
    resourcePool: new FormControl(''),
    vmFolder: new FormControl(''),
    datastore: new FormControl(''),
    aviDataSeGroup: new FormControl(''),
    aviDataNetworkName: new FormControl(''),
    aviDataNetworkCIDR: new FormControl('', Validators.pattern(this.CIDRPattern)),
    defaultEnv: this.editEnvList
  });

  refresh (state?: ClrDatagridStateInterface): void {
    this.loading = true;
    const filters = [{ key: 'organization', value: [this.orgId] }];
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.providerService.getAllProviders(state?.page?.current ? state.page.current - 1 : 0, state?.page?.size ? state.page.size : Number(this.pagination.pageSize), filters).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.PROVIDER.GET-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.providers = data.data?.ClusterProviders?.edge?.nodes ? data.data.ClusterProviders.edge.nodes : [];
      this.selectedProvider = null;
      this.total = data.data?.ClusterProviders?.totalCount ? data.data.ClusterProviders.totalCount : 0;
      this.certResult = checkCerts(this.providers);
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.PROVIDER.GET-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => { this.loading = false });
  }

  getAllRegistryName (): void {
    this.repositoryService.getImageRegistrysNames(this.orgId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.REPOSITORY.GET-REPO-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.registryList = data.data?.ImageRegistries?.edge?.nodes ? data.data.ImageRegistries.edge.nodes : [];
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.REPOSITORY.GET-REPO-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
  }

  getAllLBName (): void {
    this.lbService.getLbNames(this.orgId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.LOADBALANCER.GET-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.lbList = data.data?.LoadBalancers?.edge?.nodes ? data.data.LoadBalancers.edge.nodes : [];
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.LOADBALANCER.GET-LB-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
  }

  getAllNamespace (providerId: string, ns: string): void {
    this.providerService.getAllNamespaceName(providerId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.PROVIDERDETAILS.NAMESPACE.GET-SUPERVISR-NS-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.ClusterProvider?.spec?.namespaces) {
        this.namespaceList = prividerNamespacediffset(data.data.ClusterProvider.spec.namespaces);
        if (!this.namespaceList.find(x => x.name === ns)) {
          this.selectedNsUnexist = true;
        }
      } else {
        this.namespaceList = [];
        this.selectedNsUnexist = true;
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.PROVIDERDETAILS.NAMESPACE.GET-SUPERVISR-NS-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
  }

  showAddProvider (): void {
    this.getAllLBName();
    this.getAllRegistryName();
    this.addProviderOpen = true;
  }

  resetCheckStatus (): void {
    this.checkStatus.checkSuccess = false;
    this.checkStatus.checked = false;
  }

  addNewClusterEnv (): void {
    this.envList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  deleteClusterEnv (index: number): void {
    this.envList.removeAt(index);
  }

  showAddFileData (update?: boolean): void {
    this.showAddFile = true;
    this.updated = update;
  }

  addFileToData (): void {
    const fileUpload = this.fileUpload.nativeElement;
    const file = fileUpload.files[0];
    if (file) {
      const reader = new FileReader();
      reader.readAsText(file, 'UTF-8');
      reader.onload = () => {
        const fileContext = reader.result;
        if (!this.updated) {
          this.tkgmKubernetesClusterParam.patchValue({
            kubeConfig: fileContext
          })
        } else {
          this.editTkgmKubernetesClusterParam.patchValue({
            kubeConfig: fileContext
          })
        }
      };
    }
    this.cancelAddFile();
  }

  cancelAddFile (): void {
    this.fileUpload.nativeElement.value = null;
    this.filename = null;
    this.showAddFile = false;
    this.updated = false;
  }

  getFileName (): void {
    const fileUpload = this.fileUpload.nativeElement;
    const file = fileUpload.files[0];
    this.filename = file ? file.name : '';
  }

  checkKeyVaild (key: string): boolean {
    const unallowedKeyList: string[] = ['VSPHERE_NETWORK', 'VSPHERE_RESOURCE_POOL', 'VSPHERE_FOLDER', 'VSPHERE_DATASTORE', 'AVI_SERVICE_ENGINE_GROUP', 'AVI_DATA_NETWORK', 'AVI_DATA_NETWORK_CIDR'];
    this.unallowKeyExist = unallowedKeyList.some(x => x === key);
    return this.unallowKeyExist;
  }

  setGLBVaildtor (enabled: boolean): void {
    if (enabled) {
      this.basicProviderForm.controls['loadBalancerId'].setValidators(Validators.required);
      this.basicProviderForm.controls['loadBalancerId'].updateValueAndValidity();
    } else {
      this.basicProviderForm.controls['loadBalancerId'].clearValidators();
      this.basicProviderForm.controls['loadBalancerId'].updateValueAndValidity();
    }
  }

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  changeFormVaild ($event: any): void {
    const ns = $event.target.value;
    this.selectedNsUnexist = false;
    if (ns === 'CreateNew') {
      this.tkgmKubernetesClusterParam.controls['newNamespace'].setValidators(Validators.required);
      this.tkgmKubernetesClusterParam.controls['newNamespace'].updateValueAndValidity();
    } else {
      this.tkgmKubernetesClusterParam.controls['newNamespace'].clearValidators();
      this.tkgmKubernetesClusterParam.controls['newNamespace'].setErrors(null);
      this.tkgmKubernetesClusterParam.controls['newNamespace'].updateValueAndValidity();
    }
  }

  getSelectedRegistry (update?: boolean): ImageRegistry[] {
    return update ? this.editProviderForm.value.registry : this.basicProviderForm.value.registry;
  }

  private doProviderReset (): void {
    this.wizardnewprovider.reset();
    this.basicProviderForm.reset();
    this.basicProviderForm.patchValue({
      enableGlobalLB: false
    })
    this.setGLBVaildtor(false);
    this.tanzuKubernetesClusterParam.reset();
    this.tanzuKubernetesClusterParam.patchValue({
      skipTLS: false
    });
    this.tkgmKubernetesClusterParam.reset();
    while (this.envList.length !== 0) {
      this.envList.removeAt(0);
    }
    this.addProviderOpen = false;
    this.lbList = [];
    this.registryList = [];
    this.namespaceList = [];
    this.unallowKeyExist = false;
    this.noSupervisorNSAlert = false;
    this.resetCheckStatus();
  }

  public doProviderFinish (): void {
    const name = this.basicProviderForm.value.name;
    const desc = this.basicProviderForm.value.description;
    const type = this.basicProviderForm.value.type;
    switch (type) {
      case 'TKGS': {
        const addTKGSImageRegistry: AddClusterProviderImageRegistry[] = [];
        for (const registry of this.basicProviderForm.value.registry) {
          if (registry.id !== this.basicProviderForm.value.defaultRegistry) {
            addTKGSImageRegistry.push({
              id: registry.id,
              default: false
            })
          } else {
            addTKGSImageRegistry.push({
              id: registry.id,
              default: true
            })
          }
        }
        const tkgsParam: CreateTKGSClusterProvider = {
          name: name,
          description: desc,
          spec: {
            endpoint: this.tanzuKubernetesClusterParam.value.endpoint,
            caCert: !this.tanzuKubernetesClusterParam.value.skipTLS ? this.tanzuKubernetesClusterParam.value.caCert : null,
            insecureSkipTLSVerify: this.tanzuKubernetesClusterParam.value.skipTLS,
            authUser: this.tanzuKubernetesClusterParam.value.username,
            authPassword: this.tanzuKubernetesClusterParam.value.password,
            defaultNamespace: this.tanzuKubernetesClusterParam.value.namespace
          },
          imageRegistries: addTKGSImageRegistry,
          loadBalancer: this.basicProviderForm.value.enableGlobalLB ? this.basicProviderForm.value.loadBalancerId : null
        }
        this.providerService.createTKGSProvider(tkgsParam).subscribe(data => {
          if (data.errors) {
            for (const each of data.errors) {
              if (each.message) {
                this.toast.show(this.translate.instant('MESSAGE.SHARED.ADDERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.PROVIDER.ADD-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
              }
            }
          }
          if (data.data?.CreateTKGSClusterProvider) {
            this.toast.show(`${data.data.CreateTKGSClusterProvider.name} ` + this.translate.instant('MESSAGE.PROVIDER.ADD-PROVIDER-MSG'), this.translate.instant('MESSAGE.PROVIDER.ADD-PROVIDER-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
          }
        }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.PROVIDER.ADD-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
          this.doProviderReset();
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
        });
        break;
      }
      case 'TKGM': {
        const addTKGMImageRegistry: AddClusterProviderImageRegistry[] = [];
        for (const registry of this.basicProviderForm.value.registry) {
          if (registry.id !== this.basicProviderForm.value.defaultRegistry) {
            addTKGMImageRegistry.push({
              id: registry.id,
              default: false
            })
          } else {
            addTKGMImageRegistry.push({
              id: registry.id,
              default: true
            })
          }
        }
        const tkgmParam: CreateTKG2MClusterProvider = {
          name: name,
          description: desc,
          spec: {
            kubeConfig: this.tkgmKubernetesClusterParam.value.kubeConfig,
            defaultNamespace: this.tkgmKubernetesClusterParam.value.namespace === 'CreateNew' ? this.tkgmKubernetesClusterParam.value.newNamespace : this.tkgmKubernetesClusterParam.value.namespace,
            vsphereNetwork: this.tkgmKubernetesClusterParam.value.vsphereNetwork,
            vsphereResourcePool: this.tkgmKubernetesClusterParam.value.resourcePool,
            vsphereFolder: this.tkgmKubernetesClusterParam.value.vmFolder,
            vsphereDatastore: this.tkgmKubernetesClusterParam.value.datastore,
            aviServiceEngineGroup: this.tkgmKubernetesClusterParam.value.aviDataSeGroup,
            aviDataNetwork: this.tkgmKubernetesClusterParam.value.aviDataNetworkName,
            aviDataNetworkCidr: this.tkgmKubernetesClusterParam.value.aviDataNetworkCIDR
          },
          imageRegistries: addTKGMImageRegistry,
          loadBalancer: this.basicProviderForm.value.enableGlobalLB ? this.basicProviderForm.value.loadBalancerId : null
        }
        this.providerService.createTKGMProvider(tkgmParam).subscribe(data => {
          if (data.errors) {
            for (const each of data.errors) {
              if (each.message) {
                this.toast.show(this.translate.instant('MESSAGE.SHARED.ADDERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.PROVIDER.ADD-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
              }
            }
          }
          if (data.data?.CreateTKG2MClusterProvider) {
            this.toast.show(`${data.data.CreateTKG2MClusterProvider.name} ` + this.translate.instant('MESSAGE.PROVIDER.ADD-PROVIDER-MSG'), this.translate.instant('MESSAGE.PROVIDER.ADD-PROVIDER-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
          }
        }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.PROVIDER.ADD-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
          this.doProviderReset();
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
        });
        break;
      }
      default:
        console.log('type is not support.')
    }
  }

  public doProviderCancel (): void {
    this.doProviderReset();
  }

  public checkCreateTKGSProviderAuth (): void {
    this.namespaceList = null;
    const input: CreateTKGSClusterProvider = {
      name: this.basicProviderForm.value.name,
      spec: {
        endpoint: this.tanzuKubernetesClusterParam.value.endpoint,
        caCert: !this.tanzuKubernetesClusterParam.value.skipTLS ? this.tanzuKubernetesClusterParam.value.caCert : null,
        insecureSkipTLSVerify: this.tanzuKubernetesClusterParam.value.skipTLS,
        authUser: this.tanzuKubernetesClusterParam.value.username,
        authPassword: this.tanzuKubernetesClusterParam.value.password,
        defaultNamespace: 'demo'
      },
      imageRegistries: [{
        id: this.basicProviderForm.value.defaultRegistry,
        default: true
      }],
      validate: true,
      dryrun: true
    };
    this.providerService.checkCreateTKGSProviderConnection(input).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.CHECKERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.PROVIDER.CHECK-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.CreateTKGSClusterProvider?.spec?.namespaces) {
        this.checkStatus.checkSuccess = true;
        this.namespaceList = prividerNamespacediffset(data.data.CreateTKGSClusterProvider.spec.namespaces);
        this.noSupervisorNSAlert = !data.data.CreateTKGSClusterProvider.spec.namespaces;
      } else {
        this.checkStatus.checkSuccess = false;
        this.namespaceList = [];
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.PROVIDER.CHECK-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.checkStatus.checked = true;
    });
  }

  getAllNsList (): TKGMClusterProviderSpecNamespace[] {
    return this.namespaceList;
  }

  public checkCreateTKGMProviderAuth (kubeconfig: string, registry: string): void {
    this.namespaceList = null;
    this.providerService.checkCreateTKGMProviderConnection(kubeconfig, registry).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.CHECKERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.PROVIDER.CHECK-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (!!data.data) {
        this.checkStatus.checkSuccess = true;
        if(data.data?.CreateTKG2MClusterProvider?.spec?.namespaces) {
          this.namespaceList = prividerNamespacediffset(data.data.CreateTKG2MClusterProvider.spec.namespaces);
        } else {
          this.namespaceList = [];
        }
      } else {
        this.checkStatus.checkSuccess = false;
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.PROVIDER.CHECK-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.checkStatus.checked = true;
    });
  }

  showProviderDetails (provider: ClusterProvider): void {
    this.selectedProvider = provider;
    if (provider.class === 'TKGM' || provider.class === 'TKG2M') {
      if (provider.spec.defaultEnv) {
        const params = provider.spec.defaultEnv;
        for (const key of Object.keys(params)) {
          if (key === 'VSPHERE_NETWORK') {
            this.editTkgmKubernetesClusterParam.patchValue({
              vsphereNetwork: params[key]
            })
          } else if (key === 'VSPHERE_RESOURCE_POOL') {
            this.editTkgmKubernetesClusterParam.patchValue({
              resourcePool: params[key]
            })
          } else if (key === 'VSPHERE_FOLDER') {
            this.editTkgmKubernetesClusterParam.patchValue({
              vmFolder: params[key]
            })
          } else if (key === 'VSPHERE_DATASTORE') {
            this.editTkgmKubernetesClusterParam.patchValue({
              datastore: params[key]
            })
          } else if (key === 'AVI_SERVICE_ENGINE_GROUP') {
            this.editTkgmKubernetesClusterParam.patchValue({
              aviDataSeGroup: params[key]
            })
          } else if (key === 'AVI_DATA_NETWORK') {
            this.editTkgmKubernetesClusterParam.patchValue({
              aviDataNetworkName: params[key]
            })
          } else if (key === 'AVI_DATA_NETWORK_CIDR') {
            this.editTkgmKubernetesClusterParam.patchValue({
              aviDataNetworkCIDR: params[key]
            })
          } else {
            this.editEnvList.push(new FormGroup({
              key: new FormControl(key),
              value: new FormControl(params[key])
            }));
          }
        }
      }
    }
    this.showProviderDetailsModel = true;
  }

  public doProviderDetailsReset (): void {
    this.showProviderDetailsModel = false;
    this.selectedProvider = null;
    this.editTkgmKubernetesClusterParam.reset();
    while (this.editEnvList.length !== 0) {
      this.editEnvList.removeAt(0);
    }
  }

  showEditProvider (provider: ClusterProvider): void {
    this.getAllLBName();
    this.getAllRegistryName();
    this.getAllNamespace(provider.id, provider.spec.defaultNamespace);
    this.selectedProvider = provider;
    const type = provider.class;
    switch (type) {
      case 'TKGS': {
        const selectedTKGSRegistryList: ImageRegistry[] = [];
        for (const providerRegistry of provider.imageRegistries.edge.nodes) {
          if (providerRegistry.default) {
            this.editProviderForm.patchValue({
              defaultRegistry: providerRegistry.imageRegistry.id
            })
          }
          selectedTKGSRegistryList.push({
            id: providerRegistry.imageRegistry.id,
            name: providerRegistry.imageRegistry.name
          })
          this.removeImageRegistries.push({
            id: providerRegistry.imageRegistry.id
          })
        }
        this.editProviderForm.patchValue({
          name: provider.name,
          description: provider.description,
          type: provider.class,
          registry: selectedTKGSRegistryList,
          enableGlobalLB: false,
          loadBalancerId: ''
        });
        if (provider.loadBalancer) {
          this.editProviderForm.patchValue({
            enableGlobalLB: true,
            loadBalancerId: provider.loadBalancer.id
          });
        }
        this.editTanzuKubernetesClusterParam.setValue({
          endpoint: provider.spec.endpoint,
          caCert: provider.spec.caCert,
          skipTLS: provider.spec.insecureSkipTLSVerify,
          username: provider.spec.authUser,
          password: '',
          namespace: provider.spec.defaultNamespace
        })
        break;
      }
      case 'TKG2M':
      case 'TKGM': {
        const selectedTKGMRegistryList: ImageRegistry[] = [];
        for (const providerRegistry of provider.imageRegistries.edge.nodes) {
          if (providerRegistry.default) {
            this.editProviderForm.patchValue({
              defaultRegistry: providerRegistry.imageRegistry.id
            })
          }
          selectedTKGMRegistryList.push({
            id: providerRegistry.imageRegistry.id,
            name: providerRegistry.imageRegistry.name
          })
          this.removeImageRegistries.push({
            id: providerRegistry.imageRegistry.id
          })
        }
        this.editProviderForm.patchValue({
          name: provider.name,
          description: provider.description,
          type: provider.class === 'TKG2M' ? 'TKGM' : provider.class,
          registry: selectedTKGMRegistryList,
          enableGlobalLB: false,
          loadBalancerId: ''
        });
        if (provider.loadBalancer) {
          this.editProviderForm.patchValue({
            enableGlobalLB: true,
            loadBalancerId: provider.loadBalancer.id
          });
        }
        this.editTkgmKubernetesClusterParam.patchValue({
          kubeConfig: '',
          namespace: provider.spec.defaultNamespace
        })
        if (provider.spec.defaultEnv) {
          const params = provider.spec.defaultEnv;
          for (const key of Object.keys(params)) {
            if (key === 'VSPHERE_NETWORK') {
              this.editTkgmKubernetesClusterParam.patchValue({
                vsphereNetwork: params[key]
              })
            } else if (key === 'VSPHERE_RESOURCE_POOL') {
              this.editTkgmKubernetesClusterParam.patchValue({
                resourcePool: params[key]
              })
            } else if (key === 'VSPHERE_FOLDER') {
              this.editTkgmKubernetesClusterParam.patchValue({
                vmFolder: params[key]
              })
            } else if (key === 'VSPHERE_DATASTORE') {
              this.editTkgmKubernetesClusterParam.patchValue({
                datastore: params[key]
              })
            } else if (key === 'AVI_SERVICE_ENGINE_GROUP') {
              this.editTkgmKubernetesClusterParam.patchValue({
                aviDataSeGroup: params[key]
              })
            } else if (key === 'AVI_DATA_NETWORK') {
              this.editTkgmKubernetesClusterParam.patchValue({
                aviDataNetworkName: params[key]
              })
            } else if (key === 'AVI_DATA_NETWORK_CIDR') {
              this.editTkgmKubernetesClusterParam.patchValue({
                aviDataNetworkCIDR: params[key]
              })
            } else {
              this.editEnvList.push(new FormGroup({
                key: new FormControl(key),
                value: new FormControl(params[key])
              }));
            }
          }
        }
        break;
      }
      default:
        console.log('not support type.')
    }
    this.editProviderOpen = true;
  }

  addNewEditClusterEnv (): void {
    this.editEnvList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  deleteEditClusterEnv (index: number): void {
    this.editEnvList.removeAt(index);
  }

  setEditGLBVaildtor (enabled: boolean): void {
    if (enabled) {
      this.editProviderForm.controls['loadBalancerId'].setValidators(Validators.required);
      this.editProviderForm.controls['loadBalancerId'].updateValueAndValidity();
    } else {
      this.editProviderForm.controls['loadBalancerId'].clearValidators();
      this.editProviderForm.controls['loadBalancerId'].updateValueAndValidity();
    }
  }

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  changeEditFormVaild ($event: any): void {
    const ns = $event.target.value;
    this.selectedNsUnexist = false;
    if (ns === 'CreateNew') {
      this.editTkgmKubernetesClusterParam.controls['newNamespace'].setValidators(Validators.required);
      this.editTkgmKubernetesClusterParam.controls['newNamespace'].updateValueAndValidity();
    } else {
      this.editTkgmKubernetesClusterParam.controls['newNamespace'].clearValidators();
      this.editTkgmKubernetesClusterParam.controls['newNamespace'].setErrors(null);
      this.editTkgmKubernetesClusterParam.controls['newNamespace'].updateValueAndValidity();
    }
  }

  public doEditProviderFinish (): void {
    const type = this.editProviderForm.value.type;
    switch (type) {
      case 'TKGS': {
        const addTKGSImageRegistry: AddClusterProviderImageRegistry[] = [];
        for (const registry of this.editProviderForm.value.registry) {
          if (registry.id !== this.editProviderForm.value.defaultRegistry) {
            addTKGSImageRegistry.push({
              id: registry.id,
              default: false
            })
          } else {
            addTKGSImageRegistry.push({
              id: registry.id,
              default: true
            })
          }
        }
        const tkgsParam: UpdateTKGSClusterProvider = {
          id: this.selectedProvider?.id,
          description: this.editProviderForm.value.description,
          spec: {
            endpoint: this.editTanzuKubernetesClusterParam.value.endpoint,
            caCert: !this.editTanzuKubernetesClusterParam.value.skipTLS ? this.editTanzuKubernetesClusterParam.value.caCert : null,
            insecureSkipTLSVerify: this.editTanzuKubernetesClusterParam.value.skipTLS,
            authUser: this.editTanzuKubernetesClusterParam.value.username,
            defaultNamespace: this.editTanzuKubernetesClusterParam.value.namespace
          },
          addImageRegistries: addTKGSImageRegistry,
          removeImageRegistries: this.removeImageRegistries,
          loadBalancer: this.editProviderForm.value.enableGlobalLB ? this.editProviderForm.value.loadBalancerId : null
        }
        if (this.editTanzuKubernetesClusterParam.value.password) {
          tkgsParam.spec.authPassword = this.editTanzuKubernetesClusterParam.value.password
        }
        this.providerService.updateTKGSProvider(tkgsParam).subscribe(data => {
          if (data.errors) {
            for (const each of data.errors) {
              if (each.message) {
                this.toast.show(this.translate.instant('MESSAGE.SHARED.UPDATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.PROVIDER.UPDATE-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
              }
            }
          }
          if (data.data?.UpdateTKGSClusterProvider) {
            this.toast.show(`${data.data.UpdateTKGSClusterProvider.name} ` + this.translate.instant('MESSAGE.PROVIDER.UPDATE-PROVIDER-MSG'), this.translate.instant('MESSAGE.PROVIDER.UPDATE-PROVIDER-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
          }
        }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.PROVIDER.UPDATE-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
          this.doEditProviderReset();
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
        });
        break;
      }
      case 'TKGM': {
        const addTKGMImageRegistry: AddClusterProviderImageRegistry[] = [];
        for (const registry of this.editProviderForm.value.registry) {
          if (registry.id !== this.editProviderForm.value.defaultRegistry) {
            addTKGMImageRegistry.push({
              id: registry.id,
              default: false
            })
          } else {
            addTKGMImageRegistry.push({
              id: registry.id,
              default: true
            })
          }
        }
        const tkgmParam: UpdateTKG2MClusterProvider = {
          id: this.selectedProvider?.id,
          description: this.editProviderForm.value.description,
          spec: {
            defaultNamespace: this.editTkgmKubernetesClusterParam.value.namespace === 'CreateNew' ? this.editTkgmKubernetesClusterParam.value.newNamespace : this.editTkgmKubernetesClusterParam.value.namespace,
            vsphereNetwork: this.editTkgmKubernetesClusterParam.value.vsphereNetwork,
            vsphereResourcePool: this.editTkgmKubernetesClusterParam.value.resourcePool,
            vsphereFolder: this.editTkgmKubernetesClusterParam.value.vmFolder,
            vsphereDatastore: this.editTkgmKubernetesClusterParam.value.datastore,
            aviServiceEngineGroup: this.editTkgmKubernetesClusterParam.value.aviDataSeGroup,
            aviDataNetwork: this.editTkgmKubernetesClusterParam.value.aviDataNetworkName,
            aviDataNetworkCidr: this.editTkgmKubernetesClusterParam.value.aviDataNetworkCIDR
          },
          addImageRegistries: addTKGMImageRegistry,
          removeImageRegistries: this.removeImageRegistries,
          loadBalancer: this.editProviderForm.value.enableGlobalLB ? this.editProviderForm.value.loadBalancerId : null
        }
        if (this.editTkgmKubernetesClusterParam.value.kubeConfig) {
          tkgmParam.spec.kubeConfig = this.editTkgmKubernetesClusterParam.value.kubeConfig
        }
        this.providerService.updateTKGMProvider(tkgmParam).subscribe(data => {
          if (data.errors) {
            for (const each of data.errors) {
              if (each.message) {
                this.toast.show(this.translate.instant('MESSAGE.SHARED.UPDATEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.PROVIDER.UPDATE-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
              }
            }
          }
          if (data.data?.UpdateTKG2MClusterProvider) {
            this.toast.show(`${data.data.UpdateTKG2MClusterProvider.name} ` + this.translate.instant('MESSAGE.PROVIDER.UPDATE-PROVIDER-MSG'), this.translate.instant('MESSAGE.PROVIDER.UPDATE-PROVIDER-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
          }
        }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.PROVIDER.UPDATE-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
          this.doEditProviderReset();
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
        });
        break;
      }
      default:
        console.log('not support.')
    }
  }

  public doEditProviderCancel (): void {
    this.doEditProviderReset();
  }

  private doEditProviderReset (): void {
    this.wizardeditprovider.reset();
    this.editProviderForm.reset();
    this.editProviderForm.patchValue({
      enableGlobalLB: false
    });
    this.setEditGLBVaildtor(false);
    this.editTanzuKubernetesClusterParam.reset();
    this.editTkgmKubernetesClusterParam.reset();
    while (this.editEnvList.length !== 0) {
      this.editEnvList.removeAt(0);
    }
    this.editProviderOpen = false;
    this.lbList = [];
    this.registryList = [];
    this.namespaceList = [];
    this.selectedNsUnexist = false;
    this.removeImageRegistries = [];
    this.resetCheckStatus();
  }

  public checkUpdateTKGSProviderAuth (): void {
    this.namespaceList = null;
    const input: UpdateTKGSClusterProvider = {
      id: this.selectedProvider?.id,
      spec: {
        endpoint: this.editTanzuKubernetesClusterParam.value.endpoint,
        caCert: !this.editTanzuKubernetesClusterParam.value.skipTLS ? this.editTanzuKubernetesClusterParam.value.caCert : null,
        insecureSkipTLSVerify: this.editTanzuKubernetesClusterParam.value.skipTLS,
        authUser: this.editTanzuKubernetesClusterParam.value.username,
        authPassword: this.editTanzuKubernetesClusterParam.value.password
      },
      validate: true,
      dryrun: true
    };
    this.providerService.checkUpdateTKGSProviderConnection(input).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.CHECKERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.PROVIDER.CHECK-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.UpdateTKGSClusterProvider?.spec?.namespaces) {
        this.checkStatus.checkSuccess = true;
        this.namespaceList = prividerNamespacediffset(data.data.UpdateTKGSClusterProvider.spec.namespaces);
        if (!this.namespaceList.find(x => x.name === this.selectedProvider.spec.defaultNamespace)) {
          this.selectedNsUnexist = true;
        }
      } else {
        this.checkStatus.checkSuccess = false;
        this.namespaceList = [];
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.PROVIDER.CHECK-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.checkStatus.checked = true;
    });
  }

  public checkUpdateTKGMProviderAuth (kubeconfig: string): void {
    this.namespaceList = null;
    this.providerService.checkUpdateTKGMProviderConnection(this.selectedProvider?.id, kubeconfig).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.CHECKERROR') + ` ${each.message}${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.PROVIDER.CHECK-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      if (data.data?.UpdateTKG2MClusterProvider?.spec?.namespaces) {
        this.checkStatus.checkSuccess = true;
        this.namespaceList = prividerNamespacediffset(data.data.UpdateTKG2MClusterProvider.spec.namespaces);
        if (!this.namespaceList.find(x => x.name === this.selectedProvider.spec.defaultNamespace)) {
          this.selectedNsUnexist = true;
        }
      } else {
        this.checkStatus.checkSuccess = false;
        this.namespaceList = [];
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.PROVIDER.CHECK-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.checkStatus.checked = true;
    });
  }

  showDeleteProvider (selectedProvider: ClusterProvider): void {
    this.selectedProvider = selectedProvider;
    const deleteObjectEvent: DeleteObjectEvent = {
      title: this.translate.instant('PROVIDERMGMT.ACTION.DELETEPROVIDER'),
      message: this.translate.instant('PROVIDERMGMT.ACTION.DELETEPAGEDESC'),
      formInputLabel: this.translate.instant('SHARED.DELETE.DELETEDESC'),
      alerts: [],
      objectName: selectedProvider.name,
      objectType: DELETEOBJECTTYPE.CLUSTERPROVIDER,
      button: DELETEOBJECTBUTTON.CONFIRM_CANCEL
    };
    this.deleteDialogView.open(deleteObjectEvent);
  }

  deleteProvider (event: DeleteConfirmationAction): void {
    if (event.objectType === DELETEOBJECTTYPE.CLUSTERPROVIDER && event.state === CONFIRMATIONSTATE.CONFIRMED) {
      this.providerService.deleteProvider(this.selectedProvider?.id).subscribe(data => {
        if (data.errors) {
          for (const each of data.errors) {
            if (each.message) {
              if (each.message) {
                if (each.message === 'object is busy') {
                  this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ' 请先解除集群关联后再执行此操作！', this.translate.instant('MESSAGE.PROVIDER.DELETE-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
                } else {
                  this.toast.show(this.translate.instant('MESSAGE.SHARED.DELETEERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.PROVIDER.DELETE-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
                }
              }
            }
          }
        }
        if (data.data?.DeleteClusterProvider) {
          this.toast.show(`${data.data.DeleteClusterProvider.name} ` + this.translate.instant('MESSAGE.PROVIDER.DELETE-PROVIDER-MSG'), this.translate.instant('MESSAGE.PROVIDER.DELETE-PROVIDER-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 })
        }
      }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.PROVIDER.DELETE-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
        if (this.pagination.currentPage === 1) {
          this.refresh();
        } else {
          this.pagination.currentPage = 1;
        }
      });
    }
  }
}
