import { Component, OnInit } from '@angular/core';
import {environment} from "../../../../environments/environment";
import {CommonService} from "../../../services/common.service";

@Component({
  selector: 'app-about-dialog',
  templateUrl: './about-dialog.component.html',
  styleUrls: ['./about-dialog.component.css']
})
export class AboutDialogComponent implements OnInit {

  version: string;
  opened: boolean = false;
  logoImage: string;

  constructor(
    private commonService: CommonService
  ) {
    this.logoImage = environment.assetsPath + 'images/tanzu-back.png';
  }

  ngOnInit(): void {
  }

  getAppVersion(): void {
    this.commonService.getAppVersion().subscribe( data => {
      if (data?.status === 200){
        this.version = 'v' + data.body.version + '-' + data.body.revision;
        this.opened = true;
      } else {
        this.version = '';
      }
    })
  }

  public open(): void {
    this.getAppVersion();
  }

  public close(): void {
      this.opened = false;
  }

}
