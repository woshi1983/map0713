import { Component, OnInit, ViewChild, Input } from '@angular/core';
import { FormControl, FormGroup, Validators } from "@angular/forms";
import { AuthService } from "../../../../services/auth.service";
import { ClrDatagridPagination } from "@clr/angular";
import { ActivatedRoute } from "@angular/router";
import { JobsService } from "../../../../services/jobs.service";
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from "@ngx-translate/core";
import { ClustersService } from "../../../../services/clusters.service";

@Component({
  selector: 'app-cluster-workloads',
  templateUrl: './cluster-workloads.component.html',
  styleUrls: ['./cluster-workloads.component.css']
})
export class ClusterWorkloadsComponent implements OnInit {

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  selectWorkloadTypeForm = new FormGroup({
    type: new FormControl('Deployments', [Validators.required]),
  });
  @Input() clusterId: string;
  loading = true;
  allWorkloads = [];

  constructor(
    public authService: AuthService,
    private route: ActivatedRoute,
    private clusterService: ClustersService,
    private jobService: JobsService,
    private toast: ToastrService,
    private translate: TranslateService) { }

  ngOnInit(): void {
  }

  refresh(): void {
    this.loading = true;
    switch (this.selectWorkloadTypeForm.value.type) {
      case 'Deployments':
        this.clusterService.getAllDeployments(this.clusterId).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.WORKLOAD.GET-DEPLOY-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          this.allWorkloads = !!data.data?.Deployments ? data.data.Deployments : [];
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.WORKLOADS.GET-DEPLOY-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false);
        break;
      case 'Jobs':
        this.jobService.getAllJobs(this.clusterId).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.WORKLOAD.GET-JOB-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          this.allWorkloads = !!data.data?.Jobs ? data.data.Jobs : [];
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.WORKLOAD.GET-JOB-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false);
        break;
      default:
        this.allWorkloads = [];
        break;
    }

  }

}
