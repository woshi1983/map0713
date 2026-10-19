import { Component, OnInit } from '@angular/core';
import { AuthService } from '../../../services/auth.service';
import { ActivatedRoute, Router } from '@angular/router';
import { ClustersService } from '../../../services/clusters.service';
import {Cluster, handleObject, EnableClusterExtension, VMClass, ClusterExtension, UpdateCluster} from '../../../classes';
import { FormControl, FormGroup, Validators } from "@angular/forms";
import semver from 'semver';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from "@ngx-translate/core";
import { saveAs } from 'file-saver';
import {OrganizationService} from "../../../services/organization.service";
import {ClusterExtensionService} from "../../../services/cluster-extension.service";
import {FEATUREGATES} from "../../../utils/static-entries/shared.const";

@Component({
  selector: 'app-clusterdetail',
  templateUrl: './clusterdetail.component.html',
  styleUrls: ['./clusterdetail.component.css']
})
export class ClusterdetailComponent implements OnInit {

  //admin role can handle cluster create/update/delete
  get canWrite(): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER);
  }

  //admin and user role can handle cluster resources
  get canWriteResources(): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER_RESOURCE);
  }

  //only superadmin can handle cluster force delete
  get canForceDelete(): boolean {
    return this.authService.can('NULL', handleObject.CLUSTER);
  }

  featureGates = FEATUREGATES
  orgId: string;
  clusterId: string;
  clusterName: string;
  cluster: Cluster;

  deleteClusterDialog = false;
  removeClusterDialog = false;
  forceDeleteDialog = false;
  allowUpdateClusterMaster =false;
  showUpdateWorkerDialog = false;
  updateWorkerChanged = false;
  showUpdateTKGSMasterDialog = false;
  showUpdateTKGMMasterDialog = false;
  showClusterUpdateDialog = false;
  allowUpdateClusterVersion = false;
  showUpdateTKGSVMclassDialog = false;
  updateTKGSVMclassChanged = false;
  showUpdateTKGMVMclassDialog = false;
  updateTKGMVMclassChanged = false;
  showUpdateClusterDescDialog = false;
  showNewCLusterVersionAlert = false;
  showUnableGetResourceAlert = false;
  showNoAdjacentVersion = false;
  allowUpdateVersionList = [];

  updateTKGSClusterWorkerForm = new FormGroup({
    workerCount: new FormControl(1, [Validators.required, Validators.min(1), Validators.max(150)])
  });
  updateTKGMClusterMasterForm = new FormGroup({
    masterCount: new FormControl(1, [Validators.required])
  });
  updateClusterVersionForm = new FormGroup({
    version: new FormControl('', [Validators.required]),
  });
  updateTKGSClusterVMClassForm = new FormGroup({
    masterVMClass: new FormControl('', [Validators.required]),
    workerVMClass: new FormControl('', [Validators.required])
  });
  updateTKGMClusterVMClassForm = new FormGroup({
    masterCPU: new FormControl(''),
    masterRAM: new FormControl(''),
    masterDisk: new FormControl(''),
    workerCPU: new FormControl(''),
    workerRAM: new FormControl(''),
    workerDisk: new FormControl(''),
  });
  updateClusterDescForm = new FormGroup({
    description: new FormControl('')
  })
  deleteClusterForm = new FormGroup({
    confirm: new FormControl('', [Validators.required]),
  });

  terminalExtension: ClusterExtension;
  tbsExtension: ClusterExtension;
  contourExtension: ClusterExtension;
  changeExtensionStatus: boolean;
  showUpdateExtensionDialog: string;
  contourInfo = false;
  octantLink: string;
  showTerminal = false;

  tkcVersionList: string[] = [];
  tkcvmClassList: VMClass[] = [];
  contourIpList: string[] = [];
  contourPortList: string[] = [];

  loading = false;

  tabLinkStatusList = {
    'summaryActive': true,
    'computeActive': false,
    'namespaceActive': false,
    'instanceActive': false,
    'workloadActive': false,
    'backupActive': false,
    'locationActive': false
  }

  ingressExample = 'apiVersion: networking.k8s.io/v1\n' +
    'kind: Ingress\n' +
    'metadata:\n' +
    '  name: ingress-example\n' +
    'spec:\n' +
    '  ingressClassName: contour\n' +
    '  rules:\n' +
    '  - host: ingress.example.com\n' +
    '    http:\n' +
    '      paths:\n' +
    '      - path: /\n' +
    '        pathType: Prefix\n' +
    '        backend:\n' +
    '          service:\n' +
    '            name: example\n' +
    '            port:\n' +
    '              number: 8080'

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    public authService: AuthService,
    private clusterService: ClustersService,
    private toast: ToastrService,
    private translate: TranslateService,
    private orgService: OrganizationService,
    private clusterExtensionsService: ClusterExtensionService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  ngOnInit(): void {
    this.refresh()
  }

  refresh(): void {
    this.clusterId = this.route.snapshot.paramMap.get('id');
    this.clusterName = this.route.snapshot.paramMap.get('name');
    this.loading = true;
    this.showUnableGetResourceAlert = false;
    this.allowUpdateClusterVersion = false;
    this.showNewCLusterVersionAlert = false;
    this.cluster = null;
    this.contourExtension = null;
    this.tbsExtension = null;
    this.tkcVersionList = [];
    this.contourIpList = [];
    this.contourPortList = [];
    this.clusterService.getClusterInfo(this.clusterId).subscribe(data => {
      if (!!data?.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
      }
      if (!!data?.data?.Cluster) {
        this.cluster = data.data.Cluster;
        if(this.cluster.status === 'PENDING' || this.cluster.status === 'INITIALIZING'){
          this.showUnableGetResourceAlert = true
        }
        if(!!this.cluster.extensions?.edge?.nodes){
          this.tbsExtension = !!this.cluster.extensions.edge.nodes.find(x => x.name === 'TBS') ? this.cluster.extensions.edge.nodes.find(x => x.name === 'TBS') : null;
          this.contourExtension = !!this.cluster.extensions.edge.nodes.find(x => x.name === 'CONTOUR') ? this.cluster.extensions.edge.nodes.find(x => x.name === 'CONTOUR') : null;
          if(!!this.contourExtension?.endpoints){
            const endpoints = this.contourExtension?.endpoints;
            for(let ep of endpoints){
              const splitList = ep.split(':');
              if(!this.contourIpList.find(x => x === splitList[1].slice(2))){
                this.contourIpList.push(splitList[1].slice(2))
              }
              const port = splitList[0] === 'http' ? splitList[0] + ': ' + (!!splitList[2] ? splitList[2] : '80') : splitList[0] + ': ' + (!!splitList[2] ? splitList[2] : '443');
              if(!this.contourPortList.find(x => x === port)){
                this.contourPortList.push(port)
              }
            }
          }
        }
        this.getOctantLink();
        if(this.cluster.type === 'PROVISIONED' && !!this.cluster.provider.spec.clusterReleases){
          //some logic in tkgs and tkgm
          this.tkcVersionList = this.cluster.provider.spec.clusterReleases;
          // new version exist
          if(semver.gt(semver.maxSatisfying(this.tkcVersionList, '*'), this.cluster.version)){
            this.allowUpdateClusterVersion = true;
            this.showNewCLusterVersionAlert = true;
          }
        }
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {this.loading = false});
  }

  backToManagement(): void {
    this.router.navigate(['/portal/cluster/management']);
  }

  showClusterCause(): void {

  }

  downloadKubeconfig(): void {
    const blob = new Blob([this.cluster.kubeConfig], { type: 'text/json' });
    saveAs(blob, this.cluster.name + '-cluster.kubeconfig');
  }

  downloadSSHKey(): void {
    const blob = new Blob([this.cluster.sshPrivateKey], { type: 'text' });
    saveAs(blob, this.cluster.name + '-sshkey.pem');
  }

  getOctantLink(): void {
    this.octantLink = null;
    this.clusterExtensionsService.getClusterExtensionRelated(this.clusterId, 'OCTANT').subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-OCTANT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if(!!data.data.ClusterExtensions?.edge?.nodes && data.data.ClusterExtensions.edge.nodes[0].state === 'ENABLED'){
        this.octantLink = data.data.ClusterExtensions.edge.nodes[0].endpoints[0]
      }
    },(e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.GET-CLUSTER-OCTANT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
  }

  openOctantLink(): void {
    window.open(this.octantLink, "_blank");
  }

  showTerminalHelp(): void {
    this.showTerminal = true;
    this.terminalExtension = !!this.cluster.extensions.edge.nodes.find(x => x.name === 'WEB_TERMINAL') ? this.cluster.extensions.edge.nodes.find(x => x.name === 'WEB_TERMINAL') : null;
  }

  openTerminalURL(): void {
    window.open(this.terminalExtension.endpoints[0], "_blank");
  }

  showClusterUpdate(): void {
    this.showClusterUpdateDialog = true;
    // this param is check ova template version is next to.
    this.showNoAdjacentVersion = false;
    // current version
    const currentVerson = this.cluster.version;
    // sorted version list
    const allTkcVersion = this.tkcVersionList.concat([currentVerson])
    const sortedVersionList = allTkcVersion.sort((n1, n2) => {
      if(semver.gt(n1, n2)){return 1;}
      if(semver.lt(n1,n2)){return -1;}
      return 0;
    });
    // remove duplicate item
    for(let i =0;i < sortedVersionList.length - 1;i++){
      if(semver.eq(sortedVersionList[i], sortedVersionList[i+1])){
        sortedVersionList.splice(i,1);
      }
    }
    // get current major and minor max version, such as 1.19
    const currentMajorAndMinor = semver.major(currentVerson) + '.' + semver.minor(currentVerson);
    // get current max version, such as 1.19.20
    const currentMaxVersion = semver.maxSatisfying(sortedVersionList, currentMajorAndMinor);
    // get current version index and current max version index, such as 2, 4
    // currentIndex can be -1
    let currentIndex = sortedVersionList.indexOf(currentVerson);
    let currentMaxIndex = sortedVersionList.indexOf(currentMaxVersion, currentIndex);

    if (currentMaxIndex === sortedVersionList.length - 1) {
      // get all allow update versions in current version
      this.allowUpdateVersionList = sortedVersionList.slice(currentIndex + 1);
    } else {
      // get next version
      const nextVersion = sortedVersionList[currentMaxIndex + 1];
      // check adjoin with currentVersion and nextVersion
      if(Number(semver.minor(nextVersion)) - Number(semver.minor(currentVerson)) === 1){
        // get next max version, 1.20
        const nextMajorAndMinor = semver.major(nextVersion) + '.' + semver.minor(nextVersion);
        const nextMaxVersion = semver.maxSatisfying(sortedVersionList, nextMajorAndMinor);
        // get next max version index
        const nextMaxVersionIndex = sortedVersionList.indexOf(nextMaxVersion, currentMaxIndex);
        // get all allow update versions
        this.allowUpdateVersionList = sortedVersionList.slice(currentIndex + 1, nextMaxVersionIndex + 1);
      } else {
        // can update, but no adjacent version found
        this.showNoAdjacentVersion = this.allowUpdateVersionList.length === 0;
      }
    }
  }

  showDeleteClusterDialog(): void {
    this.deleteClusterDialog = true;
  }

  deleteClusterById(id: string): void {
    this.deleteClusterDialog = false;
    this.clusterService.deleteCluster(id).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             if(each.message === 'object is busy'){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-ALERT'), this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             } else {
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
           }
        }
      }
      if (!!data.data?.DeleteCluster){
        this.toast.show(`${data.data.DeleteCluster.name}`+this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
        this.router.navigate(['/portal/cluster/management']);
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  showRemoveClusterDialog(): void {
    this.removeClusterDialog = true;
  }

  removeClusterById(id: string): void {
    this.removeClusterDialog = false;
    this.clusterService.deleteCluster(id).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             if(each.message === 'object is busy'){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-ALERT'), this.translate.instant('MESSAGE.CLUSTER.REMOVE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             } else {
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.REMOVE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
           }
        }
      }
      if (!!data.data?.DeleteCluster){
        this.toast.show(`${data.data.DeleteCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.REMOVE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.REMOVE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
        this.router.navigate(['/portal/cluster/management']);
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.REMOVE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  showForceRemoveDialog(): void {
    this.forceDeleteDialog = true;
  }

  forceDeleteClusterById(id: string): void {
    this.forceDeleteDialog = false;
    this.clusterService.forceDeleteCluster(id).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             if(each.message === 'object is busy'){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-ALERT'), this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             } else {
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
           }
        }
      }
      if (!!data.data?.DeleteCluster){
        this.toast.show(`${data.data.DeleteCluster.name}`+this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
        this.router.navigate(['/portal/cluster/management']);
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.DELETE-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  doReinitCluster(): void {
    this.clusterService.ReinitCluster(this.clusterId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-REINIT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.UpdateCluster){
        this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-REINIT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
    this.refresh();
    })
  }

  showUpdateClusterDesc(): void {
    this.updateClusterDescForm.setValue({
      description: this.cluster.description,
    });
    this.showUpdateClusterDescDialog = true;
  }

  doUpdateClusterDescription(): void {
    this.showUpdateClusterDescDialog = false;
    this.clusterService.updateClusterDesc(this.clusterId, this.updateClusterDescForm.value.description).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-DESC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.UpdateCluster){
        this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-DESC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
    this.refresh();
    })
  }

  showUpdateClusterVMClass(): void {
    switch (this.cluster.class) {
      case 'TKGS':
        this.showUpdateTKGSVMclassDialog = true;
        this.updateTKGSVMclassChanged = false;
        this.tkcvmClassList = this.cluster.provider.spec.vmClasses;
        this.updateTKGSClusterVMClassForm.patchValue({
          masterVMClass: this.cluster.spec.ctlPlaneClass,
          workerVMClass: this.cluster.spec.dataPlaneClass
        });
        break;
      case 'TKG2M':
      case 'TKGM':
        this.showUpdateTKGMVMclassDialog = true;
        this.updateTKGMVMclassChanged = false;
        this.updateTKGMClusterVMClassForm.patchValue({
          masterCPU: this.cluster.class === 'TKG2M' ? this.cluster.spec.ctlPlaneNumCpus : this.cluster.spec.ctlPlaneCPUs,
          masterRAM: (this.cluster.class === 'TKG2M' ? this.cluster.spec.ctlPlaneMemMiB : this.cluster.spec.ctlPlaneMemMi) / 1024,
          masterDisk: this.cluster.class === 'TKG2M' ? this.cluster.spec.ctlPlaneDiskGiB : this.cluster.spec.ctlPlaneStorageCapGi,
          workerCPU: this.cluster.class === 'TKG2M' ? this.cluster.spec.workerNumCpus : this.cluster.spec.dataPlaneCPUs,
          workerRAM: (this.cluster.class === 'TKG2M' ? this.cluster.spec.workerMemMiB : this.cluster.spec.dataPlaneMemMi) / 1024,
          workerDisk: this.cluster.class === 'TKG2M' ? this.cluster.spec.workerDiskGiB : this.cluster.spec.dataPlaneStorageCapGi
        })
        break;
      default:
        console.log('not support')
    }
  }

  doUpdateClusterVMClass(): void {
    switch (this.cluster.class) {
      case 'TKGS': {
        this.tkcvmClassList = [];
        this.showUpdateTKGSVMclassDialog = false;
        let input: UpdateCluster = {
          id: this.clusterId,
          spec: {
            tkgs: {
              ctlPlaneClass: this.updateTKGSClusterVMClassForm.value.masterVMClass,
              dataPlaneClass: this.updateTKGSClusterVMClassForm.value.workerVMClass
            }
          }
        };
        this.clusterService.updateClusterVMClass(input).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-VMCLASS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.UpdateCluster){
            this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-VMCLASS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
        this.refresh();
        })
        break;
      }
      case 'TKG2M':
      case 'TKGM': {
        this.showUpdateTKGMVMclassDialog = false;
        const oldCtlResourceSetting = this.cluster.class === 'TKG2M'
          ? this.cluster.spec.ctlPlaneNumCpus + this.cluster.spec.ctlPlaneMemMiB / 1024 + this.cluster.spec.ctlPlaneDiskGiB
          : this.cluster.spec.ctlPlaneCPUs + this.cluster.spec.ctlPlaneMemMi / 1024 + this.cluster.spec.ctlPlaneStorageCapGi;
        const oldDataResourceSetting = this.cluster.class === 'TKG2M'
          ? this.cluster.spec.workerNumCpus + this.cluster.spec.workerMemMiB / 1024 + this.cluster.spec.workerDiskGiB
          : this.cluster.spec.dataPlaneCPUs + this.cluster.spec.dataPlaneMemMi / 1024 + this.cluster.spec.dataPlaneStorageCapGi;
        const newCtlResourceSetting = this.updateTKGMClusterVMClassForm.value.masterCPU + this.updateTKGMClusterVMClassForm.value.masterRAM + this.updateTKGMClusterVMClassForm.value.masterDisk;
        const newDataResourceSetting = this.updateTKGMClusterVMClassForm.value.workerCPU + this.updateTKGMClusterVMClassForm.value.workerRAM + this.updateTKGMClusterVMClassForm.value.workerDisk;
        let input: UpdateCluster = {
          id: this.clusterId,
          spec: this.cluster.class === 'TKG2M' ? {
            tkg2m: {
              ctlPlaneNumCpus: oldCtlResourceSetting !== newCtlResourceSetting ? this.updateTKGMClusterVMClassForm.value.masterCPU : null,
              ctlPlaneMemMiB: oldCtlResourceSetting !== newCtlResourceSetting ? this.updateTKGMClusterVMClassForm.value.masterRAM * 1024 : null,
              ctlPlaneDiskGiB: oldCtlResourceSetting !== newCtlResourceSetting ? this.updateTKGMClusterVMClassForm.value.masterDisk : null,
              workerNumCpus: oldDataResourceSetting !== newDataResourceSetting ? this.updateTKGMClusterVMClassForm.value.workerCPU : null,
              workerMemMiB: oldDataResourceSetting !== newDataResourceSetting ? this.updateTKGMClusterVMClassForm.value.workerRAM * 1024 : null,
              workerDiskGiB: oldDataResourceSetting !== newDataResourceSetting ? this.updateTKGMClusterVMClassForm.value.workerDisk : null
            }
          } : {
            tkgm: {
              ctlPlaneCPUs: oldCtlResourceSetting !== newCtlResourceSetting ? this.updateTKGMClusterVMClassForm.value.masterCPU : null,
              ctlPlaneMemMi: oldCtlResourceSetting !== newCtlResourceSetting ? this.updateTKGMClusterVMClassForm.value.masterRAM * 1024 : null,
              ctlPlaneStorageCapGi: oldCtlResourceSetting !== newCtlResourceSetting ? this.updateTKGMClusterVMClassForm.value.masterDisk : null,
              dataPlaneCPUs: oldDataResourceSetting != newDataResourceSetting ? this.updateTKGMClusterVMClassForm.value.workerCPU : null,
              dataPlaneMemMi: oldDataResourceSetting != newDataResourceSetting ? this.updateTKGMClusterVMClassForm.value.workerRAM * 1024 : null,
              dataPlaneStorageCapGi: oldDataResourceSetting != newDataResourceSetting ? this.updateTKGMClusterVMClassForm.value.workerDisk : null
            }
          }
        };
        if (oldCtlResourceSetting !== newCtlResourceSetting || oldDataResourceSetting != newDataResourceSetting){
          this.clusterService.updateClusterVMClass(input).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-VMCLASS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.UpdateCluster){
              this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-VMCLASS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          this.refresh();
         })
        } else {
          this.toast.show(this.translate.instant('MESSAGE.CLUSTER.UPDATE-RESOURCE-NOCHANGE'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
        }
        break;
      }
      default:
        console.log('not support.')
    }
  }

  showUpdateClustrMaster(): void {
    this.allowUpdateClusterMaster = false;
    switch (this.cluster.class) {
      case 'TKGS':
        this.showUpdateTKGSMasterDialog = true;
        if (this.cluster.spec.ctlPlaneCount === 1) {
          this.allowUpdateClusterMaster = true;
        }
        break;
      case 'TKGM':
        this.showUpdateTKGMMasterDialog = true;
        this.updateTKGMClusterMasterForm.setValue({
          masterCount: this.cluster.spec.ctlPlaneCount
        })
        break;
      default:
        console.log('not support.')
        break;
    }
  }

  doUpdateClusterMaster(): void {
    switch (this.cluster.class) {
      case 'TKGS':
        this.showUpdateTKGSMasterDialog = false;
        this.clusterService.updateTKGSClusterMasterCount(this.clusterId).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSATER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.UpdateCluster){
            this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSATER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
        this.refresh();
        })
        break;
      case 'TKGM':
        this.showUpdateTKGMMasterDialog = false;
        this.clusterService.updateTKGMClusterMasterCount(this.clusterId, this.updateTKGMClusterMasterForm.value.masterCount).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSATER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.UpdateCluster){
            this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSATER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
        this.refresh();
        })
        break;
      default:
        console.log('not support')
    }
  }

  showUpdateClusterWorker(): void {
    this.showUpdateWorkerDialog = true;
    this.updateWorkerChanged = false;
    this.updateTKGSClusterWorkerForm.patchValue({
      workerCount: this.cluster.spec.dataPlaneCount
    });
  }

  doUpdateClusterWorker(): void {
    this.showUpdateWorkerDialog = false;
    switch (this.cluster.class) {
      case 'TKGS':
         this.clusterService.updateTKGSClusterWorkerCount(this.clusterId, this.updateTKGSClusterWorkerForm.value.workerCount).subscribe(data => {
           if (!!data.errors){
              for (const each of data.errors){
                 if(!!each.message){
                   this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-WORKER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
                 }
              }
            }
          if (!!data.data?.UpdateCluster){
            this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-WORKER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
         this.refresh();
         })
        break;
      case 'TKGM':
         this.clusterService.updateTKGMClusterWorkerCount(this.clusterId, this.updateTKGSClusterWorkerForm.value.workerCount).subscribe(data => {
           if (!!data.errors){
              for (const each of data.errors){
                 if(!!each.message){
                   this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-WORKER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
                 }
              }
            }
          if (!!data.data?.UpdateCluster){
            this.toast.show(`${data.data.UpdateCluster.name} `+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-WORKER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
         this.refresh();
         })
        break;
      default:
        console.log('not support')
    }
  }

  doUpdateClusterVersion(id: string, version: string): void {
    this.showClusterUpdateDialog = false;
     this.clusterService.updateClusterVersion(id, version).subscribe(data => {
       if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-VERSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
        }
      if (!!data.data?.UpdateCluster){
        this.toast.show(`${data.data.UpdateCluster.name}`+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-VERSION-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
     this.refresh();
     })
  }

  showUpdateExtensionStatus(extansion: string ,changeStatus: boolean): void {
    this.changeExtensionStatus = changeStatus;
    this.showUpdateExtensionDialog = extansion;
  }

  updateExtensionStatus(): void {
    if(this.changeExtensionStatus){
      // enable
      const param: EnableClusterExtension = {
        name: this.showUpdateExtensionDialog,
        cluster: this.clusterId
    }
    this.clusterExtensionsService.enableClusterExtensions(param).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-TBS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.EnableClusterExtension){
        this.toast.show(`${data.data.EnableClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-TBS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
       this.changeExtensionStatus = null;
       this.showUpdateExtensionDialog = null;
       this.refresh();
     })
    } else {
      // disable
      const id = this.showUpdateExtensionDialog === 'TBS' ? this.tbsExtension?.id : this.contourExtension?.id;
      this.clusterExtensionsService.disableClusterExtensions(id).subscribe( data => {
        if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-TBS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
        }
        if (!!data.data?.DisableClusterExtension){
        this.toast.show(`${data.data.DisableClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-TBS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
       this.changeExtensionStatus = null;
       this.showUpdateExtensionDialog = null;
       this.refresh();
     })
    }
  }

  updateTerminalExtension(enable: boolean){
    if(enable){
      const param: EnableClusterExtension = {
        name: 'WEB_TERMINAL',
        cluster: this.clusterId
      }
      this.clusterExtensionsService.enableClusterExtensions(param).subscribe(data => {
        if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-TBS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
        }
        if (!!data.data?.EnableClusterExtension){
          this.toast.show(`${data.data.EnableClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
        }
      }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-TBS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
         this.terminalExtension = null;
         this.showTerminal = false;
         this.refresh();
       })
    } else {
      this.clusterExtensionsService.disableClusterExtensions(this.terminalExtension.id).subscribe( data => {
        if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-TBS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
        }
        if (!!data.data?.DisableClusterExtension){
          this.toast.show(`${data.data.DisableClusterExtension.cluster.name}`+this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-MSG'), this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
        }
      }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER.UPDATE-CLUSTER-TBS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () =>{
         this.terminalExtension = null;
         this.showTerminal = false;
         this.refresh();
       })
    }
  }

  showContourInfo(): void {
    this.contourInfo = true;
  }



}
