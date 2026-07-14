import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { AppComponent } from './app.component';

import { BaseModule } from './components/base/base.module';
import { SharedModule } from './utils/shared/shared.module';
import { ClusterModule } from './components/cluster/cluster.module';
import { SystemModule } from './components/system/system.module';
import { AuthModule } from './components/auth/auth.module';
import { AppmgmtModule } from './components/appmgmt/appmgmt.module';
import { AppbuildModule } from './components/appbuild/appbuild.module';
import { UserMgmtModule } from './components/usermgmt/usermgmt.module';
import { OverviewModule } from './components/overview/overview.module';
import { SharedComponentsModule } from './components/shared-components/shared-components.module';

@NgModule({
  declarations: [
    AppComponent
  ],
  imports: [
    BrowserModule,
    SharedModule,
    ClusterModule,
    SystemModule,
    AuthModule,
    AppmgmtModule,
    AppbuildModule,
    BaseModule,
    UserMgmtModule,
    OverviewModule,
    SharedComponentsModule
  ],
  providers: [
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
