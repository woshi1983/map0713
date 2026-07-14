import {Component, OnInit, ViewChild} from '@angular/core';
import {AboutDialogComponent} from "../../base/about-dialog/about-dialog.component";
import {ModalEvent} from "../../../classes";
import {AuthService} from "../../../services/auth.service";


@Component({
  selector: 'app-logout',
  templateUrl: './logout.component.html',
  styleUrls: ['./logout.component.css']
})
export class LogoutComponent implements OnInit {

  @ViewChild(AboutDialogComponent)
  aboutDialogView: AboutDialogComponent;

  constructor(
    private authService: AuthService
    ) { }

  ngOnInit(): void {
    this.authService.logout();
  }

  backToConsole(): void {
    const href = window.location.protocol + '//' +  window.location.host + '/portal';
    window.location.href = window.location.origin + '/api/auth/.well-known/login?redirectURI=' + encodeURI(href);
  }

  openModal(event: ModalEvent): void {
    switch (event.modalName) {
      case 'ABOUT':
          this.aboutDialogView.open();
          break;
      default:
          break;
    }
  }


}
