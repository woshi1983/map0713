import { Component, OnInit, Output, EventEmitter } from '@angular/core';
import {TranslateService} from '@ngx-translate/core';
import {AuthService} from '../../../services/auth.service';
import {Organization, ModalEvent} from "../../../classes";
import {EventbusService} from "../../../services/eventbus.service";
import {OrganizationService} from "../../../services/organization.service";
import {Router} from "@angular/router";
import {LANG_LOCALSTORAGE_KEY, DeFaultLang, languageNames} from "../../../utils/static-entries/shared.const";
import {ToastrService} from "ngx-toastr";
import { environment} from '../../../../environments/environment';

@Component({
  selector: 'app-navigator',
  templateUrl: './navigator.component.html',
  styleUrls: ['./navigator.component.css']
})
export class NavigatorComponent implements OnInit {

  @Output() showDialogModalAction = new EventEmitter<ModalEvent>();

  constructor(
    private toast: ToastrService,
    private translate: TranslateService,
    public authService: AuthService,
    public orgService: OrganizationService,
    private router: Router,
    private eventbusService: EventbusService
  ) { }

  logoImage = environment.assetsPath + 'images/logo.png';

  selectedLang: string;
  languageNames = languageNames;
  orgList: Organization[] = [];

  //trigger by org change
  refreshTriggerdByEvent(): void {
    this.eventbusService.refreshUserOrgEmitter.subscribe(
      data => {
        this.getUserAllOrg();
      }
    );
  }

  useLanguage(language: string): void {
    this.selectedLang = language;
    localStorage.setItem(LANG_LOCALSTORAGE_KEY, language);
    this.translate.use(language).subscribe(
      // can do something when lang changed
    );
  }

  matchLang(lang: string): boolean {
    return lang.trim() === this.selectedLang;
  }

  public get langShowName(): string {
    return this.languageNames[this.selectedLang];
  }

  // Open about dialog
  openAboutModal(): void {
    this.showDialogModalAction.emit({
        modalName: 'ABOUT',
        modalFlag: true
    });
  }

  // Open change password dialog
  openChangePasswordModal(): void {
    this.showDialogModalAction.emit({
        modalName: 'CHANGE_PWD',
        modalFlag: true
    });
  }

  // Open change user info dialog
  openUserInfoModal(): void {
    this.showDialogModalAction.emit({
        modalName: 'USER_PROFILE',
        modalFlag: true
    });
  }

  // Open mapd logs dialog
  openLogsModal(): void {
    this.showDialogModalAction.emit({
        modalName: 'MAPD_LOGS',
        modalFlag: true
    });
  }

  // fix click home when in org-init page
  get getRouteUrl(): string {
    return this.router.url;
  }

  getUserAllOrg(): void {
    this.orgService.getAllOrgNames().subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.ORG.GET-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data?.data?.Organizations?.edge?.nodes) {
        const tempList = data.data.Organizations.edge.nodes;
        this.orgList = tempList.sort((n1, n2) => {
              if(n1.name > n2.name){return 1;}
              if(n1.name < n2.name){return -1;}
              return 0
        });
        if(this.orgList.findIndex(x => x.id === this.orgService.getSelectedOrgId) === -1){
          // selected org id not exist or don't aloow by user, reset to first org
          this.changeOrg(this.orgList[0]);
        }
      } else {
        this.orgList = [];
        if (!data.errors){
          //request success without error, user not in any org
          this.router.navigate(['/portal/logout'], {replaceUrl: true});
          this.toast.show( this.translate.instant('MESSAGE.SHARED.GETSUCCESS')+this.translate.instant('MESSAGE.ORG.NOORG'), this.translate.instant('MESSAGE.ORG.GET-ORG-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
        }
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.ORG.GET-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  changeOrg(org: Organization): void {
    if(org.id != this.orgService.getSelectedOrgId){
      this.orgService.setSelectOrg(org.id, org.name);
      window.setTimeout(function(){window.location.href='/portal/cluster/management'}, 1000);
    }
  }

  matchOrg(org: Organization): boolean {
    return org.id === this.orgService.getSelectedOrgId;
  }

  ngOnInit(): void {
    this.selectedLang = localStorage.getItem(LANG_LOCALSTORAGE_KEY);
    if (this.selectedLang == null) {
      // setup default lang
      this.selectedLang = DeFaultLang;
    }
    this.useLanguage(this.selectedLang);
    //if user not in logout and no-org page & get all org
    if(this.getRouteUrl !== '/portal/logout' && this.getRouteUrl !== '/portal/no-org'){
      this.getUserAllOrg();
    }
    this.refreshTriggerdByEvent();
  }

}
