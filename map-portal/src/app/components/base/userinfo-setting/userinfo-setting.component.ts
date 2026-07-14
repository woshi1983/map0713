import {Component, OnInit} from '@angular/core';
import {AuthService} from "../../../services/auth.service";
import {UsersService} from "../../../services/users.service";
import {FormControl, FormGroup, Validators} from "@angular/forms";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {UpdateUser, User} from "../../../classes";

@Component({
  selector: 'app-userinfo-setting',
  templateUrl: './userinfo-setting.component.html',
  styleUrls: ['./userinfo-setting.component.css']
})
export class UserinfoSettingComponent implements OnInit {

  opened: boolean = false;
  user: User;
  currentUserId: string;

  constructor(
    public authService: AuthService,
    private usersService: UsersService,
    private toast: ToastrService,
    private translate: TranslateService
  ) { }

  editUserForm = new FormGroup({
    name: new FormControl('', [Validators.required]),
    displayName: new FormControl('', [Validators.required, Validators.maxLength(24)])
  });

  ngOnInit(): void {
    this.currentUserId = this.authService.loginUserId;
  }

  open(): void {
    this.usersService.getUserInfoById(this.currentUserId).subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.USER.GET-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if(!!data.data.User){
        this.user = data.data.User;
        this.editUserForm.setValue({
          name: data.data.User.name,
          displayName: data.data.User.displayName
        })
        this.opened = true;
      } else {
        this.user = null;
        this.editUserForm.reset();
        this.opened = false;
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.USER.GET-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
  }

  cancel(): void {
    this.user = null;
    this.editUserForm.reset();
    this.opened = false;
  }

  doUpdateUserInfo(): void {
    let input: UpdateUser = {
      id: this.currentUserId,
      displayName: this.editUserForm.value.displayName
    };
    this.usersService.updateUserInfo(input).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.USER.UPDATE-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
     if(!!data.data?.UpdateUser){
       this.toast.show(`${data.data.UpdateUser.name} `+this.translate.instant('MESSAGE.USER.UPDATE-USER-MSG'), this.translate.instant('MESSAGE.USER.UPDATE-USER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000});
       this.authService.setUserDisplayName(data.data.UpdateUser.displayName);
     }
     }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.USER.UPDATE-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.cancel();
    })
  }

}
