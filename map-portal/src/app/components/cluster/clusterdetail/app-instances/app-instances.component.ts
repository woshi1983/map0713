import {Component, Input, OnInit, ViewChild} from '@angular/core';
import {AuthService} from "../../../../services/auth.service";
import {NamespaceService} from "../../../../services/namespace.service";
import {ClustersService} from "../../../../services/clusters.service";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {ClrDatagridPagination, ClrDatagridStateInterface} from "@clr/angular";
import {Instance} from "../../../../classes";
import {InstanceService} from "../../../../services/instance.service";
import {Router} from "@angular/router";

@Component({
  selector: 'app-app-instances',
  templateUrl: './app-instances.component.html',
  styleUrls: ['./app-instances.component.css']
})
export class AppInstancesComponent implements OnInit {

  @Input() clusterId: string;
  allInstances: Instance[] = [];

  showSearch = false;
  searchName: string;
  total = 0;
  loading = false;

  @ViewChild('pagination') pagination: ClrDatagridPagination;

  constructor(
    public authService: AuthService,
    private instanceService: InstanceService,
    private namespaceService: NamespaceService,
    private clusterService: ClustersService,
    private toast: ToastrService,
    private translate: TranslateService,
    private router: Router
  ) { }

  ngOnInit(): void {
  }

  refresh(state?: ClrDatagridStateInterface): void {
    this.loading = true;
    const filters = [];
    filters.push({ key: 'cluster', value: [this.clusterId] });
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.instanceService.getInstancesByCluster(!!state?.page?.current ? state.page.current - 1 : 0, !!state?.page?.size ? state.page.size : Number(this.pagination.pageSize), filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.INSTANCE.GET-INSTANCE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.allInstances = !!data.data?.Instances?.edge?.nodes ? data.data.Instances.edge.nodes : [];
      this.total = !!data.data?.Instances?.totalCount ? data.data.Instances.totalCount : 0;
      }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.INSTANCE.GET-INSTANCE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false);
  }

  routeToComponent(componentId: string, componentName: string): void {
    localStorage.setItem('app.tabs.active', 'componentActive')
    this.router.navigate(['/portal/appmgmt/components/' + componentId + '/' + componentName]);
  }

}
