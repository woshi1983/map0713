import {Component, OnInit, ViewChild} from '@angular/core';
import {
  Cluster,
  handleObject,
  Ingress,
  Namespace,
  Secret,
  Service,
  CreateIngress,
  IngressRuleInput,
  HTTPIngressPathInput,
  IngressTLSInput,
  DeleteIngress,
  IngressClass,
  IngressTLS, UpdateIngress
} from "../../../../classes";
import {AuthService} from "../../../../services/auth.service";
import {ClrDatagridPagination, ClrWizard} from "@clr/angular";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {IngressService} from "../../../../services/ingress.service";
import {OrganizationService} from "../../../../services/organization.service";
import {
  CustomValidatorDirective,
  nameValidator
} from "../../../../utils/validators/custom-validator.directive";
import {FormArray, FormBuilder, FormControl, FormGroup, Validators} from "@angular/forms";
import {namespacediffset} from "../../../../utils/operator";
import {ClustersService} from "../../../../services/clusters.service";
import {ServiceService} from "../../../../services/service.service";
import {VolumeService} from "../../../../services/volume.service";

@Component({
  selector: 'app-ingress',
  templateUrl: './ingress.component.html',
  styleUrls: ['./ingress.component.css']
})
export class IngressComponent implements OnInit {

  //admin and user role can handle cluster resources
  get canWriteResources(): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER_RESOURCE);
  }

  orgId: string;
  ingresses: Ingress[] = [];
  selectedIngress: Ingress;
  selectedIngressName: string;
  clusterList: Cluster[] = [];
  namespaceList: Namespace[] = [];
  ingressClassList: IngressClass[] = [];
  secretList: Secret[] = [];
  serviceList: Service[] = [];

  createIngressOpen: boolean;
  showNoNonDefaultNamespaceAlert = false;
  showNoIngressClassAlert = false;
  showNoTLSSecretAlert = false;
  showNoServiceAlert = false;
  updateIngressOpen: boolean;
  showUpdateNoIngressClass = false;
  removeDialog = false;

  showSearch = true;
  searchName: string;
  loading = false;

  readonly disAllowLabels = CustomValidatorDirective.disAllowLabels;
  readonly disAllowAnnotations = CustomValidatorDirective.disAllowAnnotations;
  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardingress') wizardnewingress: ClrWizard;
  readonly hostsPattern = CustomValidatorDirective.hostsPattern;
  readonly commonNamePattern = CustomValidatorDirective.commonNamePattern;
  labelsList = new FormArray([]);
  annotationsList = new FormArray([]);
  basicIngressForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.commonNamePattern), nameValidator([])]),
    cluster: new FormControl('', Validators.required),
    namespace: new FormControl('', Validators.required),
    ingressClass: new FormControl('', Validators.required),
    labels: this.labelsList,
    annotations: this.annotationsList
  });
  ruleList = new FormArray([]);
  allPathList = [];
  addRuleForm = new FormGroup({
    rules: this.ruleList,
  });
  @ViewChild('wizardeditingress') wizardeditingress: ClrWizard;
  editLabelsList = new FormArray([]);
  editAnnotationsList = new FormArray([]);
  editIngressForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.commonNamePattern), nameValidator([])]),
    cluster: new FormControl('', Validators.required),
    namespace: new FormControl('', Validators.required),
    ingressClass: new FormControl('', Validators.required),
    labels: this.editLabelsList,
    annotations: this.editAnnotationsList
  });
  editRuleList = new FormArray([]);
  editAllPathList = [];
  editAddRuleForm = new FormGroup({
    rules: this.editRuleList,
  });
  deleteIngressForm = new FormGroup({
    confirm: new FormControl('', [Validators.required]),
  });

  constructor(
    public authService: AuthService,
    private ingressService: IngressService,
    private toast: ToastrService,
    private translate: TranslateService,
    private fb: FormBuilder,
    private orgService: OrganizationService,
    private clusterService: ClustersService,
    private serviceService: ServiceService,
    private volumeService: VolumeService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  ngOnInit(): void {
  }

  refresh(): void {
    this.loading = true;
    const filters = [{ key: 'organization', value: [this.orgId] }];
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.ingressService.getIngresses(filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.GET-INGRESS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.ingresses = !!data.data?.Ingresses ?  data.data.Ingresses : [];
      this.selectedIngress = null;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.GET-INGRESS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false);
  }

  getAllHealthyClusterIds(): void {
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

  getAllIngressClass(clusterId: string, update?: boolean, ingresClass?: string): void {
     this.showNoIngressClassAlert = false;
     this.ingressClassList = [];
     this.showUpdateNoIngressClass = false;
     this.ingressService.getAllIngressClass(clusterId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.GET-INGCLASS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.ingressClassList = !!data.data?.IngressClasses ? data.data.IngressClasses : [];
      if(!!update){
        this.showUpdateNoIngressClass = !this.ingressClassList.find(x => x.name === ingresClass);
      } else {
        this.showNoIngressClassAlert = this.ingressClassList.length <= 0;
      }
     }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.GET-INGCLASS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
  }

  refreshNsInCluster($event: any): void {
    const clusterId = $event.target.value;
    this.showNoNonDefaultNamespaceAlert = false;
    this.namespaceList = [];
    this.clusterService.getClusterNamespaces(clusterId).subscribe(data => {
      // get namespace in cluster
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
     this.namespaceList = !!data.data?.Cluster?.namespaces ? namespacediffset(data.data.Cluster.namespaces) : [];
     this.showNoNonDefaultNamespaceAlert = this.namespaceList.length <= 0;
     this.getAllIngressClass(clusterId);
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  refreshServicesInNamespace($event: any): void {
    const ns = $event.target.value;
    const cluster = this.basicIngressForm.value.cluster;
    this.serviceList = [];
    this.showNoServiceAlert = false;
    this.serviceService.getApplicationService(cluster, ns).subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.GET-SERVICE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.serviceList = !!data.data?.Services ? data.data?.Services : [];
      this.showNoServiceAlert = this.serviceList.length <= 0;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.GET-SERVICE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
  }

  getAllServicesInNamespace(cluster: string, ns: string): void {
    this.serviceList = [];
    this.serviceService.getApplicationService(cluster, ns).subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.GET-SERVICE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.serviceList = !!data.data?.Services ? data.data?.Services : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.GET-SERVICE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
  }

  getPortInService(service: string): any[] {
    for(let svc of this.serviceList){
      if(svc.name === service){
        return svc.spec.ports;
      }
    }
    return [];
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

  showCreateIngress(): void {
    this.getAllHealthyClusterIds();
    this.addRule();
    this.createIngressOpen = true;
  }

  checkTLSHostExist(host: string, tlsHostList: IngressTLS[]): boolean {
    //check tls null
    if(!tlsHostList){
      return false
    }
    let hostsList: string[] = [];
    for(let each of tlsHostList){
      hostsList.push(each.hosts[0])
    }
    return !!hostsList.find( x => x === host)
  }

  addNewIngressLabel(): void {
    this.labelsList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  deleteIngressLable(index: number): void{
    this.labelsList.removeAt(index);
  }

  addNewIngressAnnotation(): void {
    this.annotationsList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('')
    }));
  }

  deleteIngressAnnotation(index: number): void{
    this.annotationsList.removeAt(index);
  }

  deleteRule(ruleIndex: number): void {
    this.ruleList.removeAt(ruleIndex);
    this.allPathList.splice(ruleIndex, 1);
  }

  addRule(): void {
    this.allPathList[this.ruleList.length] = new FormArray([]);
    this.ruleList.push(this.fb.group({
      host: new FormControl('', [Validators.required, Validators.pattern(this.hostsPattern)]),
      enableHttps: new FormControl(false),
      secretName: new FormControl(''),
      paths: this.allPathList[this.ruleList.length]
    }));
    this.addNewPath(this.ruleList.length - 1);
  }

  addNewPath(ruleIndex: number): void {
    this.allPathList[ruleIndex].push(this.fb.group({
      pathType: new FormControl('Prefix'),
      path: new FormControl('/'),
      service: new FormControl('', Validators.required),
      port: new FormControl('', Validators.required)
    }));
  }

  deletePath(ruleIndex: number, pathIndex: number): void {
    this.allPathList[ruleIndex].removeAt(pathIndex);
  }

  setHttpsSecretVaildtor(index: number, enabled: boolean){
    if(enabled){
      this.volumeService.getTLSSecretsByNamespace(this.basicIngressForm.value.cluster, this.basicIngressForm.value.namespace).subscribe( data => {
        if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.SECRET.GET-SECRET-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
        }
        const tempSecretList = !!data.data?.Secrets ? data.data.Secrets : [];
        for(let secret of tempSecretList){
          if(secret.__typename === 'kubernetes.io/tls' && !this.secretList.find(x => x.name === secret.name)){
            this.secretList.push(secret)
          }
        }
        this.showNoTLSSecretAlert = this.secretList.length <= 0
      }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SECRET.GET-SECRET-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
      this.ruleList.at(index).get('secretName').setValidators(Validators.required);
      this.ruleList.at(index).get('secretName').updateValueAndValidity();
    } else {
      this.ruleList.at(index).get('secretName').clearValidators();
      this.ruleList.at(index).get('secretName').updateValueAndValidity();
    }
  }

  doIngressFinish(): void {
    const labelsMap: JSON = JSON.parse(JSON.stringify({}));
    if (!!this.basicIngressForm.value.labels) {
      for (const each of this.basicIngressForm.value.labels) {
        labelsMap[each.key] = each.value;
      }
    }
    const annotationMap: JSON = JSON.parse(JSON.stringify({}));
    if (!!this.basicIngressForm.value.annotations) {
      for (const each of this.basicIngressForm.value.annotations) {
        annotationMap[each.key] = each.value;
      }
    }
    const rules: IngressRuleInput[] = [];
    const tls: IngressTLSInput[] = [];
    for(var i=0;i < this.ruleList.length; i++){
      // handle rules
      const rule = this.ruleList.controls[i];
      const paths: HTTPIngressPathInput[] = [];
      for(let path of this.allPathList[i].controls){
        paths.push({
          path: path.value.path,
          pathType: path.value.pathType,
          backend: {
            service: {
              name: path.value.service,
              port: {
                number: path.value.port
              }
            }
          }
        })
      }
      rules.push({
        host: rule.value.host,
        http: {
          paths: paths
        }
      })
      // handle tls
      if(rule.value.enableHttps){
        tls.push({
          hosts: [rule.value.host],
          secret: rule.value.secretName
        })
      }
    }
    const param: CreateIngress = {
      name: this.basicIngressForm.value.name,
      cluster: this.basicIngressForm.value.cluster,
      namespace: this.basicIngressForm.value.namespace,
      labels: labelsMap,
      annotations: annotationMap,
      spec: {
        ingressClass: this.basicIngressForm.value.ingressClass,
        tls: tls,
        rules: rules
      }
    }
    this.ingressService.createIngress(param).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.CREATE-INGRESS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.CreateIngress){
        this.toast.show(`${data.data.CreateIngress.name} `+this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.CREATE-INGRESS-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.CREATE-INGRESS-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.CREATE-INGRESS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doIngressReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    })
  }

  doIngressCancel(): void {
    this.doIngressReset();
  }

  doIngressReset(): void {
    this.createIngressOpen = false;
    this.wizardnewingress.reset();
    this.basicIngressForm.reset();
    while(this.labelsList.length !== 0) {
      this.labelsList.removeAt(0);
    }
    while(this.annotationsList.length !== 0) {
      this.annotationsList.removeAt(0);
    }
    this.addRuleForm.reset();
    this.allPathList = [];
    while(this.ruleList.length !== 0) {
      this.ruleList.removeAt(0);
    }
    this.clusterList = [];
    this.serviceList = [];
    this.secretList = [];
    this.namespaceList = [];
    this.ingressClassList = [];
  }

  showUpdateIngress(ingress: Ingress): void {
    this.clusterList[0]={id: ingress.cluster.id, name: ingress.cluster.name};
    this.namespaceList[0]={name: ingress.namespace};
    this.getAllIngressClass(ingress.cluster.id, true, ingress.spec.ingressClass);
    for (const key of Object.keys(ingress.labels)) {
      if(!this.disAllowLabels.find(x => x === key)){
        this.editLabelsList.push(new FormGroup({
            key: new FormControl(key),
            value: new FormControl(ingress.labels[key])
          }
        ))
      }
    }
    if(!!ingress.annotations){
      for (const key of Object.keys(ingress.annotations)) {
        if(!this.disAllowAnnotations.find(x => x === key)){
          this.editAnnotationsList.push(new FormGroup({
              key: new FormControl(key),
              value: new FormControl(ingress.annotations[key])
            }
          ))
        }
      }
    }
    this.getAllServicesInNamespace(ingress.cluster.id, ingress.namespace);
    this.editIngressForm.patchValue({
      name: ingress.name,
      cluster: ingress.cluster.id,
      namespace: ingress.namespace,
      ingressClass: ingress.spec.ingressClass
    })
    for(const rule of ingress.spec.rules){
      this.editAllPathList[this.editRuleList.length] = new FormArray([]);
      for(const path of rule.http.paths){
        const pathControl = this.fb.group({
          pathType: new FormControl(path.pathType),
          path: new FormControl(path.path),
          service: new FormControl(path.backend.service.name, Validators.required),
          port: new FormControl(path.backend.service.port.number, Validators.required)
        })
        this.editAllPathList[this.editRuleList.length].push(pathControl)
      }
      const ruleControl = this.fb.group({
        host: new FormControl(rule.host, [Validators.required, Validators.pattern(this.hostsPattern)]),
        enableHttps: new FormControl(this.checkTLSHostExist(rule.host, ingress.spec.tls)),
        secretName: new FormControl(this.checkTLSHostExist(rule.host, ingress.spec.tls) ? this.getTLSSecretName(rule.host, ingress.spec.tls, ingress.cluster.id, ingress.namespace) : ''),
        paths: this.editAllPathList[this.editRuleList.length]
      })
      this.editRuleList.push(ruleControl)
    }
    this.updateIngressOpen = true;
  }

  getTLSSecretName(host: string, tlsHostList: IngressTLS[], cluster: string, ns: string): string {
    this.volumeService.getTLSSecretsByNamespace(cluster, ns).subscribe( data => {
        if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.SECRET.GET-SECRET-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
        }
        const tempSecretList = !!data.data?.Secrets ? data.data.Secrets : [];
        for(let secret of tempSecretList){
          if(secret.__typename === 'kubernetes.io/tls' && !this.secretList.find(x => x.name === secret.name)){
            this.secretList.push(secret)
          }
        }
      }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SECRET.GET-SECRET-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
    //check tls null
    if(!tlsHostList){
      return ''
    }
    for(let each of tlsHostList){
      if(host === each.hosts[0]){
        return each.secret
      }
    }
  }

  checkSecretExist(secret: string): boolean {
    if(!!this.secretList && !!secret){
      return !this.secretList.find(x => x.name === secret);
    }
    return false;
  }

  checkServiceExist(service: string): boolean {
    if(!!this.serviceList && !!service){
      return !this.serviceList.find(x => x.name === service);
    }
    return false;
  }

  addEditNewIngressLabel(): void {
    this.editLabelsList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  deleteEditIngressLable(index: number): void{
    this.editLabelsList.removeAt(index);
  }

  addEditNewIngressAnnotation(): void {
    this.editAnnotationsList.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('')
    }));
  }

  deleteEditIngressAnnotation(index: number): void{
    this.editAnnotationsList.removeAt(index);
  }

  deleteEditRule(ruleIndex: number): void {
    this.editRuleList.removeAt(ruleIndex);
    this.editAllPathList.splice(ruleIndex, 1);
  }

  addEditRule(): void {
    this.editAllPathList[this.editRuleList.length] = new FormArray([]);
    this.editRuleList.push(this.fb.group({
      host: new FormControl('', [Validators.required, Validators.pattern(this.hostsPattern)]),
      enableHttps: new FormControl(false),
      secretName: new FormControl(''),
      paths: this.editAllPathList[this.editRuleList.length]
    }));
    this.addEditNewPath(this.editRuleList.length - 1);
  }

  addEditNewPath(ruleIndex: number): void {
    this.editAllPathList[ruleIndex].push(this.fb.group({
      pathType: new FormControl('Prefix'),
      path: new FormControl('/'),
      service: new FormControl('', Validators.required),
      port: new FormControl('', Validators.required)
    }));
  }

  deleteEditPath(ruleIndex: number, pathIndex: number): void {
    this.editAllPathList[ruleIndex].removeAt(pathIndex);
  }

  setEditHttpsSecretVaildtor(index: number, enabled: boolean){
    if(enabled){
      this.volumeService.getTLSSecretsByNamespace(this.editIngressForm.value.cluster, this.editIngressForm.value.namespace).subscribe( data => {
        if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.VOLUME.SECRET.GET-SECRET-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
        }
        const tempSecretList = !!data.data?.Secrets ? data.data.Secrets : [];
        for(let secret of tempSecretList){
          if(secret.__typename === 'kubernetes.io/tls' && !this.secretList.find(x => x.name === secret.name)){
            this.secretList.push(secret)
          }
        }
        this.showNoTLSSecretAlert = this.secretList.length <= 0
      }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.VOLUME.SECRET.GET-SECRET-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
      this.editRuleList.at(index).get('secretName').setValidators(Validators.required);
      this.editRuleList.at(index).get('secretName').updateValueAndValidity();
    } else {
      this.editRuleList.at(index).get('secretName').clearValidators();
      this.editRuleList.at(index).get('secretName').updateValueAndValidity();
    }
  }

  doUpdateIngressFinish(): void {
    const labelsMap: JSON = JSON.parse(JSON.stringify({}));
    if (!!this.editIngressForm.value.labels) {
      for (const each of this.editIngressForm.value.labels) {
        labelsMap[each.key] = each.value;
      }
    }
    const annotationMap: JSON = JSON.parse(JSON.stringify({}));
    if (!!this.editIngressForm.value.annotations) {
      for (const each of this.editIngressForm.value.annotations) {
        annotationMap[each.key] = each.value;
      }
    }
    const rules: IngressRuleInput[] = [];
    const tls: IngressTLSInput[] = [];
    for(var i=0;i < this.editRuleList.length; i++){
      // handle rules
      const rule = this.editRuleList.controls[i];
      const paths: HTTPIngressPathInput[] = [];
      for(let path of this.editAllPathList[i].controls){
        paths.push({
          path: path.value.path,
          pathType: path.value.pathType,
          backend: {
            service: {
              name: path.value.service,
              port: {
                number: path.value.port
              }
            }
          }
        })
      }
      rules.push({
        host: rule.value.host,
        http: {
          paths: paths
        }
      })
      // handle tls
      if(rule.value.enableHttps){
        tls.push({
          hosts: [rule.value.host],
          secret: rule.value.secretName
        })
      }
    }
    const param: UpdateIngress = {
      name: this.editIngressForm.value.name,
      cluster: this.editIngressForm.value.cluster,
      namespace: this.editIngressForm.value.namespace,
      labels: labelsMap,
      annotations: annotationMap,
      spec: {
        ingressClass: this.editIngressForm.value.ingressClass,
        tls: tls,
        rules: rules
      }
    }
    this.ingressService.updateIngress(param).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.CREATE-INGRESS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.UpdateIngress){
        this.toast.show(`${data.data.UpdateIngress.name} `+this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.CREATE-INGRESS-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.CREATE-INGRESS-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.CREATE-INGRESS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doUpdateIngressReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    })
  }

  doUpdateIngressCancel(): void {
    this.doUpdateIngressReset();
  }

  doUpdateIngressReset(): void {
    this.updateIngressOpen = false;
    this.wizardeditingress.reset();
    this.editIngressForm.reset();
    while(this.editLabelsList.length !== 0) {
      this.editLabelsList.removeAt(0);
    }
    while(this.editAnnotationsList.length !== 0) {
      this.editAnnotationsList.removeAt(0);
    }
    this.editAddRuleForm.reset();
    this.editAllPathList = [];
    while(this.editRuleList.length !== 0) {
      this.editRuleList.removeAt(0);
    }
    this.clusterList = [];
    this.serviceList = [];
    this.secretList = [];
    this.namespaceList = [];
    this.ingressClassList = [];
  }

  showDeleteIngress(ingress: Ingress): void {
    this.selectedIngress = ingress;
    this.selectedIngressName = ingress.name;
    this.removeDialog = true;
  }

  deleteIngressByName(ingress: Ingress): void {
    this.removeDialog = false;
    const param: DeleteIngress = {
      name: ingress.name,
      namespace: ingress.namespace,
      cluster: ingress.cluster.id
    }
    this.ingressService.deleteIngress(param).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.DELETE-INGRESS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.DeleteIngress){
          this.toast.show(`${data.data.DeleteIngress.name} `+this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.DELETE-INGRESS-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.DELETE-INGRESS-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.INGRESS.DELETE-INGRESS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }
}
