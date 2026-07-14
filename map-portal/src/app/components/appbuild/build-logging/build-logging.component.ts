import { Component, OnChanges, Input, Output, EventEmitter } from '@angular/core';
import { Apollo, gql } from 'apollo-angular';
import { AuthService } from '../../../services/auth.service';
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from '@ngx-translate/core';
import { Subscription } from 'rxjs';

@Component({
  selector: 'app-build-logging',
  templateUrl: './build-logging.component.html',
  styleUrls: ['./build-logging.component.css']
})
export class BuildLoggingComponent implements OnChanges {
  @Input() clusterId: string;
  @Input() namespace: string;
  // set build-logging type, support deployment, source_build, binary_build, dockerfile_build
  @Input() controller: string;
  // set container name, or build name
  @Input() issuer: string;
  @Input() showLogs = false;

  @Output() close = new EventEmitter<string>();

  loggingMessage = '';
  loading = false;
  subscription: Subscription;

  constructor (
    public auth: AuthService,
    private apollo: Apollo,
    private toast: ToastrService,
    private translate: TranslateService
  ) { }

  ngOnChanges (): void {
    if (this.showLogs && !!this.controller && !!this.clusterId && !!this.namespace && !!this.issuer) {
      this.getBuildLogging();
    }
  }

  getBuildLogging (): void {
    this.loading = true;
    this.subscription = this.apollo.query<any>({
      query: gql`
        query {
          SmartPodLog(cluster: "${this.clusterId}", namespace: "${this.namespace}", controller: "${this.controller}", issuer: "${this.issuer}") {
            name
            namespace
            container
            message
          }
        }`,
      errorPolicy: 'all'
    }).subscribe((data) => {
      if (data.data.SmartPodLog) {
        this.loggingMessage = data.data.SmartPodLog.message;
      } else {
        this.loggingMessage = '';
        if (data.errors) {
          for (const each of data.errors) {
            if (each.message) {
              this.toast.show(this.translate.instant('MESSAGE.SHARED.GETERROR') + ` ${each.message} ${each.extensions?.reason ? ' : ' + each.extensions?.reason : ''}`, this.translate.instant('MESSAGE.LOGGING.GET-LOG-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 })
            }
          }
        }
      }
    }, (e) => this.toast.show(this.translate.instant('MESSAGE.SHARED.UNGET') + ` ${e}`, this.translate.instant('MESSAGE.LOGGING.GET-LOG-FAILED'), { toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0 }), () => { this.loading = false; })
  }
}
