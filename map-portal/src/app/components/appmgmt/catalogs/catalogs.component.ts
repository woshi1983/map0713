import { Component, OnInit } from '@angular/core';
import {RepositorysService} from "../../../services/repositorys.service";
import {ChartRepo, ImageRegistry} from "../../../classes";
import {OrganizationService} from "../../../services/organization.service";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {AuthService} from "../../../services/auth.service";
import {FormControl, FormGroup, Validators} from "@angular/forms";
import { environment} from '../../../../environments/environment';

@Component({
  selector: 'app-catalogs',
  templateUrl: './catalogs.component.html',
  styleUrls: ['./catalogs.component.css']
})
export class CatalogsComponent implements OnInit {

  constructor(
    private repositoryService: RepositorysService,
    private orgService: OrganizationService,
    private toast: ToastrService,
    private translate: TranslateService,
    public authService: AuthService
  ) {
    this.orgId = this.orgService.getSelectedOrgId;
    this.backgroupImage = environment.assetsPath + 'images/tanzu-back.png';
  }

  orgId: string;
  registryList: ImageRegistry[] = [];
  projectList: string[] = [];
  chartReposList: ChartRepo[] = [];

  selectForm = new FormGroup({
    registry: new FormControl('', Validators.required),
    project: new FormControl('', Validators.required)
  });

  showSearch = true;
  searchName: string;
  showNoProjectAlert = false;
  loading = false;
  noChartRepo = false;

  backgroupImage: string;
  imagePath: string = environment.assetsPath + 'images';

  ngOnInit(): void {
    this.getAllRegistryAndProject();
  }

  getAllRegistryAndProject(): void {
    this.repositoryService.getAllRegistrys(this.orgId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.REPOSITORY.GET-REPO-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.ImageRegistries?.edge?.nodes) {
        const tempList = data.data.ImageRegistries.edge.nodes;
        this.registryList = tempList.sort((n1, n2) => {
              if(n1.name > n2.name){return 1;}
              if(n1.name < n2.name){return -1;}
              return 0
        });
        this.projectList = this.registryList[0].projects;
        if (this.projectList.length > 0){
          this.selectForm.setValue({
              registry: this.registryList[0].id,
              project: this.projectList[0]
            }
          )
          this.showNoProjectAlert = false;
          this.refresh();
        } else {
         this.selectForm.setValue({
            registry: this.registryList[0].id,
            project: ''
         })
          this.showNoProjectAlert = true;
          this.noChartRepo = true;
        }
      } else {
        this.registryList = [];
        this.projectList = []
        this.selectForm.setValue({
          registry: '',
          project: ''
          }
        )
        this.noChartRepo = true;
        this.showNoProjectAlert = false;
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.REPOSITORY.GET-REPO-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}),  );
  }

  changeProjectList($event: any): void {
    const registryId = $event.target.value;
    for (const registry of this.registryList){
      if (registry.id === registryId){
        this.projectList = registry.projects;
        if (this.projectList.length > 0){
          this.selectForm.setValue({
            registry: registryId,
            project: this.projectList[0]
          })
          this.noChartRepo = false;
          this.refresh();
        } else {
          this.selectForm.setValue({
            registry: registryId,
            project: ''
          })
          this.chartReposList = [];
          this.noChartRepo = true;
        }
      }
    }
  }

  refresh(): void {
    this.loading  = true
    const filters = [];
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.repositoryService.getAllChartRepos(this.selectForm.value.registry, this.selectForm.value.project, filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CATALOGS.GET-CHARTS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.ChartRepos) {
        this.chartReposList = data.data.ChartRepos;
        this.noChartRepo = false;
      } else {
        this.chartReposList = [];
        this.noChartRepo = true;
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CATALOGS.GET-CHARTS-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false );
  }
}
