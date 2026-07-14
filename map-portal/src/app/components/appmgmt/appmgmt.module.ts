import { NgModule } from '@angular/core';
import { SharedModule } from '../../utils/shared/shared.module';

import { RouteComponent } from './approutes/route/route.component';
import { RouteDetailsComponent } from './approutes/route/route-details/route-details.component';
import { ConfigmapComponent } from './appvolumes/configmap/configmap.component';
import { SecretsComponent } from './appvolumes/secrets/secrets.component';
import { PvcComponent } from './appvolumes/pvc/pvc.component';
import { AppvolumesComponent } from './appvolumes/appvolumes.component';
import { CatalogsComponent } from './catalogs/catalogs.component';
import { AppcomponentsComponent } from './appcomponents/appcomponents.component';
import { AppcomponentDetailsComponent } from './appcomponent-details/appcomponent-details.component';
import { CatalogDetailsComponent } from './catalog-details/catalog-details.component';
import { ChartDetailDependencyComponent } from './catalog-details/chart-detail-dependency/chart-detail-dependency.component';
import { ChartDetailSummaryComponent } from './catalog-details/chart-detail-summary/chart-detail-summary.component';
import { ChartDetailValuesComponent } from './catalog-details/chart-detail-values/chart-detail-values.component';
import { ReleaseComponent } from './appcomponent-details/release/release.component';
import { IngressComponent } from './approutes/ingress/ingress.component';
import { ApproutesComponent } from './approutes/approutes.component';
import { StorageclassComponent } from './appvolumes/storageclass/storageclass.component';
import {AppbuildModule} from "../appbuild/appbuild.module";

@NgModule({
  declarations: [
    RouteComponent,
    RouteDetailsComponent,
    ConfigmapComponent,
    SecretsComponent,
    PvcComponent,
    AppvolumesComponent,
    CatalogsComponent,
    AppcomponentsComponent,
    AppcomponentDetailsComponent,
    CatalogDetailsComponent,
    ChartDetailDependencyComponent,
    ChartDetailSummaryComponent,
    ChartDetailValuesComponent,
    ReleaseComponent,
    IngressComponent,
    ApproutesComponent,
    StorageclassComponent
  ],
  imports: [
    SharedModule,
    AppbuildModule
  ]
})
export class AppmgmtModule { }
