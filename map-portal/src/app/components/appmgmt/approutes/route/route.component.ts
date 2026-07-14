import { Component, ElementRef, OnInit, ViewChild } from '@angular/core';
import { LoadbalancerService } from '../../../../services/loadbalancer.service';
import { RouteService } from '../../../../services/route.service';
import { ClrDatagridPagination, ClrDatagridStateInterface, ClrWizard } from '@clr/angular';
import { FormControl, FormGroup, Validators, FormArray, FormBuilder } from '@angular/forms';
import { AuthService } from '../../../../services/auth.service';
import {
  Route,
  CreateRoute,
  Service,
  UpdateRoute,
  Instance,
  RemoveRouteBackend,
  handleObject,
  CustomComponent,
  Application,
  LoadBalancerType,
  LBAlgorithm,
  LBAlgorithmShortName,
  AlgorithmHash,
  AlgorithmHashShortName,
  AviNetworkProfile
} from '../../../../classes';
import {CustomValidatorDirective, nameValidator} from "../../../../utils/validators/custom-validator.directive";
import { ApplicationsService } from 'app/services/applications.service';
import {ActivatedRoute} from "@angular/router";
import { ToastrService } from 'ngx-toastr';
import {TranslateService} from "@ngx-translate/core";
import {InstanceFedService} from "../../../../services/instance-fed.service";
import {OrganizationService} from "../../../../services/organization.service";
import { EChartsOption } from 'echarts';
import {ComponentService} from "../../../../services/component.service";
import {InstanceService} from "../../../../services/instance.service";

@Component({
  selector: 'app-route',
  templateUrl: './route.component.html',
  styleUrls: ['./route.component.css']
})
export class RouteComponent implements OnInit {

  //admin and user role can handle component
  get canWriteRoute(): boolean {
    return this.authService.can(this.orgId, handleObject.ROUTE);
  }

  //only superadmin can handle route force delete
  get canForceDelete(): boolean {
    return this.authService.can('NULL', handleObject.ROUTE);
  }

  orgId: string;
  lbId: string = '';
  glbRoutes: Route[] = [];
  noAVILBAlert = false;
  total = 0;
  selectedRoute: Route;
  selectedRouteId: string;

  applicationList: Application[] = [];
  componentList: CustomComponent[] = [];
  instanceList: Instance[] = [];
  servicesInInstanceM: {[id: string]: Service[]} = {};
  networkProfile: AviNetworkProfile[] = [];

  createRouteOpen = false;
  updateRouteOpen = false;
  deleteRouteOpen = false;
  forceDeleteRouteOpen = false;
  showSearch = true;
  searchName: string;

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardroute') wizardnewroute: ClrWizard;
  readonly LBAlgorithm = LBAlgorithm;
  readonly LBAlgorithmShortName = LBAlgorithmShortName;
  readonly AlgorithmHash = AlgorithmHash;
  readonly AlgorithmHashShortName = AlgorithmHashShortName;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  basicRouteForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    application: new FormControl('', Validators.required),
    component: new FormControl('', Validators.required),
    port: new FormControl('', [Validators.required, Validators.min(1), Validators.max(65535)]),
    protocol: new FormControl('TCP'),
    lbAlgorithm: new FormControl(LBAlgorithm.LB_ALGORITHM_LEAST_CONNECTIONS),
    lbAlgorithmHash: new FormControl(AlgorithmHash.LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS),
    lbAlgorithmConsistentHashHeader: new FormControl(''),
    enableVsvip: new FormControl(false),
    vsvipRef: new FormControl(''),
    persistentVsvip: new FormControl(false)
  });
  appGlbParamsArray = new FormArray([]);
  appGlbParamsRatio = new FormControl(20);
  appGlbRouteForm = new FormGroup({
    backends: this.appGlbParamsArray,
    ratio: this.appGlbParamsRatio,
    networkProfile: new FormControl(''),
  });

  @ViewChild('wizardeditroute') wizardeditroute: ClrWizard;
  @ViewChild('slider1') slider1: ElementRef<HTMLInputElement>;
  editRouteForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    application: new FormControl('', Validators.required),
    component: new FormControl('', Validators.required),
    protocol: new FormControl('TCP'),
    port: new FormControl('', [Validators.required, Validators.min(1), Validators.max(65535)]),
    lbAlgorithm: new FormControl(LBAlgorithm.LB_ALGORITHM_LEAST_CONNECTIONS),
    lbAlgorithmHash: new FormControl(AlgorithmHash.LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS),
    lbAlgorithmConsistentHashHeader: new FormControl(''),
    persistentVsvip: new FormControl(false)
  });
  editappGlbParamsArray = new FormArray([]);
  editappGlbParamsRatio = new FormControl(20);
  editappGlbRouteForm = new FormGroup({
    backends: this.editappGlbParamsArray,
    ratio: this.editappGlbParamsRatio,
    networkProfile: new FormControl(''),
  });
  oldRouteBackendIdList: RemoveRouteBackend[] = [];
  detailsModalOpen = false;
  loading = true;
  deleteRouteForm = new FormGroup({
    confirm: new FormControl('', [Validators.required]),
  });

  constructor(
    private glbRouteService: RouteService,
    private glbService: LoadbalancerService,
    private fb: FormBuilder,
    private appService: ApplicationsService,
    public authService: AuthService,
    private route: ActivatedRoute,
    private toast: ToastrService,
    private translate: TranslateService,
    private instancefedService: InstanceFedService,
    private lbService: LoadbalancerService,
    private orgService: OrganizationService,
    private cptService: ComponentService,
    private instanceService: InstanceService,
    private applicationService: ApplicationsService
    ) {
      this.orgId = this.orgService.getSelectedOrgId;
    }

  ngOnInit(): void {
    this.checkLoadBalancers();
  }

  refresh(state?: ClrDatagridStateInterface): void {
    this.loading = true;
    const filters = [{ key: 'organization', value: [this.orgId] }];
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.glbRouteService.getRoutes(!!state?.page?.current ? state.page.current - 1 : 0, !!state?.page?.size ? state.page.size : Number(this.pagination.pageSize), filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.GET-ROUTE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.glbRoutes = !!data.data?.Routes?.edge?.nodes ? data.data.Routes.edge.nodes : [];
      this.selectedRoute = null;
      this.total = !!data?.data?.Routes?.totalCount ? data.data.Routes.totalCount : 0;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.GET-ROUTE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false);
  }

  getAllApplication(): void {
    this.applicationService.getAllApplicationsByOrg(this.orgId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APPLICATION.GET-APP-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.applicationList = !!data.data?.Applications?.edge?.nodes ? data.data.Applications.edge.nodes : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APPLICATION.GET-APP-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  checkLoadBalancers(): void {
    const filters = [{ key: 'organization', value: [this.orgId] }];
    this.lbService.getAllLB( 0,  -1, filters).subscribe(data => {
      if(!!data.data?.LoadBalancers?.edge?.nodes){
        const typeList = [];
        for (const lb of data.data.LoadBalancers.edge.nodes){
          typeList.push(lb.type)
        }
        this.noAVILBAlert = !typeList.includes(LoadBalancerType.AVI);
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.LOADBALANCER.GET-LB-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  openRouteDetailsPage(routeId: string): void {
    this.detailsModalOpen = true;
    this.selectedRouteId = routeId;
  }

  refreshComponentInApplication($event: any): void {
    const appId = $event.target.value;
    this.componentList = null;
    this.cptService.getAllComponentsByApplication(appId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.componentList = !!data.data?.Components?.edge?.nodes ? data.data.Components.edge.nodes : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  getChartOption(backends: any[]): EChartsOption {
    let data: any[] = [];
    for(var i = 0; i < backends.length; i++){
          i === 0 ?
          data.push({ value: backends[0].ratio*5, name: `${backends[0].name} (${backends[0].instance.cluster.name}/${backends[0].instance.namespace})`}) :
            data.push({ value: backends[1].ratio*5, name: `${backends[1].name} (${backends[1].instance.cluster.name}/${backends[1].instance.namespace}) `})
        }
    let chartOption: EChartsOption;
    chartOption = {
      color: ['#0072a3', '#60b515'],
      tooltip: {
        show: false
      },
      legend: {
        show: false
      },
      series: [
        {
          type: 'pie',
          name: 'backend',
          width: '32px',
          height: '32px',
          radius: ['40%', '70%'],
          avoidLabelOverlap: true,
          itemStyle: {
            borderRadius: 0,
            borderColor: '#fff',
            borderWidth: 1
          },
          emphasis: {
            scale: false,
            scaleSize: 0
          },
          animation: false,
          label: {
            show: false
          },
          labelLine: {
            show: false
          },
          data: data
        }
      ]
    };
    return chartOption
  }

  showCreateRoute(): void {
    this.getAllApplication();
    this.addFieldValue();
    this.createRouteOpen = true;
  }

  setVsvipVaildtor(enabled: boolean){
    if(enabled){
      this.basicRouteForm.controls['vsvipRef'].setValidators(Validators.required);
      this.basicRouteForm.controls['vsvipRef'].updateValueAndValidity();
    } else {
      this.basicRouteForm.controls['vsvipRef'].clearValidators();
      this.basicRouteForm.controls['vsvipRef'].updateValueAndValidity();
    }
  }

  refreshInstanceInComponent($event: any): void {
    const component = $event.target.value;
    this.instanceList = null;
    this.instanceService.getInstancesByComponentId(component).subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.INSTANCE.GET-INSTANCE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      const tempList = [];
      if(!!data.data?.Instances?.edge?.nodes){
        for(const instance of data.data.Instances.edge.nodes){
          // only support of instance in success status and with avi loadbalancer
          if(instance.status.state === 'SUCCESS' && instance.cluster.loadBalancer.type === LoadBalancerType.AVI){
            tempList.push(instance)
          }
        }
      }
      this.instanceList = tempList;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.INSTANCE.GET-INSTANCE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
  }

  refreshServicesInInstance($event: any): void {
    const instanceId = $event.target.value;
    let serviceList: Service[] = [];
    this.networkProfile = null;
    for(const instance of this.instanceList){
      if (instance.id === instanceId){
        this.lbId = instance.cluster.loadBalancer.id;
        // only support service of loadbalancer type
        for(const service of instance.services){
          if(service.spec.type === 'LoadBalancer'){
            serviceList.push(service)
          }
        }
        this.servicesInInstanceM[instanceId] = serviceList;
        // query networkprofile in lb
        this.lbService.getLbNetworkprofile(this.lbId, this.basicRouteForm.value.protocol).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.LOADBALANCER.GET-LB-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if(!!data.data?.LoadBalancer?.status?.networkProfiles){
            this.networkProfile = data.data.LoadBalancer.status.networkProfiles;
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.LOADBALANCER.GET-LB-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
      }
    }
  }

  getInstanceInSameLB(): Instance[] {
    const sameLbInstance: Instance[] = [];
    if(this.lbId){
      for(const instance of this.instanceList){
        if (instance.cluster.loadBalancer.id === this.lbId){
          sameLbInstance.push(instance)
        }
      }
      return sameLbInstance;
    }
    return [];
  }

  servicesInInstance(instanceId: string): Service[] {
    if (!!instanceId) {
      return !!this.servicesInInstanceM[instanceId] ? this.servicesInInstanceM[instanceId] : [];
    }
    return [];
  }

  ipsInService(instanceId: string, name: string): any[] {
    if (!!instanceId && !!name && !!this.servicesInInstanceM) {
      const serviceList: Service[] = this.servicesInInstanceM[instanceId]
      if (!!serviceList){
        for(const service of serviceList){
          if(service.name === name){
            return !!service?.status.ingress ? service?.status.ingress : [];
          }
        }
      }
    }
    return [];
  }

  portsInService(instanceId: string, name: string): any[] {
    if (!!instanceId && !!name) {
      const serviceList: Service[] = this.servicesInInstanceM[instanceId]
      if(!!serviceList){
        for(const service of serviceList){
          if(service.name === name){
            return !!service?.spec.ports ? service?.spec.ports  : [];
          }
        }
      }
    }
    return [];
  }

  deleteFieldValue(index: number): void {
    this.appGlbParamsArray.removeAt(index);
  }

  addFieldValue(): void {
    this.appGlbParamsArray.push(this.fb.group({
      instanceId: new FormControl('', Validators.required),
      service: new FormControl('', Validators.required),
      ip: new FormControl('', Validators.required),
      port: new FormControl('', Validators.required)
    }));
    const ratio = this.appGlbParamsArray.length > 1 ? 10: 20;
    this.appGlbParamsRatio.setValue(ratio);
  }

  doRouteReset(): void {
    this.wizardnewroute.reset();
    this.basicRouteForm.reset();
    this.basicRouteForm.patchValue({
      protocol: 'TCP',
      lbAlgorithm: LBAlgorithm.LB_ALGORITHM_LEAST_CONNECTIONS,
      lbAlgorithmHash: AlgorithmHash.LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS,
      enableVsvip: false,
      persistentVsvip: false
    })
    this.appGlbRouteForm.reset();
    this.appGlbParamsRatio.setValue(20);
    while (this.appGlbParamsArray.length !== 0) {
      this.appGlbParamsArray.removeAt(0);
    }
    this.setVsvipVaildtor(false);
    this.createRouteOpen = false;
    this.applicationList = [];
    this.componentList = [];
    this.instanceList = [];
    this.lbId = '';
    this.servicesInInstanceM = {};
  }

  public doRouteFinish(): void {
    const params: CreateRoute = {
      name: this.basicRouteForm.value.name,
      description: this.basicRouteForm.value.description,
      protocol: this.basicRouteForm.value.protocol,
      port: this.basicRouteForm.value.port,
      backends: [],
      spec: {
        avi: {
          networkProfile: !!this.appGlbRouteForm.value.networkProfile ? this.appGlbRouteForm.value.networkProfile : null,
          lbAlgorithm: this.basicRouteForm.value.lbAlgorithm,
          lbAlgorithmHash: this.basicRouteForm.value.lbAlgorithm ===  LBAlgorithm.LB_ALGORITHM_CONSISTENT_HASH ? this.basicRouteForm.value.lbAlgorithmHash : null,
          lbAlgorithmConsistentHashHeader: this.basicRouteForm.value.lbAlgorithm == LBAlgorithm.LB_ALGORITHM_CONSISTENT_HASH && this.basicRouteForm.value.lbAlgorithmHash ===  AlgorithmHash.LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER ? this.basicRouteForm.value.lbAlgorithmConsistentHashHeader : null,
          persistentVsVIP: this.basicRouteForm.value.persistentVsvip
        }
      },
      component: this.basicRouteForm.value.component,
      loadBalancer: this.lbId
    };
    if (this.basicRouteForm.value.enableVsvip) {
      params.spec.avi.vsVip = this.basicRouteForm.value.vsvipRef;
    }

    switch (this.appGlbParamsArray.length) {
      case 1:
        params.backends.push({
          name: this.appGlbParamsArray.value[0].service,
          type: 'Service',
          ip: this.appGlbParamsArray.value[0].ip,
          port: this.appGlbParamsArray.value[0].port,
          instance: this.appGlbParamsArray.value[0].instanceId,
          ratio: 20,
        });
        break;
      case 0:
        break;
      default:
        params.backends.push({
          name: this.appGlbParamsArray.value[0].service,
          ip: this.appGlbParamsArray.value[0].ip,
          port: this.appGlbParamsArray.value[0].port,
          instance: this.appGlbParamsArray.value[0].instanceId,
          ratio: this.appGlbParamsRatio.value,
        });
        params.backends.push({
          name: this.appGlbParamsArray.value[1].service,
          ip: this.appGlbParamsArray.value[1].ip,
          port: this.appGlbParamsArray.value[1].port,
          instance: this.appGlbParamsArray.value[1].instanceId,
          ratio: 20 - this.appGlbParamsRatio.value,
        });
    }
    this.glbRouteService.createRoute(params).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.CREATE-ROUTE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.CreateRoute) {
          this.toast.show(`${data.data.CreateRoute.name} `+this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.CREATE-ROUTE-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.CREATE-ROUTE-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.CREATE-ROUTE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doRouteReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  public doRouteCancel(): void {
    this.doRouteReset();
  }

  showUpdateRoute(selectedRoute: Route): void {
    this.selectedRoute = selectedRoute;
    this.lbId = selectedRoute.loadBalancer.id;
    this.instanceService.getInstancesByComponentId(selectedRoute.component.id).subscribe( data => {
      if(!!data.data?.Instances?.edge?.nodes){
        for(const instance of data.data.Instances.edge.nodes){
          if(instance.status.state === 'SUCCESS'){
            this.instanceList.push(instance)
          }
        }
        for(const instance of this.instanceList){
          let serviceList: Service[] = [];
          // only support service of loadbalancer type
          for(const service of instance.services){
            if(service.spec.type === 'LoadBalancer'){
              serviceList.push(service)
            }
            this.servicesInInstanceM[instance.id] = serviceList;
          }
        }
        if (!!selectedRoute.backends) {
          for (const each of selectedRoute.backends) {
            this.oldRouteBackendIdList.push({
              id: each.id
            })
            const control = this.fb.group({
              instanceId: new FormControl(each.instance.id),
              noInstanceAlert: new FormControl(this.instanceList.length > 0 ? !this.instanceList.find(x => x.id === each.instance.id) : true),
              service: new FormControl(each.name),
              ip: new FormControl(each.ip),
              port: new FormControl(each.port),
              disable: new FormControl('disabled')
            });
            this.editappGlbParamsArray.push(control);
          }
          if (selectedRoute.backends.length > 0) {
            this.editappGlbParamsRatio.setValue(selectedRoute.backends[0].ratio);
          }
        }
      } else {
        this.lbId = '';
        this.instanceList = [];
        this.servicesInInstanceM = {};
      }
    })
    this.editRouteForm.setValue({
      name: selectedRoute.name,
      description: selectedRoute.description,
      application: selectedRoute.application.id,
      component: selectedRoute.component.id,
      protocol: selectedRoute.protocol,
      port: selectedRoute.port,
      lbAlgorithm: selectedRoute.spec.lbAlgorithm,
      lbAlgorithmHash: selectedRoute.spec.lbAlgorithmHash,
      lbAlgorithmConsistentHashHeader: selectedRoute.spec.lbAlgorithmConsistentHashHeader,
      persistentVsvip: selectedRoute.spec.persistentVsVIP
    });
    this.editappGlbRouteForm.patchValue({
      networkProfile: selectedRoute.spec.networkProfile
    })
    this.updateRouteOpen = true;
  }

  doUpdateRouteFinish(): void {
    const params: UpdateRoute = {
      id: this.selectedRoute.id,
      description: this.editRouteForm.value.description,
      port: this.editRouteForm.value.port,
      addBackends: [],
      removeBackends: [],
      spec: {
        avi: {
          lbAlgorithm: this.editRouteForm.value.lbAlgorithm,
          lbAlgorithmHash: this.editRouteForm.value.lbAlgorithm ===  LBAlgorithm.LB_ALGORITHM_CONSISTENT_HASH ? this.editRouteForm.value.lbAlgorithmHash : null,
          lbAlgorithmConsistentHashHeader: this.editRouteForm.value.lbAlgorithm ===  LBAlgorithm.LB_ALGORITHM_CONSISTENT_HASH && this.editRouteForm.value.lbAlgorithmHash ===  AlgorithmHash.LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER ? this.editRouteForm.value.lbAlgorithmConsistentHashHeader : null,
          persistentVsVIP: this.editRouteForm.value.persistentVsvip
        }
      }
    };
    params.removeBackends = this.oldRouteBackendIdList;
    switch (this.editappGlbParamsArray.length) {
      case 1:
        params.addBackends.push({
          name: this.editappGlbParamsArray.value[0].service,
          type: 'Service',
          ip: this.editappGlbParamsArray.value[0].ip,
          port: this.editappGlbParamsArray.value[0].port,
          instance: this.editappGlbParamsArray.value[0].instanceId,
          ratio: 20,
        });
        break;
      case 0:
        break;
      default:
        params.addBackends.push({
          name: this.editappGlbParamsArray.value[0].service,
          type: 'Service',
          instance: this.editappGlbParamsArray.value[0].instanceId,
          ip: this.editappGlbParamsArray.value[0].ip,
          port: this.editappGlbParamsArray.value[0].port,
          ratio: this.editappGlbRouteForm.value.ratio,
        });
        params.addBackends.push({
          name: this.editappGlbParamsArray.value[1].service,
          type: 'Service',
          instance: this.editappGlbParamsArray.value[1].instanceId,
          ip: this.editappGlbParamsArray.value[1].ip,
          port: this.editappGlbParamsArray.value[1].port,
          ratio: 20 - this.editappGlbRouteForm.value.ratio,
        });
    }
    this.glbRouteService.updateRoute(params).subscribe((data) => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.UPDATE-ROUTE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.UpdateRoute) {
        this.toast.show(`${data.data.UpdateRoute.name} `+this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.UPDATE-ROUTE-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.UPDATE-ROUTE-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.UPDATE-ROUTE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doEditRouteReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doUpdateRouteCancel(): void {
    this.doEditRouteReset();
  }

  doEditRouteReset(): void {
    this.wizardeditroute.reset();
    this.editappGlbRouteForm.reset();
    this.editappGlbParamsRatio.setValue(20);
    this.editRouteForm.reset();
    this.editRouteForm.patchValue({
      lbAlgorithm: LBAlgorithm.LB_ALGORITHM_LEAST_CONNECTIONS,
      lbAlgorithmHash: AlgorithmHash.LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS,
      persistentVsvip: false
    })
    while (this.editappGlbParamsArray.length !== 0) {
      this.editappGlbParamsArray.removeAt(0);
    }
    this.updateRouteOpen = false;
    this.componentList = [];
    this.lbId = '';
    this.instanceList = [];
    this.servicesInInstanceM = {};
    this.oldRouteBackendIdList = [];
  }

  deleteFieldValueForUpdate(index: number): void {
    this.editappGlbParamsArray.removeAt(index);
  }

  addFieldValueForUpdate(): void {
    this.editappGlbParamsArray.push(this.fb.group({
      instanceId: new FormControl('', Validators.required),
      noInstanceAlert: new FormControl(false),
      service: new FormControl('', Validators.required),
      ip: new FormControl('', Validators.required),
      port: new FormControl('', Validators.required),
      disable: new FormControl('')
    }));
    const ratio = this.editappGlbParamsArray.length > 1 ? 10: 20;
    this.editappGlbParamsRatio.setValue(ratio);
  }

  showDeleteRoute(): void {
    this.deleteRouteOpen = true;
  }

  deleteRoute(): void {
    this.deleteRouteOpen = false;
    this.glbRouteService.deleteRoute(this.selectedRoute.id).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.DELETE-ROUTE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.DeleteRoute) {
          this.toast.show(`${data.data.DeleteRoute.name} `+this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.DELETE-ROUTE-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.DELETE-ROUTE-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.DELETE-ROUTE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  showForceDeleteRoute(): void {
    this.forceDeleteRouteOpen = true;
  }

  forceDeleteRoute(): void {
    this.forceDeleteRouteOpen = false;
    this.glbRouteService.forceDeleteRoute(this.selectedRoute.id).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.DELETE-ROUTE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.DeleteRoute) {
          this.toast.show(`${data.data.DeleteRoute.name} `+this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.DELETE-ROUTE-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.DELETE-ROUTE-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.DELETE-ROUTE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

}
