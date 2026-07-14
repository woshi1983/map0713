import {Component, Input, OnInit, Output, EventEmitter, SimpleChanges, OnChanges, OnDestroy} from '@angular/core';
import { Route } from '../../../../../classes';
import { RouteService } from "../../../../../services/route.service";
import { ToastrService } from 'ngx-toastr';
import { TranslateService } from "@ngx-translate/core";
import { EChartsOption } from 'echarts';

@Component({
  selector: 'app-route-details',
  templateUrl: './route-details.component.html',
  styleUrls: ['./route-details.component.css'],
  exportAs: 'app-route-details'
})
export class RouteDetailsComponent implements OnInit, OnChanges, OnDestroy {

  @Input() routeId: string;
  @Input() open = false;

  route: Route;
  data: any[] = []
  chartOption: EChartsOption;

  @Output() close = new EventEmitter<string>();

  constructor(
    private routeService: RouteService,
    private toast: ToastrService,
    private translate: TranslateService) { }

  ngOnChanges(changes: SimpleChanges): void {
    this.data = [];
    if (!!this.routeId) {
      this.routeService.getRoute(this.routeId).subscribe((d) => {
        if (!!d.data?.Route) {
          this.route = d.data.Route;
          for(var i = 0; i < this.route.backends.length; i++){
            i === 0 ?
            this.data.push({ value: this.route.backends[0].ratio*5, name: `${this.route.backends[0].name} (${this.route.backends[0].instance.cluster.name}/${this.route.backends[0].instance.namespace})`}) :
              this.data.push({ value: this.route.backends[1].ratio*5, name: `${this.route.backends[1].name} (${this.route.backends[1].instance.cluster.name}/${this.route.backends[1].instance.namespace}) `})
          }
          this.chartOption = {
            color: ['#0072a3', '#60b515'],
            tooltip: {
              trigger: 'item'
            },
            legend: {
              left: 'center'
            },
            series: [
              {
                name: 'backend',
                type: 'pie',
                radius: ['50%', '70%'],
                avoidLabelOverlap: true,
                itemStyle: {
                  borderRadius: 8,
                  borderColor: '#fff',
                  borderWidth: 3
                },
                label: {
                  show: true,
                  formatter: '{d}%'
                },
                labelLine: {
                  show: true
                },
                data: this.data
              }
            ]
          };
        }
      }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.APP-DETAILS.ROUTE.GET-ROUTE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}));
    }
  }

  routeStatusChartInit($event: any): void {
    var routeStatusChart = $event;
    window.onresize = function() {
      routeStatusChart.resize();
    };
  }

  ngOnDestroy() {
    this.data = [];
    this.route = null;
  }

  ngOnInit(): void {
  }

}

