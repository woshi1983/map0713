import {Component, OnInit, ViewChild, Input} from '@angular/core';
import {AuthService} from "../../../../services/auth.service";
import {
  ClusterProviderSpecTrustedCA,
  AddClusterProviderTrustedCA,
  RemoveClusterProviderTrustedCA
} from "../../../../classes";
import {ClrDatagridPagination, ClrWizard} from "@clr/angular";
import {ProviderService} from "../../../../services/provider.service";
import {ActivatedRoute} from "@angular/router";
import {FormArray, FormBuilder, FormControl, FormGroup, Validators} from "@angular/forms";
import {CAFormatValidator, CustomValidatorDirective, nameValidator} from "../../../../utils/validators/custom-validator.directive";
import { ToastrService } from 'ngx-toastr';
import {TranslateService} from "@ngx-translate/core";

@Component({
  selector: 'app-catrust',
  templateUrl: './catrust.component.html',
  styleUrls: ['./catrust.component.css']
})
export class CatrustComponent implements OnInit {

  @Input() providerId;

  constructor(
    public authService: AuthService,
    private providerService: ProviderService,
    private route: ActivatedRoute,
    private fb: FormBuilder,
    private toast: ToastrService,
    private translate: TranslateService) {
  }

  allCaTrust: ClusterProviderSpecTrustedCA[] = [];
  selectedCas: ClusterProviderSpecTrustedCA[] = [];
  providerCLass: string;

  loading = false;
  showSearch = false;

  addCaTrustOpen = false;
  removeDialog = false;

  caCertDialog = false;
  caCert: string;

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  @ViewChild('wizardCaTrust') wizardCaTrust: ClrWizard;
  readonly restricNamePattern = CustomValidatorDirective.restricNamePattern;
  caList = new FormArray([]);
  trustCAFrom = new FormGroup({
    trustCA: this.caList
  });
  deleteCaForm = new FormGroup({
    confirm: new FormControl('', [Validators.required]),
  });

  ngOnInit(): void {
  }

  refresh(): void {
    this.loading = true;
    this.providerService.getCaTrustByProvider(this.providerId).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.GET-CA-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if (!!data.data?.ClusterProvider) {
        this.providerCLass = data.data.ClusterProvider.class;
        this.allCaTrust = !!data.data.ClusterProvider.spec.trustedCAs ? data.data.ClusterProvider.spec.trustedCAs : [];
        this.selectedCas = [];
      } else {
        this.providerCLass = '';
        this.allCaTrust = [];
        this.selectedCas = [];
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.GET-CA-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false);
  }


  showAddCa(): void {
     this.addCaTrustOpen = true;
  }

  private doCaTrustReset(): void {
    this.wizardCaTrust.reset();
    this.addCaTrustOpen = false;
    while (this.caList.length !== 0) {
      this.caList.removeAt(0);
    }
  }

  public doAddCaTrustFinish(): void {
    const caTrustList: AddClusterProviderTrustedCA[] = [];
    if (!!this.trustCAFrom.value.trustCA) {
      for (const each of this.trustCAFrom.value.trustCA) {
        caTrustList.push({name: each.caName, cert: each.caCert});
      }
    }
    switch (this.providerCLass) {
      case 'TKGS':
        this.providerService.addTKGSProviderCaTrust(this.providerId, caTrustList).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.ADDERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.ADD-CA-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.UpdateTKGSClusterProvider) {
              this.toast.show(this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.ADD-CA-MSG'), this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.ADD-CA-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.ADD-CA-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          this.doCaTrustReset();
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
        });
        break;
      case 'TKGM':
        this.providerService.addTKGMProviderCaTrust(this.providerId, caTrustList).subscribe(data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.ADDERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.ADD-CA-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.UpdateTKGMClusterProvider) {
              this.toast.show(this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.ADD-CA-MSG'), this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.ADD-CA-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.ADD-CA-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          this.doCaTrustReset();
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
        });
        break;
      default:
        break;
    }

  }

  public doAddCaTrustCancel(): void {
    this.doCaTrustReset();
  }


  deleteCA(index: number): void {
    this.caList.removeAt(index);
  }

  addCA(): void {
    this.caList.push(this.fb.group({
      caName: new FormControl('', [Validators.required, Validators.pattern(this.restricNamePattern), nameValidator([])]),
      caCert: new FormControl('', [Validators.required, CAFormatValidator()]),
    }));
  }

  deleteCaByName(): void {
    this.removeDialog = false;
    const removeCaList: RemoveClusterProviderTrustedCA[] = [];
    let caInfo: ClusterProviderSpecTrustedCA;
    for (caInfo of this.selectedCas) {
       removeCaList.push({name: caInfo.name});
    }
    switch (this.providerCLass) {
      case 'TKGS':
        this.providerService.deleteTKGSProviderCaTrust(this.providerId, removeCaList).subscribe( data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.DELETE-CA-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.UpdateTKGSClusterProvider){
              this.toast.show(this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.DELETE-CA-MSG'), this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.DELETE-CA-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.DELETE-CA-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
        })
        break;
      case 'TKGM':
        this.providerService.deleteTKGMProviderCaTrust(this.providerId, removeCaList).subscribe( data => {
          if (!!data.errors){
            for (const each of data.errors){
               if(!!each.message){
                 this.toast.show( this.translate.instant('MESSAGE.SHARED.DELETEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.DELETE-CA-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
               }
            }
          }
          if (!!data.data?.UpdateTKGMClusterProvider){
              this.toast.show(this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.DELETE-CA-MSG'), this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.DELETE-CA-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
          }
        }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.PROVIDERDETAILS.CA.DELETE-CA-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
          if (this.pagination.currentPage === 1) {
            this.refresh();
          } else {
            this.pagination.currentPage = 1;
          }
        })
        break;
      default:
        break;
    }
  }

  showCaCert(ca: string): void {
    this.caCertDialog = true;
    this.caCert = ca;
  }
}


