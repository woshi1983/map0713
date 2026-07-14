import { Component, OnInit } from '@angular/core';
import {AuthService} from "../../../services/auth.service";
import {ActivatedRoute} from "@angular/router";

@Component({
  selector: 'app-approutes',
  templateUrl: './approutes.component.html',
  styleUrls: ['./approutes.component.css']
})
export class ApproutesComponent implements OnInit {

  constructor(public authService: AuthService, private route: ActivatedRoute) { }

  tabLinkNavList = [
    {
      linkName: 'ingress',
      showTabName: 'APPROUTE.TABS.INGRESS',
      tabLinkInOverflow: false
    },
    {
      linkName: 'route',
      showTabName: 'APPROUTE.TABS.ROUTES',
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
