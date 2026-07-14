import { NgModule } from '@angular/core';
import { SharedModule } from '../../utils/shared/shared.module';

import { LogoutComponent } from './logout/logout.component';
import {BaseModule} from "../base/base.module";

@NgModule({
    declarations: [
        LogoutComponent
    ],
    exports: [
    ],
    imports: [
        SharedModule,
        BaseModule
    ]
})
export class AuthModule { }
