import { Component, ViewChild } from '@angular/core';
import { ClrDatagridPagination, ClrDatagridStateInterface } from '@clr/angular';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from '@ngx-translate/core';
import { AuthService } from '../../../services/auth.service';
import { StandardPackageRepositoryService } from '../../../services/standard-package-repository.service';
import {
  CreateStandardPackageRepository,
  StandardPackageRepository,
  UpdateStandardPackageRepository
} from '../../../classes';
import { CustomValidatorDirective, nameValidator } from '../../../utils/validators/custom-validator.directive';

@Component({
  selector: 'app-standard-package-repository',
  templateUrl: './standard-package-repository.component.html',
  styleUrls: ['./standard-package-repository.component.css']
})
export class StandardPackageRepositoryComponent {
  constructor (
    private standardPackageRepositoryService: StandardPackageRepositoryService,
    public authService: AuthService,
    private toast: ToastrService,
    private translate: TranslateService
  ) { }

  @ViewChild('pagination') pagination: ClrDatagridPagination;

  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  readonly defaultImage = 'harbor.tanzu.cn/tkg/packages/standard/repo:v2.1.0';
  readonly imagePattern = /^[^/\s:]+(?:\/[^/\s:]+)+:[^:\s/]+$/;

  standardPackageRepositories: StandardPackageRepository[] = [];
  selectedRepository: StandardPackageRepository;
  showSearch = true;
  searchName: string;
  total = 0;
  loading = true;

  addOpen = false;
  editOpen = false;
  deleteOpen = false;

  repositoryForm = new FormGroup({
    name: new FormControl('', [Validators.required, Validators.maxLength(48), Validators.pattern(this.restricNamePattern), nameValidator([])]),
    description: new FormControl(''),
    image: new FormControl(this.defaultImage, [Validators.required, Validators.pattern(this.imagePattern)]),
    default: new FormControl(true)
  });

  refresh (state?: ClrDatagridStateInterface): void {
    this.loading = true;
    const filters = [];
    if (this.searchName) {
      filters.push({ key: 'name', value: [this.searchName] });
    }
    this.standardPackageRepositoryService.getStandardPackageRepositories(state?.page?.current ? state.page.current - 1 : 0, state?.page?.size ? state.page.size : Number(this.pagination.pageSize), filters).subscribe(data => {
      if (data.errors) {
        this.showErrors(data.errors, 'MESSAGE.STANDARDPACKAGE.GET-FAILED');
      }
      this.standardPackageRepositories = data.data?.StandardPackageRepositories?.edge?.nodes ? data.data.StandardPackageRepositories.edge.nodes : [];
      this.selectedRepository = null;
      this.total = data.data?.StandardPackageRepositories?.totalCount ? data.data.StandardPackageRepositories.totalCount : 0;
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.STANDARDPACKAGE.GET-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => { this.loading = false });
  }

  showAdd (): void {
    this.repositoryForm.reset();
    this.repositoryForm.patchValue({
      image: this.defaultImage,
      default: true
    });
    this.addOpen = true;
  }

  showEdit (repository: StandardPackageRepository): void {
    this.selectedRepository = repository;
    this.repositoryForm.setValue({
      name: repository.name,
      description: repository.description ? repository.description : '',
      image: repository.image,
      default: repository.default
    });
    this.editOpen = true;
  }

  showDelete (repository: StandardPackageRepository): void {
    this.selectedRepository = repository;
    this.deleteOpen = true;
  }

  doAddFinish (): void {
    const input: CreateStandardPackageRepository = {
      name: this.repositoryForm.value.name,
      description: this.repositoryForm.value.description,
      image: this.repositoryForm.value.image,
      default: this.repositoryForm.value.default
    };
    this.standardPackageRepositoryService.createStandardPackageRepository(input).subscribe(data => {
      if (data.errors) {
        this.showErrors(data.errors, 'MESSAGE.STANDARDPACKAGE.ADD-FAILED');
      }
      if (data.data?.CreateStandardPackageRepository) {
        this.toast.show(`${data.data.CreateStandardPackageRepository.name} ` + this.translate.instant('MESSAGE.STANDARDPACKAGE.ADD-MSG'), this.translate.instant('MESSAGE.STANDARDPACKAGE.ADD-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 });
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.STANDARDPACKAGE.ADD-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.addOpen = false;
      this.refreshFirstPage();
    });
  }

  doEditFinish (): void {
    const input: UpdateStandardPackageRepository = {
      id: this.selectedRepository?.id,
      name: this.repositoryForm.value.name,
      description: this.repositoryForm.value.description,
      image: this.repositoryForm.value.image,
      default: this.repositoryForm.value.default
    };
    this.standardPackageRepositoryService.updateStandardPackageRepository(input).subscribe(data => {
      if (data.errors) {
        this.showErrors(data.errors, 'MESSAGE.STANDARDPACKAGE.UPDATE-FAILED');
      }
      if (data.data?.UpdateStandardPackageRepository) {
        this.toast.show(`${data.data.UpdateStandardPackageRepository.name} ` + this.translate.instant('MESSAGE.STANDARDPACKAGE.UPDATE-MSG'), this.translate.instant('MESSAGE.STANDARDPACKAGE.UPDATE-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 });
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.STANDARDPACKAGE.UPDATE-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.editOpen = false;
      this.refreshFirstPage();
    });
  }

  setDefault (repository: StandardPackageRepository): void {
    this.standardPackageRepositoryService.setDefaultStandardPackageRepository(repository.id).subscribe(data => {
      if (data.errors) {
        this.showErrors(data.errors, 'MESSAGE.STANDARDPACKAGE.UPDATE-FAILED');
      }
      if (data.data?.SetDefaultStandardPackageRepository) {
        this.toast.show(`${data.data.SetDefaultStandardPackageRepository.name} ` + this.translate.instant('MESSAGE.STANDARDPACKAGE.SET-DEFAULT-MSG'), this.translate.instant('MESSAGE.STANDARDPACKAGE.UPDATE-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 });
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.STANDARDPACKAGE.UPDATE-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => this.refreshFirstPage());
  }

  deleteRepository (): void {
    this.standardPackageRepositoryService.deleteStandardPackageRepository(this.selectedRepository?.id).subscribe(data => {
      if (data.errors) {
        this.showErrors(data.errors, 'MESSAGE.STANDARDPACKAGE.DELETE-FAILED');
      }
      if (data.data?.DeleteStandardPackageRepository) {
        this.toast.show(`${data.data.DeleteStandardPackageRepository.name} ` + this.translate.instant('MESSAGE.STANDARDPACKAGE.DELETE-MSG'), this.translate.instant('MESSAGE.STANDARDPACKAGE.DELETE-SUCCESS'), { toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000 });
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.STANDARDPACKAGE.DELETE-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => {
      this.deleteOpen = false;
      this.refreshFirstPage();
    });
  }

  private refreshFirstPage (): void {
    if (this.pagination.currentPage === 1) {
      this.refresh();
    } else {
      this.pagination.currentPage = 1;
    }
  }

  private showErrors (errors: any[], titleKey: string): void {
    for (const each of errors) {
      if (each.message) {
        this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant(titleKey), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 });
      }
    }
  }
}
