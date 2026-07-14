import {Component, OnInit, ViewChild, Input} from '@angular/core';
import { AuthService} from "../../../../services/auth.service";
import {ClrDatagridPagination} from "@clr/angular";
import {ProviderService} from "../../../../services/provider.service";
import {ActivatedRoute} from "@angular/router";
import {ClusterProvider, TKGSClusterProviderSpecNamespace} from "../../../../classes";
import { ToastrService } from 'ngx-toastr';
import {TranslateService} from "@ngx-translate/core";


@Component({
  selector: 'app-supervisor-namespace',
  templateUrl: './namespace.component.html',
  styleUrls: ['./namespace.component.css']
})
export class SupervisorNamespaceComponent implements OnInit {

  @Input() providerId;

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  loading = false;
  provider: ClusterProvider;
  namespaceList: TKGSClusterProviderSpecNamespace[] = [];
  showSearch = false;

  constructor(
    public authService: AuthService,
    private providerService: ProviderService,
    private route: ActivatedRoute,
    private toast: ToastrService,
    private translate: TranslateService
  ) { }

  ngOnInit(): void {
  }

  refresh(): void {
    this.loading = true;
    this.providerService.getNamespaceByProvider(this.providerId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.PROVIDERDETAILS.NAMESPACE.GET-SUPERVISR-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.ClusterProvider?.spec?.namespaces) {
        this.provider = data.data.ClusterProvider;
        this.namespaceList = data.data.ClusterProvider.spec.namespaces;
      } else {
        this.provider = null;
        this.namespaceList = [];
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.PROVIDERDETAILS.NAMESPACE.GET-SUPERVISR-NS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false);
  }



}
