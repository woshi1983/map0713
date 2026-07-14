import { Component, ElementRef, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { AuthService } from '../../../services/auth.service';
import {
  ImageRegistry,
  LoadBalancer,
  ClusterProvider,
  UpdateTKGSClusterProvider,
  UpdateTKG2MClusterProvider,
  TKGMClusterProviderSpecNamespace,
  RemoveClusterProviderImageRegistry, AddClusterProviderImageRegistry, DeleteObjectEvent, DeleteConfirmationAction
} from '../../../classes';
import { ProviderService } from '../../../services/provider.service';
import { FormArray, FormControl, FormGroup, Validators } from '@angular/forms';
import { ClrWizard } from '@clr/angular';
import { CustomValidatorDirective, nameValidator } from '../../../utils/validators/custom-validator.directive';
import { LoadbalancerService } from '../../../services/loadbalancer.service';
import { RepositorysService } from '../../../services/repositorys.service';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from '@ngx-translate/core';
import { OrganizationService } from '../../../services/organization.service';
import { prividerNamespacediffset } from 'app/utils/operator/namespacediffset';
import { DELETEOBJECTBUTTON, DELETEOBJECTTYPE, CONFIRMATIONSTATE } from '../../../utils/static-entries/shared.const';
import { DeleteObjectComponent } from '../../shared-components/delete-object/delete-object.component';

@Component({
  selector: 'app-providerdetail',
  templateUrl: './providerdetail.component.html',
  styleUrls: ['./providerdetail.component.css']
})
export class ProviderdetailComponent implements OnInit {
  constructor (
    private route: ActivatedRoute,
    private router: Router,
    public authService: AuthService,
    private providerService: ProviderService,
    private lbService: LoadbalancerService,
    private repositoryService: RepositorysService,
    private toast: ToastrService,
    private translate: TranslateService,
    private orgService: OrganizationService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  @ViewChild('deleteObjectComponent')
  deleteDialogView: DeleteObjectComponent;

  orgId: string
  provider: ClusterProvider;
  providerId: string;
  providerName: string;

  editProviderOpen = false;
  checkProviderConnect = false;
  showConnectAlert = false;
  checkConnectOk = false;
  unallowKeyExist = false;

  registryList: ImageRegistry[] = [];
  removeImageRegistries: RemoveClusterProviderImageRegistry[] = [];
  lbList: LoadBalancer[] = [];
  namespaceList: TKGMClusterProviderSpecNamespace[] = [];
  selectedNsUnexist = false;
  loading = false;

  showAddEditFile = false;
  editFilename: string;

  @ViewChild('fileEditUpload', { static: false }) fileEditUpload: ElementRef;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  readonly endpointPattern = CustomValidatorDirective.endpointPattern;
  readonly CIDRPattern = CustomValidatorDirective.CIDRPattern;
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

  tabLinkStatusList = {
    namespaceActive: true,
    certActive: false
  }

  ngOnInit (): void {
    this.refresh();
  }

  refresh (): void {
    this.loading = true;
    this.providerId = this.route.snapshot.paramMap.get('id');
    this.providerName = this.route.snapshot.paramMap.get('name');
    this.provider = null;
    this.providerService.getProviderById(this.providerId).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.PROVIDER.GET-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.provider = data.data?.ClusterProvider ? data.data.ClusterProvider : null;
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.PROVIDER.GET-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.loading = false;
    });
  }

  backToProvider (): void {
    this.router.navigate(['/portal/system/providers']);
  }

  getALlRegistryName (): void {
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

  showEditProvider (): void {
    this.getAllLBName();
    this.getALlRegistryName();
    this.getAllNamespace(this.providerId, this.provider.spec.defaultNamespace);
    const type = this.provider.class;
    switch (type) {
      case 'TKGS':
        const selectedTKGSRegistryList: ImageRegistry[] = [];
        for (const providerRegistry of this.provider.imageRegistries.edge.nodes) {
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
          name: this.provider.name,
          description: this.provider.description,
          type: this.provider.class,
          registry: selectedTKGSRegistryList,
          enableGlobalLB: false,
          loadBalancerId: ''
        });
        if (this.provider.loadBalancer) {
          this.editProviderForm.patchValue({
            enableGlobalLB: true,
            loadBalancerId: this.provider.loadBalancer.id
          });
        }
        this.editTanzuKubernetesClusterParam.setValue({
          endpoint: this.provider.spec.endpoint,
          caCert: this.provider.spec.caCert,
          skipTLS: this.provider.spec.insecureSkipTLSVerify,
          username: this.provider.spec.authUser,
          password: '',
          namespace: this.provider.spec.defaultNamespace
        })
        break;
      case 'TKG2M':
      case 'TKGM':
        const selectedTKGMRegistryList: ImageRegistry[] = [];
        for (const providerRegistry of this.provider.imageRegistries.edge.nodes) {
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
          name: this.provider.name,
          description: this.provider.description,
          type: this.provider.class === 'TKG2M' ? 'TKGM' : this.provider.class,
          registry: selectedTKGMRegistryList,
          enableGlobalLB: false,
          loadBalancerId: ''
        });
        if (this.provider.loadBalancer) {
          this.editProviderForm.patchValue({
            enableGlobalLB: true,
            loadBalancerId: this.provider.loadBalancer.id
          });
        }
        this.editTkgmKubernetesClusterParam.patchValue({
          kubeConfig: '',
          namespace: this.provider.spec.defaultNamespace
        })
        if (this.provider.spec.defaultEnv) {
          const params = this.provider.spec.defaultEnv;
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
      default:
        console.log('not support type.')
    }
    this.editProviderOpen = true;
  }

  setEditGLBVaildtor (enabled: boolean) {
    if (enabled) {
      this.editProviderForm.controls['loadBalancerId'].setValidators(Validators.required);
      this.editProviderForm.controls['loadBalancerId'].updateValueAndValidity();
    } else {
      this.editProviderForm.controls['loadBalancerId'].clearValidators();
      this.editProviderForm.controls['loadBalancerId'].updateValueAndValidity();
    }
  }

  getSelectedRegistry (): ImageRegistry[] {
    return this.editProviderForm.value.registry;
  }

  showAddEditFileData (): void {
    this.showAddEditFile = true;
  }

  addEditFileToData (): void {
    const fileUpload = this.fileEditUpload.nativeElement;
    const file = fileUpload.files[0];
    if (file) {
      const reader = new FileReader();
      reader.readAsText(file, 'UTF-8');
      reader.onload = () => {
        const fileContext = reader.result;
        this.editTkgmKubernetesClusterParam.patchValue({
          kubeConfig: fileContext
        })
      };
    }
    this.fileEditUpload.nativeElement.value = '';
    this.editFilename = null;
    this.showAddEditFile = false;
  }

  cancelAddEditFile (): void {
    this.fileEditUpload.nativeElement.value = null;
    this.editFilename = null;
    this.showAddEditFile = false;
  }

  getEditFileName (): void {
    const fileUpload = this.fileEditUpload.nativeElement;
    const file = fileUpload.files[0];
    this.editFilename = file ? file.name : '';
  }

  public doEditProviderFinish (): void {
    const type = this.provider.class;
    switch (type) {
      case 'TKGS':
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
          id: this.providerId,
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
          this.refresh();
        });
        break;
      case 'TKG2M':
      case 'TKGM':
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
          id: this.providerId,
          description: this.editProviderForm.value.description,
          spec: {
            kubeConfig: this.editTkgmKubernetesClusterParam.value.kubeConfig ? this.editTkgmKubernetesClusterParam.value.kubeConfig : null,
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
          this.refresh();
        });
        break;
      default:
        console.log('not support.')
    }
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

  checkKeyVaild (key: string): boolean {
    const unallowedKeyList: string[] = ['VSPHERE_NETWORK', 'VSPHERE_RESOURCE_POOL', 'VSPHERE_FOLDER', 'VSPHERE_DATASTORE'];
    this.unallowKeyExist = unallowedKeyList.some(x => x === key);
    return this.unallowKeyExist;
  }

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

  getAllNsList (): TKGMClusterProviderSpecNamespace[] {
    return this.namespaceList;
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
    this.checkProviderConnect = false;
    this.showConnectAlert = false;
    this.checkConnectOk = false;
    this.lbList = [];
    this.registryList = [];
    this.namespaceList = [];
    this.selectedNsUnexist = false;
    this.removeImageRegistries = [];
  }

  public checkUpdateTKGSProviderAuth (): void {
    const input: UpdateTKGSClusterProvider = {
      id: this.providerId,
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
      this.checkProviderConnect = !!data.data;
      this.checkConnectOk = !!data.data;
      this.showConnectAlert = true;
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.PROVIDER.CHECK-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
    );
  }

  public checkUpdateTKGMProviderAuth (kubeconfig: string): void {
    this.providerService.checkUpdateTKGMProviderConnection(this.providerId, kubeconfig).subscribe(data => {
      if (data.errors) {
        for (const each of data.errors) {
          if (each.message) {
            this.toast.show(this.translate.instant('MESSAGE.SHARED.CHECKERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.PROVIDER.CHECK-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
          }
        }
      }
      this.checkProviderConnect = !!data.data;
      this.checkConnectOk = !!data.data;
      this.showConnectAlert = true;
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.PROVIDER.CHECK-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
    );
  }

  showDeleteProvider (): void {
    const deleteObjectEvent: DeleteObjectEvent = {
      title: this.translate.instant('PROVIDERMGMT.ACTION.DELETEPROVIDER'),
      message: this.translate.instant('PROVIDERMGMT.ACTION.DELETEPAGEDESC'),
      formInputLabel: this.translate.instant('SHARED.DELETE.DELETEDESC'),
      alerts: [],
      objectName: this.providerName,
      objectType: DELETEOBJECTTYPE.CLUSTERPROVIDER,
      button: DELETEOBJECTBUTTON.CONFIRM_CANCEL
    };
    this.deleteDialogView.open(deleteObjectEvent);
  }

  deleteProvider (event: DeleteConfirmationAction): void {
    if (event.objectType === DELETEOBJECTTYPE.CLUSTERPROVIDER && event.state === CONFIRMATIONSTATE.CONFIRMED) {
      this.providerService.deleteProvider(this.providerId).subscribe(data => {
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
          this.router.navigate(['/portal/system/providers']);
        }
      }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.PROVIDER.DELETE-PROVIDER-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }));
    }
  }
}
