import { NgModule } from '@angular/core';
import { SharedModule } from '../../utils/shared/shared.module';
import { SummaryComponent } from './summary/summary.component';


@NgModule({
  declarations: [
    SummaryComponent
  ],
  imports: [
    SharedModule
  ]
})
export class OverviewModule { }
