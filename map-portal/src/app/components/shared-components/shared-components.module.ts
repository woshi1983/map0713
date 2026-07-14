import { NgModule } from '@angular/core';
import { SharedModule } from '../../utils/shared/shared.module';
import { DeleteObjectComponent } from './delete-object/delete-object.component';
import { ForceDeleteObjectComponent } from './force-delete-object/force-delete-object.component';

@NgModule({
  declarations: [
    DeleteObjectComponent,
    ForceDeleteObjectComponent
  ],
  imports: [
    SharedModule
  ],
  exports: [
    DeleteObjectComponent,
    ForceDeleteObjectComponent
  ]
})
export class SharedComponentsModule { }
