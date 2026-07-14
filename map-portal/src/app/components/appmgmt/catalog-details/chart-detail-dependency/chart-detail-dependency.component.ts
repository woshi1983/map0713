import { Component, OnInit, Input, ChangeDetectionStrategy } from '@angular/core';
import { ChartDetailsDependency } from '../../../../classes'

@Component({
  selector: 'app-chart-detail-dependency',
  templateUrl: './chart-detail-dependency.component.html',
  styleUrls: ['./chart-detail-dependency.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ChartDetailDependencyComponent implements OnInit {

  @Input() dependencies: ChartDetailsDependency[];

  constructor() { }

  ngOnInit(): void {
  }

}
