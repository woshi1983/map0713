import {Component, OnInit, ViewChild} from '@angular/core';
import { map } from 'rxjs/operators';
import { Observable } from 'rxjs';
import {ActivatedRoute, Router} from "@angular/router";
import {
  Application,
  ChartDetails,
  ChartVersion,
  CreateComponent,
  CustomComponent,
  handleObject
} from "../../../classes";
import {AuthService} from "../../../services/auth.service";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {RepositorysService} from "../../../services/repositorys.service";
import { environment} from '../../../../environments/environment';
import {FormControl, FormGroup, Validators} from "@angular/forms";
import {ClrWizard} from "@clr/angular";
import {
  checkOneYamlFormat,
  CustomValidatorDirective,
  nameValidator
} from "../../../utils/validators/custom-validator.directive";
import {ApplicationsService} from "../../../services/applications.service";
import {OrganizationService} from "../../../services/organization.service";
import {ComponentService} from "../../../services/component.service";
import YAML from 'yaml';


@Component({
  selector: 'app-catalog-details',
  templateUrl: './catalog-details.component.html',
  styleUrls: ['./catalog-details.component.css']
})
export class CatalogDetailsComponent implements OnInit {

  //admin and user role can handle component
  get canWriteComponent(): boolean {
    return this.authService.can(this.orgId, handleObject.COMPONENT);
  }

  editorOptions = {theme: 'vs', language: 'yaml', automaticLayout: true};
  originValues: string;
  newValues: string;

  state$: Observable<any>;
  chartVersionList: ChartVersion[];
  chartVersionListSorted: ChartVersion[];
  chartDetails: ChartDetails;
  appList: Application[] = [];
  orgId: string;
  componentList: CustomComponent[] = [];

  loading = false;

  helmImage: string;
  defaultChartImage: string;

  showComponentModel = false;
  showCheckedAlert = false;
  checkFormateOk = false;
  parseError = '';
  componentNameExist = false;

  @ViewChild('wizardhelm') wizardnewhelm: ClrWizard;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  basicHelmComponentForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(48), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    applicationId: new FormControl('', Validators.required)
  });

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    public authService: AuthService,
    private toast: ToastrService,
    private translate: TranslateService,
    private repositoryService: RepositorysService,
    private applicationService: ApplicationsService,
    private orgService: OrganizationService,
    private cptService: ComponentService
  ) {
    this.helmImage = environment.assetsPath + 'images/helm.svg';
    this.defaultChartImage = environment.assetsPath + 'images/other.png';
    this.getParamFromRoute();
    this.orgId = this.orgService.getSelectedOrgId;
  }

  chartVersionForm = new FormGroup({
    version: new FormControl(''),
  });

  ngOnInit(): void {
    this.chartVersionListSorted = JSON.parse(localStorage.getItem('catalogs.details.chart_versions'));
    this.chartVersionForm.setValue({
      version: this.chartVersionListSorted[0].version
    })
    this.getChartDetails({target: {value: this.chartVersionListSorted[0].version}})
  }

  backToCatalogs(): void {
    localStorage.removeItem('catalogs.details.chart_versions')
    this.router.navigate(['/portal/appmgmt/catalogs']);
  }

  getAllApplication(): void {
    this.applicationService.getAllApplicationsByOrg(this.orgId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APPLICATION.GET-APP-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.appList = !!data.data?.Applications?.edge?.nodes ? data.data.Applications.edge.nodes : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APPLICATION.GET-APP-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
  }

  getAllComponentOfApp($event: any): void {
    this.componentList = [];
    const appId = $event.target.value;
    this.cptService.getAllComponentsByApplication(appId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.componentList = !!data.data?.Components?.edge?.nodes ? data.data.Components.edge.nodes : [];
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
    )
  }

  getParamFromRoute(): void {
    this.state$ = this.route.paramMap.pipe(
      map(() => window.history.state),
    )
    this.state$.subscribe( state => {
      if(!!state.versions) {
        this.chartVersionList = JSON.parse(state.versions) as ChartVersion[];
        const chartVersionListSorted = this.chartVersionList.sort();
        localStorage.setItem('catalogs.details.chart_versions', JSON.stringify(chartVersionListSorted))
      }
    })
   }

  getChartDetails($event: any): void {
    const version = $event.target.value;
    this.loading  = true
      this.repositoryService.getChartDetails(this.chartVersionListSorted[0]._imageRegistry, this.chartVersionListSorted[0]._project, this.chartVersionListSorted[0]._repo, version).subscribe(data => {
        if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CHARTS-DETAILS.GET-CHARTDETAILS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
        }
        this.chartDetails = !!data.data?.ChartDetails ? data.data.ChartDetails : [];
      }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CHARTS-DETAILS.GET-CHARTDETAILS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false );
  }

  showAddComponent(): void {
    this.getAllApplication();
    this.originValues = this.chartDetails.files['values.yaml']
    this.newValues = this.originValues;
    this.showComponentModel = true;
  }

  doComponentFinish(): void {
     const extraParam: JSON = JSON.parse(JSON.stringify({}));
     extraParam['currentChartRegistry'] = this.chartVersionListSorted[0]._imageRegistry;
     extraParam['currentChartProject'] = this.chartVersionListSorted[0]._project;
     extraParam['currentChartName'] = this.chartVersionListSorted[0]._repo;
     extraParam['currentChartVersion'] = this.chartVersionForm.value.version;
     const param: CreateComponent = {
       name: this.basicHelmComponentForm.value.name,
       description: this.basicHelmComponentForm.value.description,
       template: {
         helm: {
           url: this.chartDetails.download,
           values: YAML.parse(this.newValues),
           extraInfo: extraParam
         }
       },
       application: this.basicHelmComponentForm.value.applicationId
     }
    this.cptService.createComponent(param).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.CREATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.CreateComponent){
        this.toast.show(`${data.data.CreateComponent.name} `+this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-MSG'), this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.COMPONENT.CREATE-CPT-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doComponentReset();
    });
  }

  doComponentCancel(): void {
    this.doComponentReset();
  }

  doComponentReset(): void {
    this.wizardnewhelm.reset();
    this.showComponentModel = false;
    this.basicHelmComponentForm.reset();
    this.appList = [];
    this.componentNameExist = false;
    this.componentList = [];
    this.showCheckedAlert = false;
    this.checkFormateOk = false;
    this.parseError = '';
  }

  resetValues(): void {
    this.newValues = this.originValues;
    this.showCheckedAlert = false;
    this.checkFormateOk = false;
    this.parseError = '';
  }

  checkFileFormat(): void {
    [this.checkFormateOk, this.parseError] = checkOneYamlFormat(this.newValues)
    this.showCheckedAlert = true;
  }

  modelValueChange(): void {
    this.showCheckedAlert = false;
    this.checkFormateOk = false;
  }

  checkNameExist(name: string): boolean {
    this.componentNameExist = !!this.componentList.find(x => x.name === name);
    return !!this.componentList.find(x => x.name === name);
  }


}
