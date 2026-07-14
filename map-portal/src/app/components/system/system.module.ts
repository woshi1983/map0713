import { NgModule } from '@angular/core';
import { SharedModule } from '../../utils/shared/shared.module';
import { SharedComponentsModule } from '../shared-components/shared-components.module';
import { RepositorysComponent } from './repositorys/repositorys.component';
import { ProviderComponent } from './provider/provider.component';
import { ProviderdetailComponent } from './providerdetail/providerdetail.component';
import { SupervisorNamespaceComponent } from './providerdetail/namespace/namespace.component';
import { CatrustComponent } from './providerdetail/catrust/catrust.component';
import { LoadbalancerComponent } from './loadbalancer/loadbalancer.component';
import { StandardPackageRepositoryComponent } from './standard-package-repository/standard-package-repository.component';

@NgModule({
  declarations: [
    RepositorysComponent,
    ProviderComponent,
    ProviderdetailComponent,
    SupervisorNamespaceComponent,
    CatrustComponent,
    LoadbalancerComponent,
    StandardPackageRepositoryComponent
  ],
  imports: [
    SharedModule,
    SharedComponentsModule
  ]
})
export class SystemModule { }
