import { NgModule } from '@angular/core';
import { MarkdownModule } from 'ngx-markdown';
import { CommonModule, registerLocaleData } from '@angular/common';
import { ClarityModule, ClrIconModule, ClrDatagridModule } from '@clr/angular';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { FlexLayoutModule } from '@angular/flex-layout';
import { ToastContainerModule, ToastrModule } from 'ngx-toastr';
import { AppRoutingModule } from '../../app-routing.module';
import { TranslateLoader, TranslateModule } from '@ngx-translate/core';
import { TranslateHttpLoader } from '@ngx-translate/http-loader';
import { HttpClient, HttpClientModule } from '@angular/common/http';
import { CookieModule, CookieService } from 'ngx-cookie';
import { GraphQLModule } from '../../graphql.module';
import { environment } from '../../../environments/environment';

import { AuthService } from '../../services/auth.service';
import { BnNgIdleService } from 'bn-ng-idle';

import { CustomValidatorDirective } from '../validators/custom-validator.directive';
import { NotFoundComponent } from './not-found/not-found.component';
import { NoPermissionComponent } from './no-permission/no-permission.component';
import { MapToast } from './toast/toast.component';
import { MonacoEditorModule } from '@jinmutech/ngx-monaco-editor';

import { NgxEchartsModule } from 'ngx-echarts';
import { DatetimeTransitionPipe } from './pipes/datetime-transition.pipe';
import { FeaturesToggleDirective } from './directive/features-directive.directive';

import zh_cn from '@angular/common/locales/zh-Hans';
import zh_tw from '@angular/common/locales/zh-Hans-HK';
// add locale data for supported languages  ['en-us', 'zh-cn', 'zh-tw'];
// en-us defaulted supported
registerLocaleData(zh_cn, 'zh-cn');
registerLocaleData(zh_tw, 'zh-tw');

@NgModule({
  declarations: [
    CustomValidatorDirective,
    NotFoundComponent,
    MapToast,
    NoPermissionComponent,
    DatetimeTransitionPipe,
    FeaturesToggleDirective
  ],
  imports: [
    CommonModule,
    MarkdownModule.forRoot(),
    ClarityModule,
    ClrDatagridModule,
    BrowserModule,
    BrowserAnimationsModule,
    FormsModule,
    ReactiveFormsModule,
    FlexLayoutModule,
    HttpClientModule,
    GraphQLModule,
    AppRoutingModule,
    TranslateModule.forRoot({
      loader: {
        provide: TranslateLoader,
        useFactory: GeneralTranslatorLoader,
        deps: [HttpClient]
      }
    }),
    ToastrModule.forRoot({ toastComponent: MapToast, preventDuplicates: true, positionClass: 'my-toast-top-right' }),
    ToastContainerModule,
    CookieModule.forRoot(),
    ClrIconModule,
    MonacoEditorModule.forRoot(),
    NgxEchartsModule.forRoot({
      echarts: () => import('echarts')
    })
  ],
  entryComponents: [MapToast],
  exports: [
    CommonModule,
    ClarityModule,
    ClrDatagridModule,
    ClrIconModule,
    BrowserAnimationsModule,
    FormsModule,
    ReactiveFormsModule,
    FlexLayoutModule,
    HttpClientModule,
    GraphQLModule,
    AppRoutingModule,
    TranslateModule,
    CookieModule,
    MarkdownModule,
    MonacoEditorModule,
    NgxEchartsModule,
    DatetimeTransitionPipe,
    FeaturesToggleDirective
  ],
  providers: [
    CookieService,
    AuthService,
    BnNgIdleService
  ]
})
export class SharedModule { }

export function GeneralTranslatorLoader (http: HttpClient): any {
  const langFile = environment.assetsPath + 'il8n/';
  return new TranslateHttpLoader(http, langFile);
}
