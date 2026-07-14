import { Component, OnInit, ViewChild } from '@angular/core';
import {FormControl, FormGroup, NgForm, Validators} from '@angular/forms';
import {AuthService} from "../../../services/auth.service";
import {UsersService} from "../../../services/users.service";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {Router} from "@angular/router";

@Component({
  selector: 'app-change-password',
  templateUrl: './change-password.component.html',
  styleUrls: ['./change-password.component.css']
})
export class ChangePasswordComponent implements OnInit {

  opened: boolean;
  passwdForm = new FormGroup({
    oldPassword: new FormControl('', [Validators.required]),
    newPassword: new FormControl('', [Validators.required]),
    reNewPassword: new FormControl('', [Validators.required]),
  });
  private validationStateMap: any = {
    "newPassword": true,
    "reNewPassword": true
  };

  constructor(
    public authService: AuthService,
    private usersService: UsersService,
    private toast: ToastrService,
    private translate: TranslateService,
    private router: Router
  ) { }

  ngOnInit(): void {
  }


  getValidationState(key: string): boolean {
    return this.validationStateMap[key];
  }

  handleValidation(key: string, flag: boolean): void {
      if (flag) {
          // Checking
          let cont = this.passwdForm.controls[key];
          if (cont) {
              this.validationStateMap[key] = cont.valid;
              if (cont.valid) {
                  if (key === "reNewPassword" || key === "newPassword") {
                      let cpKey = key === "reNewPassword" ? "newPassword" : "reNewPassword";
                      let compareCont = this.passwdForm.controls[cpKey];
                      if (compareCont && compareCont.valid) {
                          this.validationStateMap["reNewPassword"] = cont.value === compareCont.value;
                      }
                  }
              }
          }
      } else {
          // Reset
          this.validationStateMap[key] = true;
      }
  }

  // If form is valid
  public get isValid(): boolean {
    return this.passwdForm.valid && this.passwdForm.value.newPassword === this.passwdForm.value.reNewPassword
  }

  open(): void {
    this.validationStateMap = {
      "newPassword": true,
      "reNewPassword": true
    };
    this.opened = true;
  }

  cancel(): void {
    this.opened = false;
    this.passwdForm.reset();
  }

  doUpdateUserPassword(): void {
    const userId = this.authService.loginUserId;
    this.usersService.updateUserPasswd(userId, this.passwdForm.value.oldPassword, this.passwdForm.value.reNewPassword).subscribe(data => {
      if(!!data.errors){
        for (const each of data.errors) {
          if (!!each.message) {
             if(each.message === 'insufficient privileges'){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+this.translate.instant('MESSAGE.USER.PASSWDERROR'), this.translate.instant('MESSAGE.USER.UPDATE-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             } else {
               this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.USER.UPDATE-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
        }
      }
      if(!!data.data?.UpdateUser){
        this.toast.show(`${data.data.UpdateUser.name} `+this.translate.instant('MESSAGE.USER.UPDATE-USER-MSG'), this.translate.instant('MESSAGE.USER.UPDATE-USER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
        this.router.navigate(['/portal/logout']);
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.USER.UPDATE-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
    this.cancel();
  }

}
