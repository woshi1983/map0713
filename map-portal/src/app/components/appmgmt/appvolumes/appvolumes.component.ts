import { Component, OnInit } from '@angular/core';
import { AuthService } from '../../../services/auth.service';
import {ActivatedRoute} from '@angular/router';



@Component({
  selector: 'app-appvolumes',
  templateUrl: './appvolumes.component.html',
  styleUrls: ['./appvolumes.component.css']
})
export class AppvolumesComponent implements OnInit {

  constructor(public authService: AuthService, private route: ActivatedRoute) { }

  tabLinkNavList = [
    {
      linkName: 'configmap',
      showTabName: 'APPVOLUME.VOLUME_TABS.CM',
      tabLinkInOverflow: false
    },
    {
      linkName: 'secret',
      showTabName: 'APPVOLUME.VOLUME_TABS.SECRET',
      tabLinkInOverflow: false
    },
    {
      linkName: 'pvc',
      showTabName: 'APPVOLUME.VOLUME_TABS.PVC',
      tabLinkInOverflow: false
    },
    {
      linkName: 'storageclass',
      showTabName: 'APPVOLUME.VOLUME_TABS.SC',
      tabLinkInOverflow: false
    }
  ];


  ngOnInit(): void {
  }

  isTabLinkInOverFlow(): any {
    return this.tabLinkNavList.some(tab => {
      return tab.tabLinkInOverflow && this.route.snapshot.children[0].routeConfig.path === tab.linkName;
    });
  }

  isDefaultTab(tab, index): any {
    return this.route.snapshot.children[0].routeConfig.path !== tab.linkName && index === 0;
  }

}
