import { Component, OnInit } from '@angular/core';
import {CommonService} from "../../../services/common.service";
import {BrandingService} from '../../../services/branding.service';

@Component({
  selector: 'app-about-dialog',
  templateUrl: './about-dialog.component.html',
  styleUrls: ['./about-dialog.component.css']
})
export class AboutDialogComponent implements OnInit {

  version: string;
  opened: boolean = false;

  constructor(
    private commonService: CommonService,
    public branding: BrandingService
  ) {}

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
