import {Component, ElementRef, OnInit, ViewChild} from '@angular/core';
import {AuthService} from "../../../services/auth.service";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {UsersService} from "../../../services/users.service";
import {CommonService} from "../../../services/common.service";
import {ClrDatagridPagination, ClrDatagridStateInterface, ClrWizard} from "@clr/angular";
import {CreateUser, RoleBinding, User} from "../../../classes";
import {CustomValidatorDirective, nameValidator} from "../../../utils/validators/custom-validator.directive";
import {FormControl, FormGroup, Validators} from "@angular/forms";

@Component({
  selector: 'app-users',
  templateUrl: './users.component.html',
  styleUrls: ['./users.component.css']
})
export class UsersComponent implements OnInit {

  constructor(
    public authService: AuthService,
    private userService: UsersService,
    private commonService: CommonService,
    private toast: ToastrService,
    private translate: TranslateService
  ) { }

  localUserDisabled: boolean = false;

  @ViewChild('search', { static: false }) search: ElementRef<HTMLInputElement>;
  @ViewChild('pagination') pagination: ClrDatagridPagination;
  usersList: User[] = [];
  selected: User;
  allRolebindings: RoleBinding[] = [];
  randomPassword: string;

  loading = false;
  total = 0;
  showSearch = true;
  searchName: string;

  showCreateOpen = false;
  removeDialog = false;
  showRandomPass = false;
  showSetupSuperadmin = false;
  showResetPasswd = false;
  showRolebindingModel = false;

  @ViewChild('wizarduser', { static: false }) wizarduser: ClrWizard;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  readonly passwordPattern = CustomValidatorDirective.passwordPattern;
  basicUserForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(24), Validators.pattern(this.restricNamePattern), nameValidator(['admin'])]),
    domain: new FormControl('@tanzu.local'),
    displayName: new FormControl('', [Validators.required, Validators.maxLength(24)]),
    randomPass: new FormControl(false),
    password: new FormControl('', [Validators.required, Validators.pattern(this.passwordPattern)]),
    confirmPassword: new FormControl('',[Validators.required]),
    superAdmin: new FormControl(false)
  });
  resetUserPasswordForm =  new FormGroup({
    newPassword: new FormControl('', [Validators.required, Validators.pattern(this.passwordPattern)]),
    reNewPassword: new FormControl('', Validators.required)
  })

  deleteUserForm = new FormGroup({
    confirm: new FormControl('', [Validators.required]),
  });

  ngOnInit(): void {
    this.checkLocalUserDisable();
  }

  refresh(state?: ClrDatagridStateInterface): void {
    this.loading = true;
    const filters = [];
    if (!!this.searchName) {
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.userService.getAllUsers(!!state?.page?.current ? state.page.current - 1 : 0, !!state?.page?.size ? state.page.size : Number(this.pagination.pageSize), filters).subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.USER.GET-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.usersList = !!data.data?.Users?.edge?.nodes ? data.data.Users.edge.nodes : [];
      this.selected = null;
      this.total = !!data?.data?.Users?.totalCount ? data.data.Users.totalCount : 0;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.USER.GET-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false)
  }

  checkLocalUserDisable(): void {
    this.commonService.getConfiguration().subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.USER.GET-LOCAL-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.localUserDisabled = !!data.data?.Metrics?.localUsersDisabled ? data.data.Metrics.localUsersDisabled : false;
    })
  }

  showUserRolebinding(rolebindingList: RoleBinding[]): void {
    this.allRolebindings = rolebindingList;
    this.showRolebindingModel = true;
  }

  showCreateUser(): void {
    this.showCreateOpen = true;
  }

  checkPasswdValidation(pass: string, repass: string): boolean {
    if(!!repass){
      return pass === repass;
    }
    return true;
  }

  setDefaultDisplayName($event: any): void {
    this.basicUserForm.patchValue({
      displayName: $event.target.value
    });
  }

  updatePasswordVaildtor(random: boolean): void {
    if(random){
      this.basicUserForm.controls['password'].clearValidators();
      this.basicUserForm.controls['password'].updateValueAndValidity();
      this.basicUserForm.controls['confirmPassword'].clearValidators();
      this.basicUserForm.controls['confirmPassword'].updateValueAndValidity();
    } else {
      this.basicUserForm.controls['password'].setValidators([Validators.required, Validators.pattern(this.passwordPattern)]);
      this.basicUserForm.controls['password'].updateValueAndValidity();
      this.basicUserForm.controls['confirmPassword'].setValidators(Validators.required);
      this.basicUserForm.controls['confirmPassword'].updateValueAndValidity();
    }
  }

  doCreateUserFinsh(): void {
     const input: CreateUser = {
       name: this.basicUserForm.value.name,
       displayName: this.basicUserForm.value.displayName,
       admin: this.basicUserForm.value.superAdmin
     }
     if(!this.basicUserForm.value.randomPass){
       input.password = this.basicUserForm.value.confirmPassword
     }
     this.userService.createUser(input).subscribe(data => {
       if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.USER.CREATE-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
       }
       if(!!data.data?.CreateUser){
          this.toast.show(`${data.data.CreateUser.name} `+this.translate.instant('MESSAGE.USER.CREATE-USER-MSG'), this.translate.instant('MESSAGE.USER.CREATE-USER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          if(this.basicUserForm.value.randomPass){
            this.randomPassword = data.data?.CreateUser.password;
            this.showRandomPass = true;
          } else {
            this.randomPassword = '';
            this.showRandomPass = false;
          }
       }
     }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.USER.CREATE-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doCreateUserReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    })
  }

  doCreateUserCancel(): void {
    this.doCreateUserReset();
  }

  doCreateUserReset(): void {
    this.wizarduser.reset();
    this.basicUserForm.reset();
    this.basicUserForm.patchValue({
      randomPass: false,
      superAdmin: false
    })
    this.showCreateOpen = false;
  }

  showSetSuperadmin(): void {
    this.showSetupSuperadmin = true;
  }

  updateUserAdmin(user: User): void {
    this.userService.updateUserAdmin(user.id, !user.admin).subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors) {
          if (!!each.message) {
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.USER.UPDATE-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
          }
        }
      }
      if(!!data.data?.UpdateUser){
        this.toast.show(`${data.data.UpdateUser.name} `+this.translate.instant('MESSAGE.USER.UPDATE-USER-MSG'), this.translate.instant('MESSAGE.USER.UPDATE-USER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.USER.UPDATE-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.refresh();
      this.showSetupSuperadmin = false;
      this.selected = null;
    })
  }

  showResetUserPassword(): void {
    this.showResetPasswd = true;
  }

  doResetUserPassword(): void {
    this.userService.resetUserPasswd(this.selected.id, this.resetUserPasswordForm.value.reNewPassword).subscribe(data => {
      if(!!data.errors){
        for (const each of data.errors) {
          if (!!each.message) {
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.USER.UPDATE-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
          }
        }
      }
      if(!!data.data?.UpdateUser){
        this.toast.show(`${data.data.UpdateUser.name} `+this.translate.instant('MESSAGE.USER.UPDATE-USER-MSG'), this.translate.instant('MESSAGE.USER.UPDATE-USER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.USER.UPDATE-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.showResetPasswd = false;
      this.resetUserPasswordForm.reset();
    })

  }

  showDeleteUser(): void {
    this.removeDialog = true;
  }

  deleteUserById(id: string) {
    this.userService.deleteUser(id).subscribe(data => {
      if(!!data.errors){
        for (const each of data.errors) {
          if (!!each.message) {
             this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.USER.DELETE-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
          }
        }
      }
      if(!!data.data?.DeleteUser){
        this.toast.show(`${data.data.DeleteUser.name} `+this.translate.instant('MESSAGE.USER.DELETE-USER-MSG'), this.translate.instant('MESSAGE.USER.DELETE-USER-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
      this.removeDialog = false;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.USER.DELETE-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

}
