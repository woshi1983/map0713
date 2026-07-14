import {Component, OnInit, ViewChild} from '@angular/core';
import {AuthService} from "../../../services/auth.service";
import {ClrDatagridStateInterface, ClrDatagridPagination} from "@clr/angular";
import {CookieService} from "ngx-cookie";
import {OrganizationService} from "../../../services/organization.service";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {AuditLog, Organization} from "../../../classes";
import {AuditlogService} from "../../../services/auditlog.service";
import { saveAs } from 'file-saver';
import {timeAgo} from '../../../utils/operator';

@Component({
  selector: 'app-auditlog',
  templateUrl: './auditlog.component.html',
  styleUrls: ['./auditlog.component.css']
})
export class AuditlogComponent implements OnInit {

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  auditLogs: AuditLog[] = [];
  orgList: Organization[] = [];
  time: number = 30;
  timeUnit: string = 'minutes';
  loading = false;
  total = 0;
  pageState: ClrDatagridStateInterface;

  constructor(
    private cookie: CookieService,
    public orgService: OrganizationService,
    public authService: AuthService,
    private auditlogService: AuditlogService,
    private toast: ToastrService,
    private translate: TranslateService) { }

  ngOnInit(): void {
    this.getAllOrgs();
  }

  refresh(state?: ClrDatagridStateInterface): void {
    this.loading = true;
    this.pageState = state;
    let filterTime = timeAgo(!!this.time ? this.time : 30, !!this.timeUnit ? this.timeUnit : 'minutes');
    const filters = [{key: 'from_time', value: [filterTime[0]]}, {key: 'until_time', value: [filterTime[1]]}];
    if(!!state?.filters){
      for (const filter of state.filters){
        if(filter.property === "organization.name"){
          const org = this.orgList.find(x => x.name === filter.value)
          if(!!org?.id){
            filters.push({key: 'organization', value: [org.id]})
          } else {
            filters.push({key: 'organization', value: ['null']})
          }
        } else {
          filters.push({key: filter.property, value: [filter.value]})
        }
      }
    }
    this.auditlogService.getAllAuditlogs(!!state?.page?.current ? state.page.current - 1 : 0, !!state?.page?.size ? state.page.size : Number(this.pagination.pageSize), filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message && each.message !== 'object not found'){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.AUDITLOG.GET-LOG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.auditLogs = !!data.data?.AuditLogs?.edge?.nodes ? data.data.AuditLogs.edge.nodes : [];
      this.total = !!data.data?.AuditLogs?.totalCount ? data.data.AuditLogs.totalCount : 0;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.AUDITLOG.GET-LOG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false);
  }

  getAllOrgs(): void {
    this.orgService.getAllOrgNames().subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.ORG.GET-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.orgList = !!data.data?.Organizations?.edge?.nodes ? data.data.Organizations.edge.nodes : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.ORG.GET-ORG-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}) );
  }

  ConvertToCSV(objArray, headerList) {
   let array = typeof objArray != 'object' ? JSON.parse(objArray) : objArray;
   let str = '';
   let row = 'S.No,';
   for (let index in headerList) {
    row += headerList[index] + ',';
   }
   row = row.slice(0, -1);
   str += row + '\r\n';
   for (let i = 0; i < array.length; i++) {
    let line = (i+1)+'';
    for (const index in headerList) {
      const head = headerList[index];
      if (head === 'organization') {
        if (array[i][head]) {
          line += ',' + array[i][head].name
        } else {
          line += ',' + '-'
        }
      } else {
        line += ',' + array[i][head];
      }
    }
    str += line + '\r\n';
   }
   return str;
  }

  exportAuditlogsToCSV(): void {
    let csvData = this.ConvertToCSV(this.auditLogs, ['userName','objectName', 'objectType', 'action', 'organization', 'createdAt']);
    const blob = new Blob(['\ufeff' + csvData], { type: 'text/csv;charset=utf-8;' });
    saveAs(blob, 'map-auditlog.csv');
  }

  exportAuditlogsToJSON(): void {
    const logsString = JSON.stringify(this.auditLogs);
    const blob = new Blob([logsString], { type: 'text/json;charset=utf-8;' });
    saveAs(blob, 'map-auditlog.json');
  }

}
