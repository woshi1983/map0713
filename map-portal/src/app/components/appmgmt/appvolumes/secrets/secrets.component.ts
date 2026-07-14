import {Component, ElementRef, OnInit, ViewChild} from '@angular/core';
import {
  Cluster,
  CreateSecret,
  Namespace,
  Secret,
  UpdateSecret,
  handleObject
} from "../../../../classes";
import { AuthService} from "../../../../services/auth.service";
import {ClustersService} from "../../../../services/clusters.service";
import {VolumeService} from "../../../../services/volume.service";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {ClrDatagridPagination, ClrWizard} from "@clr/angular";
import {KEYFormatValidator, CAFormatValidator, CustomValidatorDirective, nameValidator} from "../../../../utils/validators/custom-validator.directive";
import {FormArray, FormControl, FormGroup, Validators} from "@angular/forms";
import {OrganizationService} from "../../../../services/organization.service";
import {namespacediffset} from "../../../../utils/operator";

@Component({
  selector: 'app-secrets',
  templateUrl: './secrets.component.html',
  styleUrls: ['./secrets.component.css']
})
export class SecretsComponent implements OnInit {

  //admin and user role can handle cluster resources
  get canWriteResources(): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER_RESOURCE);
  }

  constructor(
     public authService: AuthService,
     private clusterService: ClustersService,
     private volumeService: VolumeService,
     private toast: ToastrService,
     private translate: TranslateService,
     private orgService: OrganizationService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  orgId: string;
  clusterList: Cluster[];
  allSecrets: Secret[] = [];
  selectedSecret: Secret;

  namespaceList: Namespace[] = [];
  showNoNonDefaultNamespaceAlert = false;

  showSearch = true;
  searchName: string;
  loading = false;

  createSecretOpen = false;
  editDockerConfigJsonSecretOpen = false;
  editDockerConfigSecretOpen = false;
  editBasicAuthSecretOpen = false;
  editSSHAuthSecretOpen = false;
  editTLSSecretOpen = false;
  editRBDSecretOpen = false;
  editOpaqueSecretOpen = false;
  removeDialog = false;

  showAddFile = false;
  showAddEditFile = false;
  filename: string;
  editFilename: string;

  @ViewChild("fileUpload", {static: false}) fileUpload: ElementRef;
  @ViewChild("fileEditUpload", {static: false}) fileEditUpload: ElementRef;
  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardsecret') wizardnewsecret: ClrWizard;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  basicSecretForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(40), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    cluster: new FormControl('', Validators.required),
    namespace: new FormControl('', Validators.required),
    type: new FormControl('Opaque', Validators.required)
  });

  genericSecretDataList = new FormArray([]);
  genericSecretDataForm = new FormGroup({
    data: this.genericSecretDataList,
    type: new FormControl('Opaque')
  });

  readonly packagePattern = CustomValidatorDirective.packagePattern;
  registryAuth = new FormArray([]);
  dockerConfigJSONSecretForm = new FormGroup({
    auth: this.registryAuth
  });
  dockerConfigSecretForm = new FormGroup({
    config: new FormControl('', Validators.required)
  });

  basicAuthSecretForm = new FormGroup({
    authUser: new FormControl('', Validators.required),
    authPassword: new FormControl('', Validators.required)
  });

  sshAuthSecretForm = new FormGroup({
    privateKey: new FormControl('', [Validators.required, KEYFormatValidator()])
  });

  TLSSecretForm = new FormGroup({
    key: new FormControl('', [Validators.required, KEYFormatValidator()]),
    cert: new FormControl('', [Validators.required, CAFormatValidator()])
  });

  rbdSecretForm = new FormGroup({
    key: new FormControl('', Validators.required)
  });

  editGenericSecretDataList = new FormArray([]);
  editGenericSecretDataForm = new FormGroup({
    data: this.editGenericSecretDataList
  });
  editDockerConfigJSONSecretForm = new FormGroup({
    auth: this.registryAuth
  });

  editDockerConfigSecretForm = new FormGroup({
    config: new FormControl('', Validators.required)
  });

  editBasicAuthSecretForm = new FormGroup({
    authUser: new FormControl('', Validators.required),
    authPassword: new FormControl('', Validators.required)
  });

  editSshAuthSecretForm = new FormGroup({
    privateKey: new FormControl('', [Validators.required, KEYFormatValidator()])
  });

  editTLSSecretForm = new FormGroup({
    key: new FormControl('', [Validators.required, KEYFormatValidator()]),
    cert: new FormControl('', [Validators.required, CAFormatValidator()])
  });

  editRBDSecretForm = new FormGroup({
    key: new FormControl('', Validators.required)
  });

  deleteSecretForm = new FormGroup({
    confirm: new FormControl('', [Validators.required]),
  });

  ngOnInit(): void {
  }

  refresh(): void {
    this.loading = true;
    const filters = [{ key: 'organization', value: [this.orgId] }];
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.volumeService.getAllSecrets(filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.SECRET.GET-SECRET-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.allSecrets = !!data.data?.Secrets ? data.data.Secrets : [];
      this.selectedSecret = null;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SECRET.GET-SECRET-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false);
  }

  getAllHealthyClusterIds(): void {
    this.clusterList = null;
    this.clusterService.getAllHealthyClusterId(this.orgId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.clusterList = !!data.data?.Clusters?.edge?.nodes ? data.data.Clusters.edge.nodes : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
  }

  getAllNamespaces($event: any): void {
    const clusterId = $event.target.value;
    this.showNoNonDefaultNamespaceAlert = false;
    this.namespaceList = null;
    this.clusterService.getClusterNamespaces(clusterId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
     this.namespaceList = !!data.data?.Cluster?.namespaces ? namespacediffset(data.data.Cluster.namespaces) : [];
     this.showNoNonDefaultNamespaceAlert = this.namespaceList.length <= 0
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  checkCanEditable(labels: JSON): boolean {
    //check ingress with label 'map.tanzu.vmware.com/component'
    if(!!labels){
      for (const key of Object.keys(labels)) {
        if(key === 'map.tanzu.vmware.com/component'){
          return false;
        }
      }
    }
    return true;
  }


  showCreateSecret(): void {
    this.getAllHealthyClusterIds();
    this.createSecretOpen = true;
  }

  deleteGenericSecretData(index: number): void {
    this.genericSecretDataList.removeAt(index);
  }

  addGenericSecretData(): void {
    this.genericSecretDataList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required),
      nonFile: new FormControl(true)
    }));
  }

  deleteAuth(index: number): void {
    this.registryAuth.removeAt(index);
  }

  addAuth(): void {
    this.registryAuth.push(new FormGroup({
      endpoint: new FormControl('', [Validators.required, Validators.pattern(this.packagePattern)]),
      authUser: new FormControl('', Validators.required),
      authPassword: new FormControl('', Validators.required),
      email: new FormControl('', Validators.email)
    }));
  }

  showAddFileData(): void {
    this.showAddFile = true;
  }

  addFileToData(): void {
    const fileUpload = this.fileUpload.nativeElement;
    const file = fileUpload.files[0];
    if (file) {
      var reader = new FileReader();
      reader.readAsText(file, "UTF-8");
      reader.onload = (e) => {
        const fileContext = reader.result;
        this.genericSecretDataList.push(new FormGroup({
          key: new FormControl(file.name, Validators.required),
          value: new FormControl(fileContext, Validators.required),
          nonFile: new FormControl(false)
        }));
      };
    }
    this.fileUpload.nativeElement.value = '';
    this.filename = null;
    this.showAddFile = false;
  }

  cancelAddFile(): void {
    this.fileUpload.nativeElement.value = null;
    this.filename = null;
    this.showAddFile = false;
  }

  getFileName(): void {
    const fileUpload = this.fileUpload.nativeElement;
    const file = fileUpload.files[0];
    this.filename = file ? file.name : '';
  }

  initBySecretType($event: any): void {
    const secretType = $event.target.value;
    switch (secretType) {
      case 'kubernetes.io/dockerconfigjson': {
        this.addAuth();
        break;
      }
      default: {
        while (this.registryAuth.length > 0){
          this.registryAuth.removeAt(0)
        }
        break;
      }
    }
  }

  doCreateSecretFinish(): void {
    const secretParam: CreateSecret = {
      name: this.basicSecretForm.value.name,
      namespace: this.basicSecretForm.value.namespace,
      type: this.basicSecretForm.value.type,
      cluster: this.basicSecretForm.value.cluster
    };
    switch (this.basicSecretForm.value.type) {
      case 'kubernetes.io/dockerconfigjson':
        let authList = [];
        for(const each of this.registryAuth.controls){
          //base64 auth
          const auth = btoa(each.value.authUser + ':' + each.value.authPassword);
           authList.push('"' + each.value.endpoint + '":{"username":"' + each.value.authUser + '","password":"' + each.value.authPassword + '","email":"'+ (each.value.email ? each.value.email : '') + '","auth":"' + auth + '"}');
        }
        const dockerConfigJson = '{"auths":{' + authList.join(',') + '}}'
        const dockerConfigJsonDataMap: JSON = JSON.parse(JSON.stringify({}));
        dockerConfigJsonDataMap['.dockerconfigjson'] = dockerConfigJson;
        secretParam.data = dockerConfigJsonDataMap;
        break;
      case 'kubernetes.io/dockercfg':
        const dockerConfigDataMap: JSON = JSON.parse(JSON.stringify({}));
        dockerConfigDataMap['.dockercfg'] = this.dockerConfigSecretForm.value.config;
        secretParam.data = dockerConfigDataMap
        break;
      case 'kubernetes.io/basic-auth':
        const basicAuthDataMap: JSON = JSON.parse(JSON.stringify({}));
        basicAuthDataMap['username'] = this.basicAuthSecretForm.value.authUser;
        basicAuthDataMap['password'] = this.basicAuthSecretForm.value.authPassword;
        secretParam.data = basicAuthDataMap;
        break;
      case 'kubernetes.io/ssh-auth':
        const SSHAuthDataMap: JSON = JSON.parse(JSON.stringify({}));
        SSHAuthDataMap['ssh-privatekey'] = this.sshAuthSecretForm.value.privateKey;
        secretParam.data = SSHAuthDataMap;
        break;
      case 'kubernetes.io/tls':
        const TLSDataMap: JSON = JSON.parse(JSON.stringify({}));
        TLSDataMap['tls.crt'] = this.TLSSecretForm.value.cert;
        TLSDataMap['tls.key'] = this.TLSSecretForm.value.key;
        secretParam.data = TLSDataMap;
        break;
      case 'kubernetes.io/rbd':
        const rbdDataMap: JSON = JSON.parse(JSON.stringify({}));
        rbdDataMap['key'] = this.rbdSecretForm.value.key;
        secretParam.data = rbdDataMap;
        break;
      default:
        const dataMap: JSON = JSON.parse(JSON.stringify({}));
        if (!!this.genericSecretDataForm.value.data) {
          for (const each of this.genericSecretDataForm.value.data) {
            dataMap[each.key] = each.value;
          }
        }
        secretParam.data = dataMap;
        break;
    }
    this.volumeService.createSecret(secretParam).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.SECRET.CREATE-SECRET-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.CreateSecret){
        this.toast.show(`${data.data.CreateSecret.name} `+this.translate.instant('MESSAGE.VOLUME.SECRET.CREATE-SECRET-MSG'), this.translate.instant('MESSAGE.VOLUME.SECRET.CREATE-SECRET-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SECRET.CREATE-SECRET-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doCreateSecretReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doCreateSecretCancel(): void {
    this.doCreateSecretReset()
  }

  private doCreateSecretReset(): void {
    this.wizardnewsecret.reset();
    this.basicSecretForm.reset();
    this.basicSecretForm.patchValue({
      type: 'Opaque'
    })
    this.genericSecretDataForm.reset();
    while (this.genericSecretDataList.length !== 0) {
      this.genericSecretDataList.removeAt(0);
    }
    this.dockerConfigJSONSecretForm.reset();
    while (this.registryAuth.length > 0){
      this.registryAuth.removeAt(0);
    }
    this.dockerConfigSecretForm.reset();
    this.basicAuthSecretForm.reset();
    this.sshAuthSecretForm.reset();
    this.TLSSecretForm.reset();
    this.rbdSecretForm.reset();
    this.createSecretOpen = false;
    this.clusterList = [];
    this.namespaceList = [];
  }

  showUpdateSecret(secret: Secret): void {
    this.selectedSecret = secret;
    switch (secret.type) {
      case 'kubernetes.io/dockerconfigjson':
        const dockerConfigJsonData = JSON.parse(secret.data['.dockerconfigjson']);
        let email = '';
        let user = '';
        // check dockerConfigJsonData['auths'] is {}
        if(Object.keys(dockerConfigJsonData['auths']).length > 0){
          for(const ep of Object.keys(dockerConfigJsonData['auths'])){
            if(dockerConfigJsonData['auths'][ep].hasOwnProperty('email')){
              email = dockerConfigJsonData['auths'][ep]['email'];
            }
            if(dockerConfigJsonData['auths'][ep].hasOwnProperty('username')){
              user = dockerConfigJsonData['auths'][ep]['username'];
            }
            this.registryAuth.push(new FormGroup({
              endpoint: new FormControl(ep, [Validators.required, Validators.pattern(this.packagePattern)]),
              authUser: new FormControl(user, Validators.required),
              authPassword: new FormControl('', Validators.required),
              email: new FormControl(email, Validators.email)
            }))
          }
        }
        this.editDockerConfigJsonSecretOpen = true;
        break;
      case 'kubernetes.io/dockercfg':
        this.editDockerConfigSecretForm.setValue({
          config: ''
        });
        this.editDockerConfigSecretOpen = true;
        break;
      case 'kubernetes.io/basic-auth':
        this.editBasicAuthSecretForm.setValue({
          authUser: secret.data['username'],
          authPassword: ''
        });
        this.editBasicAuthSecretOpen = true;
        break;
      case 'kubernetes.io/ssh-auth':
        this.editSshAuthSecretForm.setValue({
          privateKey: ''
        });
        this.editSSHAuthSecretOpen = true;
        break;
      case 'kubernetes.io/tls':
        this.editTLSSecretForm.setValue({
          cert: secret.data['tls.crt'],
          key: ''
        });
        this.editTLSSecretOpen = true;
        break;
      case 'kubernetes.io/rbd':
        this.editRBDSecretForm.setValue({
          key: ''
        });
        this.editRBDSecretOpen = true;
        break;
      default:
        if (!!secret.data) {
          const datas = secret.data;
          for (const key of Object.keys(datas)) {
            const value = datas[key];
            var match = /\r|\n/.exec(value);
            if (!match){
              this.editGenericSecretDataList.push(new FormGroup({
                key: new FormControl(key),
                value: new FormControl(value),
                nonFile: new FormControl(true)
              }));
            } else {
              this.editGenericSecretDataList.push(new FormGroup({
                key: new FormControl(key),
                value: new FormControl(value),
                nonFile: new FormControl(false)
              }));
            }
          }
        }
        this.editOpaqueSecretOpen = true;
        break;
    }
  }

  deleteEditGenericSecretData(index: number): void {
    this.editGenericSecretDataList.removeAt(index);
  }

  addEditGenericSecretData(): void {
    this.editGenericSecretDataList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required),
      nonFile: new FormControl(true)
    }));
  }

  showAddEditFileData(): void {
    this.showAddEditFile = true;
  }

  addEditFileToData(): void {
    const fileUpload = this.fileEditUpload.nativeElement;
    const file = fileUpload.files[0];
    if (file) {
      var reader = new FileReader();
      reader.readAsText(file, "UTF-8");
      reader.onload = (e) => {
        const fileContext = reader.result;
        this.editGenericSecretDataList.push(new FormGroup({
          key: new FormControl(file.name, Validators.required),
          value: new FormControl(fileContext, Validators.required),
          nonFile: new FormControl(false)
        }));
      };
    }
    this.fileEditUpload.nativeElement.value = '';
    this.editFilename = null;
    this.showAddEditFile = false;
  }

  cancelAddEditFile(): void {
    this.fileEditUpload.nativeElement.value = null;
    this.editFilename = null;
    this.showAddEditFile = false;
  }

  getEditFileName(): void {
    const fileUpload = this.fileEditUpload.nativeElement;
    const file = fileUpload.files[0];
    this.editFilename = file ? file.name : '';
  }

  doEditSecretFinish(): void {
    const secretParam: UpdateSecret = {
      name: this.selectedSecret.name,
      namespace: this.selectedSecret.namespace,
      cluster: this.selectedSecret.cluster.id
    };
    const type = this.selectedSecret.type;
    switch (type) {
      case 'kubernetes.io/dockerconfigjson':
        let authList = [];
        for(const each of this.registryAuth.controls){
          //base64 auth
          const auth = btoa(each.value.authUser + ':' + each.value.authPassword);
           authList.push('"' + each.value.endpoint + '":{"username":"' + each.value.authUser + '","password":"' + each.value.authPassword + '","email":"'+ (each.value.email ? each.value.email : '') + '","auth":"' + auth + '"}');
        }
        const dockerConfigJson = '{"auths":{' + authList.join(',') + '}}'
        const dockerConfigJsonDataMap: JSON = JSON.parse(JSON.stringify({}));
        dockerConfigJsonDataMap['.dockerconfigjson'] = dockerConfigJson;
        secretParam.data = dockerConfigJsonDataMap;
        break;
      case 'kubernetes.io/dockercfg':
        const dockerConfigDataMap: JSON = JSON.parse(JSON.stringify({}));
        dockerConfigDataMap['.dockercfg'] = this.editDockerConfigSecretForm.value.config;
        secretParam.data = dockerConfigDataMap
        break;
      case 'kubernetes.io/basic-auth':
        const basicAuthDataMap: JSON = JSON.parse(JSON.stringify({}));
        basicAuthDataMap['username'] = this.editBasicAuthSecretForm.value.authUser;
        basicAuthDataMap['password'] = this.editBasicAuthSecretForm.value.authPassword;
        secretParam.data = basicAuthDataMap;
        break;
      case 'kubernetes.io/ssh-auth':
        const SSHAuthDataMap: JSON = JSON.parse(JSON.stringify({}));
        SSHAuthDataMap['ssh-privatekey'] = this.editSshAuthSecretForm.value.privateKey;
        secretParam.data = SSHAuthDataMap;
        break;
      case 'kubernetes.io/tls':
        const TLSDataMap: JSON = JSON.parse(JSON.stringify({}));
        TLSDataMap['tls.crt'] = this.editTLSSecretForm.value.cert;
        TLSDataMap['tls.key'] = this.editTLSSecretForm.value.key;
        secretParam.data = TLSDataMap;
        break;
      case 'kubernetes.io/rbd':
        const rbdDataMap: JSON = JSON.parse(JSON.stringify({}));
        rbdDataMap['key'] = this.editRBDSecretForm.value.key;
        secretParam.data = rbdDataMap;
        break;
      default:
        const dataMap: JSON = JSON.parse(JSON.stringify({}));
        if (!!this.editGenericSecretDataForm.value.data) {
          for (const each of this.editGenericSecretDataForm.value.data) {
            dataMap[each.key] = each.value;
          }
        }
        secretParam.data = dataMap;
        break;
    }
    this.volumeService.updateSecret(secretParam).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.SECRET.UPDATE-SECRET-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.UpdateSecret){
        this.toast.show(`${data.data.UpdateSecret.name} `+this.translate.instant('MESSAGE.VOLUME.SECRET.UPDATE-SECRET-MSG'), this.translate.instant('MESSAGE.VOLUME.SECRET.UPDATE-SECRET-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SECRET.UPDATE-SECRET-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doEditSecretReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doEditSecretCancel(): void {
    this.doEditSecretReset();
  }

  doEditSecretReset(): void {
    this.editGenericSecretDataForm.reset();
    while (this.editGenericSecretDataList.length !== 0) {
      this.editGenericSecretDataList.removeAt(0);
    }
    this.editDockerConfigJSONSecretForm.reset();
    while (this.registryAuth.length > 0){
      this.registryAuth.removeAt(0);
    }
    this.editDockerConfigSecretForm.reset();
    this.editBasicAuthSecretForm.reset();
    this.editSshAuthSecretForm.reset();
    this.editTLSSecretForm.reset();
    this.editRBDSecretForm.reset();
    this.editDockerConfigJsonSecretOpen = false;
    this.editDockerConfigSecretOpen = false;
    this.editBasicAuthSecretOpen = false;
    this.editSSHAuthSecretOpen = false;
    this.editTLSSecretOpen = false;
    this.editRBDSecretOpen = false;
    this.editOpaqueSecretOpen = false;
  }

  showRemoveDialog(secret: Secret): void {
    this.selectedSecret = secret;
    this.removeDialog = true;
  }

  deleteSecretByName(): void {
    this.removeDialog = false;
    this.volumeService.deleteSecret(this.selectedSecret?.name, this.selectedSecret?.cluster?.id, this.selectedSecret?.namespace).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.SECRET.DELETE-SECRET-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.DeleteSecret){
          this.toast.show(`${data.data.DeleteSecret.name} `+this.translate.instant('MESSAGE.VOLUME.SECRET.DELETE-SECRET-MSG'), this.translate.instant('MESSAGE.VOLUME.SECRET.DELETE-SECRET-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SECRET.DELETE-SECRET-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

}
