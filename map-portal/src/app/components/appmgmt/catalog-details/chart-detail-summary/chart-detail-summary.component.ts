import { ChangeDetectionStrategy, Component, OnInit, Input } from '@angular/core';

@Component({
  selector: 'app-chart-detail-summary',
  templateUrl: './chart-detail-summary.component.html',
  styleUrls: ['./chart-detail-summary.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ChartDetailSummaryComponent implements OnInit {

  @Input() description: string;
  @Input() readme: string;
  @Input() homepage: string;
  @Input() download: string;
  @Input() sources: string[];
  @Input() keywords: string[];

  constructor() { }

  ngOnInit(): void {
  }

}
