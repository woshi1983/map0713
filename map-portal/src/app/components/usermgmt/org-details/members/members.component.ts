import {Component, OnInit, ViewChild, Input} from '@angular/core';
import {AuthService} from "../../../../services/auth.service";
import {ActivatedRoute} from "@angular/router";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {RolebindingService} from "../../../../services/rolebinding.service";
import {RoleBinding, User, CreateRoleBinding, UpdateRoleBinding, handleObject} from "../../../../classes";
import {ClrDatagridPagination, ClrDatagridStateInterface} from "@clr/angular";
import {FormControl, FormGroup, Validators} from '@angular/forms';
import {UsersService} from "../../../../services/users.service";

@Component({
  selector: 'app-org-members',
  templateUrl: './members.component.html',
  styleUrls: ['./members.component.css']
})
export class MembersComponent implements OnInit {

  //only admin role can handle member
  get canWrite(): boolean {
    return !this.authService.can(this.orgId, handleObject.ORGANIZATION);
  }

  @Input() orgId;

  constructor(
    public authService: AuthService,
    private rolebindingService: RolebindingService,
    private userService: UsersService,
    private route: ActivatedRoute,
    private toast: ToastrService,
    private translate: TranslateService
  ) { }

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  rolebindingList: RoleBinding[] = [];
  selected: RoleBinding;
  selectedName: string;
  usersList: User[] = [];

  loading = false;
  total = 0;
  showSearchInput = false;
  searchName: string;
  addMemberModelOpen = false;
  editMemberModelOpen = false;
  removeDialog = false;


  userInMember = false;
  userNameNotExist = false;
  alertMessage = '';
  showAlertMessage: boolean;

  basicMemberForm = new FormGroup({
    userName: new FormControl('', Validators.required),
    role: new FormControl('ADMIN', Validators.required)
  });
  editMemberForm = new FormGroup({
    userName: new FormControl('', Validators.required),
    role: new FormControl('', Validators.required)
  });

  deleteRolebindingForm = new FormGroup({
    confirm: new FormControl('', [Validators.required]),
  });


  ngOnInit(): void {
    this.showAlertMessage = localStorage.getItem('org.members.alert-show') !== 'FALSE';
  }

  refresh(state?: ClrDatagridStateInterface): void {
    this.loading = true;
    const filters = [{ key: 'organization', value: [this.orgId] }];
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.rolebindingService.getAllRolebindings(!!state?.page?.current ? state.page.current - 1 : 0, !!state?.page?.size ? state.page.size : Number(this.pagination.pageSize), filters).subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.ROLEBINDING.GET-ROLEBINDING-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.rolebindingList = !!data.data?.RoleBindings?.edge?.nodes ? data.data.RoleBindings.edge.nodes : [];
      this.selected = null;
      this.total = !!data?.data?.RoleBindings?.totalCount ? data.data.RoleBindings.totalCount : 0;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.ROLEBINDING.GET-ROLEBINDING-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false)
  }

  getAllUsers(): void {
    this.userService.getAllUserNames().subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.USER.GET-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if(!!data.data?.Users?.edge?.nodes){
        this.usersList = data.data.Users.edge.nodes;
        this.addMemberModelOpen = true;
      } else {
        this.usersList = [];
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.USER.GET-USER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false)
  }

  showAddMember(): void {
    this.getAllUsers();
  }

  checkUserNameVaild(): boolean {
    if(!!this.basicMemberForm.value.userName){
        this.userNameNotExist = false;
        this.userInMember = this.rolebindingList.some(x => x.user.name === this.basicMemberForm.value.userName.name);
        if(this.userInMember){
          this.alertMessage = 'ORGMGMTDETAILS.TABS.MEMBER-PAGE.ADDMEMBER-PAGE.USER-ERROR3'
          return true;
        }
    } else {
      this.userNameNotExist = true;
    }
    return false;
  }

  isMemberValid(): boolean {
    return !this.userInMember && !this.userNameNotExist;
  }

  doAddMemberCancel(): void {
    this.addMemberModelOpen  = false;
    this.alertMessage = '';
    this.userInMember = false;
    this.userNameNotExist = false;
    this.usersList = [];
    this.basicMemberForm.reset();
    this.basicMemberForm.patchValue({
      role: 'ADMIN'
    })
  }

  doAddMemberFinsh(): void {
    const input: CreateRoleBinding = {
      user: this.basicMemberForm.value.userName.id,
      role: this.basicMemberForm.value.role,
      organization: this.orgId
    }
     this.rolebindingService.createRolebinding(input).subscribe(data => {
       if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.ROLEBINDING.CREATE-ROLEBINDING-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
       }
       if(!!data.data?.CreateRoleBinding){
          this.toast.show(`${data.data.CreateRoleBinding.user.name} `+this.translate.instant('MESSAGE.ROLEBINDING.CREATE-ROLEBINDING-MSG'), this.translate.instant('MESSAGE.ROLEBINDING.CREATE-ROLEBINDING-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
       }
     }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.ROLEBINDING.CREATE-ROLEBINDING-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doAddMemberCancel();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    })
  }

  showUpdateMembers(): void {
    this.editMemberForm.setValue({
      userName: this.selected.user.name,
      role: this.selected.role
    })
    this.editMemberModelOpen  = true;
  }

  doUpdateMemberCancel(): void {
    this.editMemberForm.reset();
    this.editMemberModelOpen = false;
  }

  doUpdateMemberFinsh(): void {
    const input: UpdateRoleBinding = {
      id: this.selected.id,
      role: this.editMemberForm.value.role
    }
     this.rolebindingService.updateRolebinding(input).subscribe(data => {
       if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.ROLEBINDING.UPDATE-ROLEBINDING-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
        }
       if(!!data.data?.UpdateRoleBinding){
          this.toast.show(`${data.data.UpdateRoleBinding.user.name} `+this.translate.instant('MESSAGE.ROLEBINDING.CREATE-ROLEBINDING-MSG'), this.translate.instant('MESSAGE.ROLEBINDING.UPDATE-ROLEBINDING-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
       }
     }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.ROLEBINDING.UPDATE-ROLEBINDING-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doUpdateMemberCancel();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    })
  }

  showDeleteMembers(rolebinding: RoleBinding): void {
    this.selectedName = rolebinding.user.name;
    this.removeDialog = true;
  }

  deleteMemberById(id: string) {
    this.rolebindingService.deleteRolebinding(id).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.ROLEBINDING.DELETE-ROLEBINDING-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if(!!data.data?.DeleteRoleBinding){
        this.toast.show(`${data.data.DeleteRoleBinding.user.name} `+this.translate.instant('MESSAGE.ROLEBINDING.DELETE-ROLEBINDING-MSG'), this.translate.instant('MESSAGE.ROLEBINDING.DELETE-ROLEBINDING-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.ROLEBINDING.DELETE-ROLEBINDING-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.removeDialog = false;
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  disableAlertMessageShow(): void {
    this.showAlertMessage = false;
    localStorage.setItem('org.members.alert-show', 'FALSE');
  }

}
