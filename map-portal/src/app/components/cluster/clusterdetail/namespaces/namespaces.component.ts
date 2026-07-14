import { Component, OnInit, ViewChild, Input } from '@angular/core';
import { AuthService} from "../../../../services/auth.service";
import {ClrDatagridPagination, ClrWizard} from "@clr/angular";
import { ClustersService } from "../../../../services/clusters.service";
import {Namespace, CreateNamespace, handleObject} from "../../../../classes";
import { CustomValidatorDirective, nameValidator } from "../../../../utils/validators/custom-validator.directive";
import { FormArray, FormControl, FormGroup, Validators } from "@angular/forms";
import { NamespaceService } from 'app/services/namespace.service';
import { Namespacediffset } from 'app/utils/operator/namespacediffset';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from "@ngx-translate/core";
import {OrganizationService} from "../../../../services/organization.service";

@Component({
  selector: 'app-cluster-namespaces',
  templateUrl: './namespaces.component.html',
  styleUrls: ['./namespaces.component.css']
})
export class NamespacesComponent implements OnInit {

  //admin and user role can handle cluster resources
  get canWriteResources(): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER_RESOURCE);
  }

  constructor(
    public authService: AuthService,
    private namespaceService: NamespaceService,
    private clusterService: ClustersService,
    private toast: ToastrService,
    private orgService: OrganizationService,
    private translate: TranslateService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
  }

  orgId: string;
  allNamespace: Namespace[] = [];
  selectedNamespace: Namespace;
  @Input() clusterId: string;
  @Input() clusterStatus: string;

  nsCreateOpen = false;
  nsUpdateOpen = false;
  removeDialog = false;

  showNamespaceSearchInput = false;
  searchName: string;
  Loading = false;

  readonly disAllowLabels = CustomValidatorDirective.disAllowLabels;
  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardnewns') wizardnewns: ClrWizard;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  nsLabels = new FormArray([]);
  basicNamespaceForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator(['vmware-system-map'])]),
    labels: this.nsLabels
  });
  updateNsLabels = new FormArray([]);
  updateNamespaceForm = new FormGroup({
    labels: this.updateNsLabels
  });
  deleteNsForm = new FormGroup({
    confirm: new FormControl('', [Validators.required]),
  });


  ngOnInit(): void {
  }

  refresh(): void {
    this.Loading = true;
    const filters = [];
    filters.push({ key: 'cluster', value: [this.clusterId] });
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.namespaceService.getNamespaces(filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.GET-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.allNamespace = !!data?.data?.Namespaces ? data.data.Namespaces : [];
      this.selectedNamespace = null;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.GET-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.Loading = false);
  }

  checkNamespace(ns: string): boolean {
    const excludeNameList: string[] = Namespacediffset.excludedNamespaceNmaeList;
    return !excludeNameList.find(x => x === ns);
  }

  showCreateNS(): void {
    this.nsCreateOpen = true;
  }

  showUpdateNsLabels(ns: Namespace): void {
    this.nsUpdateOpen = true;
    if (!!ns.labels) {
      const labels = ns.labels;
      for (const key of Object.keys(labels)) {
        this.updateNsLabels.push(new FormGroup({
          key: new FormControl(key),
          value: new FormControl(labels[key])
        }));
      }
    }
  }

  doNamespaceReset(): void {
    this.basicNamespaceForm.reset();
    this.wizardnewns.reset();
    while (this.nsLabels.length !== 0) {
      this.nsLabels.removeAt(0);
    }
    this.nsCreateOpen = false;
    this.updateNamespaceForm.reset();
    while (this.updateNsLabels.length !== 0) {
      this.updateNsLabels.removeAt(0);
    }
    this.nsUpdateOpen = false;
  }

  doNamespaceFinish(): void {
    const namespaceParam: CreateNamespace = {
      name: this.basicNamespaceForm.value.name,
      cluster: this.clusterId
    };
    const labels: JSON = JSON.parse(JSON.stringify({}));
    if (!!this.basicNamespaceForm.value.labels) {
      for (const each of this.basicNamespaceForm.value.labels) {
        labels[each.key] = each.value;
      }
      namespaceParam.labels = labels;
    }
    this.namespaceService.createNamespace(namespaceParam).subscribe((data) => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.CREATE-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if(!!data?.data?.CreateNamespace){
          this.toast.show(`${data?.data.CreateNamespace.name} `+this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.CREATE-NS-MSG'), this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.CREATE-NS-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.CREATE-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doNamespaceReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doNamespaceCancel(): void {
    this.doNamespaceReset();
  }

  addnewnslabel(): void {
    this.nsLabels.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  deletenslable(index: number): void{
    this.nsLabels.removeAt(index);
  }

  addupdatenslabel(): void {
    this.updateNsLabels.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('', Validators.required)
    }));
  }

  deleteupdatenslable(index: number): void{
    this.updateNsLabels.removeAt(index);
  }

  updateNamespaceLables(ns: Namespace): void {
   const namespaceParam: CreateNamespace = {
      name: ns.name,
      cluster: this.clusterId
    };
    const labels: JSON = JSON.parse(JSON.stringify({}));
    if (!!this.updateNamespaceForm.value.labels) {
      for (const each of this.updateNamespaceForm.value.labels) {
        labels[each.key] = each.value;
      }
      namespaceParam.labels = labels;
    }
    this.namespaceService.updateNamespaceLabels(namespaceParam).subscribe((data) => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.UPDATE-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if(data?.data?.UpdateNamespace != null){
          this.toast.show(`${data?.data.UpdateNamespace.name} `+this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.UPDATE-NS-MSG'), this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.UPDATE-NS-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.UPDATE-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doNamespaceReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  showDeleteNamespace(): void {
    this.removeDialog = true;
  }

  deleteNamespace(ns: Namespace): void {
    this.namespaceService.deleteNamespace(ns.name, this.clusterId).subscribe((data) => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             if(each.message === 'object is busy'){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.DELETE-NS-ALERT'), this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.DELETE-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             } else {
               this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.DELETE-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
           }
        }
      }
      if(data?.data?.DeleteNamespace != null){
          this.toast.show(`${data?.data.DeleteNamespace.name} `+this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.DELETE-NS-MSG'), this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.DELETE-NS-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.NAMESPACE.DELETE-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.removeDialog= false;
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

}
