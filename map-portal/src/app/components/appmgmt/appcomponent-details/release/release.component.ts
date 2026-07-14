import {Component, OnInit, Input, ViewChild} from '@angular/core';
import {AuthService} from "../../../../services/auth.service";
import { ToastrService } from "ngx-toastr";
import { TranslateService } from "@ngx-translate/core";
import {ClrDatagridPagination, ClrDatagridStateInterface} from "@clr/angular";
import { Release } from "../../../../classes";
import {ReleaseService} from "../../../../services/release.service";
import YAML from 'yaml';
import { saveAs } from 'file-saver';

@Component({
  selector: 'app-component-release',
  templateUrl: './release.component.html',
  styleUrls: ['./release.component.css']
})
export class ReleaseComponent implements OnInit {

  @Input() componentId: string;

  constructor(
    public authService: AuthService,
    private releaseService: ReleaseService,
    private toast: ToastrService,
    private translate: TranslateService
  ) { }

  allReleases: Release[] = [];
  selectedReleaseName: string;
  manifests: any[] = [];
  manifestString: string;
  manifestLoading = false;

  showSearchInput = false;
  showManifestModel: boolean;
  loading = false;
  total = 0;

  @ViewChild('pagination') pagination: ClrDatagridPagination;

  ngOnInit(): void {
  }

  refresh(state?: ClrDatagridStateInterface): void {
    this.loading = true;
    this.releaseService.getAllReleases(!!state?.page?.current ? state.page.current - 1 : 0, !!state?.page?.size ? state.page.size : Number(this.pagination.pageSize), this.componentId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.RELEASE.GET-RELEASE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.allReleases = !!data.data?.Releases?.edge?.nodes ? data.data.Releases.edge.nodes : [];
      this.total = !!data?.data?.Releases?.totalCount ? data.data.Releases.totalCount : 0;
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.RELEASE.GET-RELEASE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false);
  }

  showReleaseManifests(releaseId: string): void {
    this.showManifestModel = true;
    this.manifestLoading = true;
    this.manifestString = '';
    const mainfestStringList: any[] = [];
    this.releaseService.getReleaseById(releaseId).subscribe(data => {
      if (!!data.errors){
          for (const each of data.errors){
             if(!!each.message){
               this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.APP-DETAILS.RELEASE.GET-RELEASE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
             }
          }
        }
      if(!!data.data.Release){
        this.manifests = data.data.Release.manifest;
        this.selectedReleaseName = data.data.Release.name + '-' + data.data.Release.num;
        for(const manifest of this.manifests){
          const objectString: string = YAML.stringify(manifest);
          mainfestStringList.push(objectString);
        }
        this.manifestString = mainfestStringList.join('---\n')
      } else {
        this.manifests = [];
        this.selectedReleaseName = '';
        this.manifestString = '';
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.RELEASE.GET-RELEASE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {this.manifestLoading = false;})
  }

  exportManifests(): void {
    const blob = new Blob([this.manifestString], { type: 'text/yaml' });
    saveAs(blob, this.selectedReleaseName + '-manifests-template.yaml');
  }

}
