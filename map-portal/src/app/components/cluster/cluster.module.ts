import { NgModule } from '@angular/core';
import { SharedModule } from '../../utils/shared/shared.module';

import { ManagementComponent } from './management/management.component';

import { ClusterdetailComponent } from './clusterdetail/clusterdetail.component';
import { SummaryComponent } from './clusterdetail/summary/summary.component';
import { ComputeComponent } from './clusterdetail/compute/compute.component';
import { BackupsRestoresComponent } from './clusterdetail/backups-restores/backups-restores.component';
import { DataProtectionComponent } from './clusterdetail/data-protection/data-protection.component';
import { NamespacesComponent } from './clusterdetail/namespaces/namespaces.component';
import { ClusterWorkloadsComponent } from './clusterdetail/cluster-workloads/cluster-workloads.component';
import { OperationComponent } from './operation/operation.component';
import { AppInstancesComponent } from './clusterdetail/app-instances/app-instances.component';
import { CronTooltipComponent } from './clusterdetail/backups-restores/cron-tooltip/cron-tooltip.component';
import {CommonModule} from "@angular/common";


@NgModule({
  declarations: [
    ManagementComponent,
    ClusterdetailComponent,
    SummaryComponent,
    ComputeComponent,
    BackupsRestoresComponent,
    DataProtectionComponent,
    NamespacesComponent,
    ClusterWorkloadsComponent,
    OperationComponent,
    AppInstancesComponent,
    CronTooltipComponent
  ],
    imports: [
        SharedModule,
        CommonModule
    ]
})
export class ClusterModule { }

