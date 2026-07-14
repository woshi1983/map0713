import { Component, OnInit, ViewChild } from '@angular/core';
import { AuthService } from "../../../services/auth.service";
import { ModalEvent } from "../../../classes";
import { AboutDialogComponent } from "../about-dialog/about-dialog.component";
import { ChangePasswordComponent } from "../change-password/change-password.component";
import { UserinfoSettingComponent } from "../userinfo-setting/userinfo-setting.component";


@Component({
  selector: 'app-no-org',
  templateUrl: './user-no-org.component.html',
  styleUrls: ['./user-no-org.component.css']
})
export class UserNoOrgComponent implements OnInit {

  constructor(
    public authService: AuthService
  ) {
  }


  @ViewChild(AboutDialogComponent)
  aboutDialogView: AboutDialogComponent;

  @ViewChild(ChangePasswordComponent)
  changePwdView: ChangePasswordComponent;

  @ViewChild(UserinfoSettingComponent)
  accountSettingView: UserinfoSettingComponent;

  ngOnInit(): void {
  }

  // Open modal dialog
  openModal(event: ModalEvent): void {
    switch (event.modalName) {
      case 'USER_PROFILE':
          this.accountSettingView.open();
          break;
      case 'CHANGE_PWD':
          this.changePwdView.open();
          break;
      case 'ABOUT':
          this.aboutDialogView.open();
          break;
      default:
          break;
    }
  }

}
