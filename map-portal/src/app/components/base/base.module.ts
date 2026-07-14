import { NgModule } from '@angular/core';
import { SharedModule } from '../../utils/shared/shared.module';

import { HomeComponent } from './home/home.component';
import { ApplevelMessageComponent } from './applevel-message/applevel-message.component';
import { NavigatorComponent } from './navigator/navigator.component';
import { ToastrModule } from 'ngx-toastr';
import { AboutDialogComponent } from './about-dialog/about-dialog.component';
import { ChangePasswordComponent } from './change-password/change-password.component';
import { UserinfoSettingComponent } from './userinfo-setting/userinfo-setting.component';
import { SubscribeMapdLogsComponent } from './subscribe-mapd-logs/subscribe-mapd-logs.component';
import { UserNoOrgComponent } from './user-no-org/user-no-org.component';

@NgModule({
  declarations: [
    HomeComponent,
    ApplevelMessageComponent,
    NavigatorComponent,
    AboutDialogComponent,
    ChangePasswordComponent,
    UserinfoSettingComponent,
    SubscribeMapdLogsComponent,
    UserNoOrgComponent
  ],
  imports: [
    SharedModule,
    ToastrModule
  ],
  exports: [
    ApplevelMessageComponent,
    NavigatorComponent,
    AboutDialogComponent,
    ChangePasswordComponent,
    UserinfoSettingComponent
  ]
})
export class BaseModule { }
