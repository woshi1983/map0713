import { NgModule } from '@angular/core';
import { SharedModule } from '../../utils/shared/shared.module';
import { SharedComponentsModule } from '../shared-components/shared-components.module';
import { BinaryImageComponent } from './binary/binary-image/binary-image.component';
import { DockerfileImageComponent } from './dockerfile/dockerfile-image/dockerfile-image.component';
import { SourceImageComponent } from './sources/source-image/source-image.component';
import { BinaryBuildComponent } from './binary/binary-build/binary-build.component';
import { BinaryImageDetailsComponent } from './binary/build-image-details/binary-image-details.component';
import { DockerfileImageDetailsComponent } from './dockerfile/dockerfile-image-details/dockerfile-image-details.component';
import { ClrLoadingModule } from '@clr/angular';
import { DockerfileBuildComponent } from './dockerfile/dockerfile-build/dockerfile-build.component';
import { SourceImageDetailsComponent } from './sources/source-image-details/source-image-details.component';
import { SourceBuildComponent } from './sources/source-build/source-build.component';
import { BuildLoggingComponent } from './build-logging/build-logging.component';

@NgModule({
  declarations: [
    BinaryImageComponent,
    DockerfileImageComponent,
    SourceImageComponent,
    BinaryBuildComponent,
    BinaryImageDetailsComponent,
    DockerfileImageDetailsComponent,
    DockerfileBuildComponent,
    SourceImageDetailsComponent,
    SourceBuildComponent,
    BuildLoggingComponent
  ],
  imports: [
    SharedModule,
    SharedComponentsModule,
    ClrLoadingModule
  ],
  exports: [
    BuildLoggingComponent
  ]
})
export class AppbuildModule { }
