import { Component, OnInit, Input } from '@angular/core';
import { Cluster } from '../../../../classes';
import { AuthService } from '../../../../services/auth.service';


@Component({
  selector: 'app-cluster-summary',
  templateUrl: './summary.component.html',
  styleUrls: ['./summary.component.css']
})
export class SummaryComponent implements OnInit {

  @Input() cluster: Cluster;

  constructor(
    public authService: AuthService,
  ) { }

  ngOnInit(): void {
  }

}
