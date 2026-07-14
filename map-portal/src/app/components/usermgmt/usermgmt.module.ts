import { NgModule } from '@angular/core';
import { SharedModule } from '../../utils/shared/shared.module';

import { OrganizationComponent } from './organization/organization.component';
import { UsersComponent } from './users/users.component';
import { OrgDetailsComponent } from './org-details/org-details.component';
import { MembersComponent } from './org-details/members/members.component';
import { QuotaComponent } from './org-details/quota/quota.component';
import { AuditlogComponent } from './auditlog/auditlog.component';

@NgModule({
  declarations: [
    OrganizationComponent,
    UsersComponent,
    OrgDetailsComponent,
    MembersComponent,
    QuotaComponent,
    AuditlogComponent
  ],
  imports: [
    SharedModule
  ]
})
export class UserMgmtModule { }
