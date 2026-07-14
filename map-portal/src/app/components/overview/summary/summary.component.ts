import {Component, OnInit, ViewChild} from '@angular/core';
import {AuthService} from '../../../services/auth.service';
import {SummaryBasic, SummaryExtensions, ClusterStatusMetrics, Extension, OrgResourceMetrics, OrgMetrics} from '../../../classes/';
import {SummaryService} from "../../../services/summary.service";
import { EChartsOption } from 'echarts';
import {environment} from "../../../../environments/environment";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {ClrDatagridPagination} from "@clr/angular";
import {getSortIndexList} from "../../../utils/operator/sortindex";
import {FEATUREGATES} from "../../../utils/static-entries/shared.const";

@Component({
  selector: 'app-summary',
  templateUrl: './summary.component.html',
  styleUrls: ['./summary.component.css']
})
export class SummaryComponent implements OnInit {

  featureGates = FEATUREGATES;

  constructor(
    public authService: AuthService,
    private summaryService: SummaryService,
    private toast: ToastrService,
    private translate: TranslateService
  ) { }

  @ViewChild('pagination') pagination: ClrDatagridPagination;
  imagePath: string = environment.assetsPath + 'images';
  appMgmt = environment.features.hasOwnProperty('appMgmt') ? environment.features.appMgmt : true;

  summaryBasic: SummaryBasic;
  basicMetricsLoading = false;
  summaryExtensions: SummaryExtensions = {
      contour: {
        extensions: [],
        readyCount: 0
      },
      octant: {
        extensions: [],
        readyCount: 0,
      },
      grafana: {
        extensions: [],
        readyCount: 0,
      },
      prometheus: {
        extensions: [],
        readyCount: 0,
      },
      elasticsearch: {
        extensions: [],
        readyCount: 0,
      },
      fluentbit: {
        extensions: [],
        readyCount: 0,
      }
  };

  healthyClusterList: ClusterStatusMetrics[] = [];
  unhealthyClusterList: ClusterStatusMetrics[] = [];
  allCluster: ClusterStatusMetrics[] = [];
  allExtensions: Extension[] = [];
  clusterMetricsLoading = false;
  showClusterDetailsModel = false;
  showClusterStatus: string;
  showExtensionName: string;
  showExtensionDetailsModel = false;

  orgResourcesLoading = false;
  allOrgResourceMetrics: OrgResourceMetrics = {
    totalResource: {
      cpu: {
        used: 0,
        total: 0
      },
      memory: {
        used: 0,
        total: 0
      },
      storage: {
        used: 0,
        total: 0
      },
    },
    orgMetrics: {
      name: [],
      cpuUsed: [],
      cpuFree: [],
      memoryUsed: [],
      memoryFree: [],
      storageUsed: [],
      storageFree: []
    }
  }


  clusterStatusChartOption1: EChartsOption;
  clusterStatusChartOption2: EChartsOption;
  allTenantCPUOption: EChartsOption;
  allTenantMemoryOption: EChartsOption;
  allTenantStoageOption: EChartsOption;
  allTenantResourceUsageOption: EChartsOption;
  CPUUsageTop5Option: EChartsOption;
  MemoryUsageTop5Option: EChartsOption;
  StorageUsageTop5Option: EChartsOption;

  ngOnInit(): void {
    this.refreshBasicMerics();
    this.refreshClusterStatusMerics();
    this.refreshOrgResourceMerics();
  }

  refreshBasicMerics(): void {
    this.basicMetricsLoading = true;
    this.summaryService.getBasicMetrics().subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.SUMMARY.GET-BASIC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      this.summaryBasic = {
        orgCount: !!data.data.Organizations ? data.data.Organizations.totalCount : 0,
        activeUsers: !!data.data.Metrics ? data.data.Metrics.activeUsers : 0,
        customComponentCount: !!data.data.Components ? data.data.Components.totalCount : 0,
        instanceCount: !!data.data.Instances ? data.data.Instances.totalCount : 0,
        imageBuildsCount: (!!data.data.BinaryImages ? data.data.BinaryImages.length : 0) + (!!data.data.DockerfileImages ? data.data.DockerfileImages.length : 0) + (!!data.data.SourceImages ? data.data.SourceImages.length : 0),
        ingressCount: !!data.data.Ingresses ? data.data.Ingresses.length : 0,
        routeCount: !!data.data.Routes ? data.data.Routes.totalCount : 0,
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.SUMMARY.GET-BASIC-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.basicMetricsLoading = false)
  }

  refreshClusterStatusMerics(): void {
    this.clusterMetricsLoading = true;
    this.summaryService.getAllClustersStatus().subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.SUMMARY.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if(!!data.data?.Clusters?.edge?.nodes){
        for (let cluster of data.data.Clusters.edge.nodes){
          if(cluster.status === 'HEALTHY'){
            this.healthyClusterList.push({
              name: cluster.name,
              organization: cluster.organization.name,
              status: 'HEALTHY'
            })
          } else {
            this.unhealthyClusterList.push({
              name: cluster.name,
              organization: cluster.organization.name,
              status: 'UNHEALTHY'
            })
          }
          if(!!cluster?.extensions?.edge?.nodes){
            for(let extension of cluster.extensions.edge.nodes){
              switch (extension.name) {
                case 'CONTOUR':
                  this.summaryExtensions.contour.extensions.push({
                    cluster: cluster.name,
                    organization: cluster.organization.name,
                    state: extension.state
                  })
                  if(extension.state === 'ENABLED'){
                    this.summaryExtensions.contour.readyCount = this.summaryExtensions.contour.readyCount + 1;
                  }
                  break;
                case 'OCTANT':
                  this.summaryExtensions.octant.extensions.push({
                    cluster: cluster.name,
                    organization: cluster.organization.name,
                    state: extension.state
                  })
                  if(extension.state === 'ENABLED'){
                    this.summaryExtensions.octant.readyCount = this.summaryExtensions.octant.readyCount + 1;
                  }
                  break;
                case 'GRAFANA':
                  this.summaryExtensions.grafana.extensions.push({
                    cluster: cluster.name,
                    organization: cluster.organization.name,
                    state: extension.state
                  })
                  if(extension.state === 'ENABLED'){
                    this.summaryExtensions.grafana.readyCount = this.summaryExtensions.grafana.readyCount + 1;
                  }
                  break;
                case 'PROMETHEUS':
                  this.summaryExtensions.prometheus.extensions.push({
                    cluster: cluster.name,
                    organization: cluster.organization.name,
                    state: extension.state
                  })
                  if(extension.state === 'ENABLED'){
                    this.summaryExtensions.prometheus.readyCount = this.summaryExtensions.prometheus.readyCount + 1;
                  }
                  break;
                case 'ELASTICSEARCH':
                  this.summaryExtensions.elasticsearch.extensions.push({
                    cluster: cluster.name,
                    organization: cluster.organization.name,
                    state: extension.state
                  })
                  if(extension.state === 'ENABLED'){
                    this.summaryExtensions.elasticsearch.readyCount = this.summaryExtensions.elasticsearch.readyCount + 1;
                  }
                  break;
                case 'FLUENTBIT':
                    this.summaryExtensions.fluentbit.extensions.push({
                      cluster: cluster.name,
                      organization: cluster.organization.name,
                      state: extension.state
                    })
                    if(extension.state === 'ENABLED'){
                      this.summaryExtensions.fluentbit.readyCount = this.summaryExtensions.fluentbit.readyCount + 1;
                    }
                    break;
              }
            }
          }
        }
      }
      this.drawClusterHealthyChart(this.healthyClusterList.length, this.unhealthyClusterList.length);
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.SUMMARY.GET-CLUSTER-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.clusterMetricsLoading = false)
  }

  refreshOrgResourceMerics(): void {
    this.orgResourcesLoading = true;
    this.summaryService.getOrgQuota().subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.SUMMARY.GET-ORG-RESOURCE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if(!!data.data?.Organizations?.edge?.nodes){
        let allOrgMemoryUsedMi = 0;
        let allOrgMemoryTotalMi = 0;
        for(let org of data.data.Organizations.edge.nodes){
          let cpuUsed = 0;
          let cpuFree = 0;
          let memoryUsed = 0;
          let memoryFree = 0;
          let storageUsed = 0;
          let storageFree = 0;
          if(org.quota.total.cpus !== 0){
            cpuFree = org.quota.total.cpus - org.quota.used.cpus;
            cpuUsed = org.quota.used.cpus;
            this.allOrgResourceMetrics.totalResource.cpu.used = this.allOrgResourceMetrics.totalResource.cpu.used + org.quota.used.cpus;
            this.allOrgResourceMetrics.totalResource.cpu.total = this.allOrgResourceMetrics.totalResource.cpu.total + org.quota.total.cpus;
          }
          if(org.quota.total.memoryMi !== 0){
            memoryFree = Math.round(((org.quota.total.memoryMi - org.quota.used.memoryMi) / 1024 + Number.EPSILON) * 1000 ) / 1000;
            memoryUsed = Math.round((org.quota.used.memoryMi / 1024 + Number.EPSILON) * 1000 ) / 1000;
            allOrgMemoryUsedMi = allOrgMemoryUsedMi + org.quota.used.memoryMi;
            allOrgMemoryTotalMi = allOrgMemoryTotalMi + org.quota.total.memoryMi;
          }
          if(org.quota.total.storageGi !== 0){
            storageFree = org.quota.total.storageGi - org.quota.used.storageGi;
            storageUsed = org.quota.used.storageGi;
            this.allOrgResourceMetrics.totalResource.storage.used = this.allOrgResourceMetrics.totalResource.storage.used + org.quota.used.storageGi;
            this.allOrgResourceMetrics.totalResource.storage.total = this.allOrgResourceMetrics.totalResource.storage.total + org.quota.total.storageGi;
          }
          if(org.quota.total.cpus !== 0 || org.quota.total.memoryMi !== 0 || org.quota.total.storageGi !== 0){
            this.allOrgResourceMetrics.orgMetrics.name.push(org.name);
            this.allOrgResourceMetrics.orgMetrics.cpuUsed.push(cpuUsed);
            this.allOrgResourceMetrics.orgMetrics.cpuFree.push(cpuFree);
            this.allOrgResourceMetrics.orgMetrics.memoryUsed.push(memoryUsed);
            this.allOrgResourceMetrics.orgMetrics.memoryFree.push(memoryFree);
            this.allOrgResourceMetrics.orgMetrics.storageUsed.push(storageUsed);
            this.allOrgResourceMetrics.orgMetrics.storageFree.push(storageFree);
          }

        }
        // compute total usage percent for all org resource
        this.allOrgResourceMetrics.totalResource.memory.used = Math.round((allOrgMemoryUsedMi / 1024 + Number.EPSILON) * 1000) / 1000;
        this.allOrgResourceMetrics.totalResource.memory.total = Math.round((allOrgMemoryTotalMi / 1024 + Number.EPSILON) * 1000) / 1000;
        const cpuPercent = this.allOrgResourceMetrics.totalResource.cpu.total === 0 ? 0 : Number((this.allOrgResourceMetrics.totalResource.cpu.used / this.allOrgResourceMetrics.totalResource.cpu.total)) * 100;
        const memoryPercent = this.allOrgResourceMetrics.totalResource.memory.total === 0 ? 0 : Number((this.allOrgResourceMetrics.totalResource.memory.used / this.allOrgResourceMetrics.totalResource.memory.total)) * 100;
        const storagePercent = this.allOrgResourceMetrics.totalResource.storage.total === 0 ? 0 : Number((this.allOrgResourceMetrics.totalResource.storage.used / this.allOrgResourceMetrics.totalResource.storage.total)) * 100;
        this.drawAllTenantCpuGauge(Math.round((cpuPercent + Number.EPSILON) * 100) /100);
        this.drawAllTenantMemoryGauge(Math.round((memoryPercent + Number.EPSILON) * 100) /100);
        this.drawAllTenantStorageGauge(Math.round((storagePercent + Number.EPSILON) * 100) /100);
        this.drawAllTenantResourceUsageBar(this.allOrgResourceMetrics.orgMetrics);
        // compare the first five
        let cpuNameList = [];
        let cpuSortedResult = [];
        if(this.allOrgResourceMetrics.orgMetrics.cpuUsed.length > 0){
          const cpuSorted = getSortIndexList(this.allOrgResourceMetrics.orgMetrics.cpuUsed);
          cpuSortedResult = cpuSorted[1].slice(0, 5);
          const cpuSortIndex = cpuSorted[0].slice(0, 5);
          for (let i of cpuSortIndex){
            cpuNameList.push(this.allOrgResourceMetrics.orgMetrics.name[i])
          }
        }
        this.drawCPUUsageTop5Bar(cpuNameList, cpuSortedResult);
        let memoryNameList = [];
        let memorySortedResult = [];
        if(this.allOrgResourceMetrics.orgMetrics.memoryUsed.length > 0){
          const memorySorted = getSortIndexList(this.allOrgResourceMetrics.orgMetrics.memoryUsed);
          memorySortedResult = memorySorted[1].slice(0, 5);
          const memorySortIndex = memorySorted[0].slice(0, 5);
          for (let j of memorySortIndex){
             memoryNameList.push(this.allOrgResourceMetrics.orgMetrics.name[j])
          }
        }
        this.drawMemoryUsageTop5Bar(memoryNameList, memorySortedResult);
        let storageNameList = [];
        let storageSortedResult = [];
        if(this.allOrgResourceMetrics.orgMetrics.storageUsed.length > 0){
          const storageSorted = getSortIndexList(this.allOrgResourceMetrics.orgMetrics.storageUsed);
          storageSortedResult = storageSorted[1].slice(0, 5);
          const storageSortIndex = storageSorted[0].slice(0, 5);
          for (let k of storageSortIndex){
             storageNameList.push(this.allOrgResourceMetrics.orgMetrics.name[k])
          }
        }
        this.drawStorageUsageTop5Bar(storageNameList, storageSortedResult);
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.SUMMARY.GET-ORG-RESOURCE-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.orgResourcesLoading = false)
  }

  drawClusterHealthyChart(healthyClusterCount: number, unhealthyClusterCount: number): void {
    this.clusterStatusChartOption1 = {
      color: ['#60b515','#c21d00'],
      tooltip: {
        trigger: 'item'
      },
      legend: {
        show: false
      },
      animation: false,
      series: [
        {
          name: 'cluster',
          type: 'pie',
          left: 'center',
          center: ['50%', '50%'],
          radius: ['100%', '130%'],
          avoidLabelOverlap: true,
          itemStyle: {
            borderRadius: 4,
            borderColor: '#fff',
            borderWidth: 2
          },
          label: {
            show: true,
            position: 'center',
            formatter: healthyClusterCount + '/' + (unhealthyClusterCount + healthyClusterCount),
            color: '#323232',
            fontSize: '20'
          },
          data: [{value: healthyClusterCount, name: this.translate.instant('OVERVIEW.HEALTHY')}, {value: unhealthyClusterCount, name: this.translate.instant('OVERVIEW.UNHEALTHY')}]
        }
      ]
    };
    this.clusterStatusChartOption2 = {
      color: ['#60b515','#c21d00'],
      tooltip: {
        trigger: 'item'
      },
      legend: {
        show: false
      },
      animation: false,
      series: [
        {
          name: 'cluster',
          type: 'pie',
          left: 'center',
          center: ['50%', '50%'],
          radius: ['70%', '90%'],
          avoidLabelOverlap: true,
          itemStyle: {
            borderRadius: 4,
            borderColor: '#fff',
            borderWidth: 2
          },
          label: {
            show: true,
            position: 'center',
            formatter: healthyClusterCount + '/' + (unhealthyClusterCount + healthyClusterCount),
            color: '#323232',
            fontSize: '20'
          },
          data: [{value: healthyClusterCount, name: this.translate.instant('OVERVIEW.HEALTHY')}, {value: unhealthyClusterCount, name: this.translate.instant('OVERVIEW.UNHEALTHY')}]
        }
      ]
    };
  }

  clusterHealthyChartInit($event: any): void {
    const clusterHealthyChart = $event;
    window.onresize = function() {
      clusterHealthyChart.resize();
    };
  }

  drawAllTenantCpuGauge(CPUUsagePercentage: number): void {
    const CPUThreshold1 = 60;
    const CPUThreshold2 = 90;
    const CPUThreshold3 = 100;
    this.allTenantCPUOption = {
      series: [
        {
          type: 'gauge',
          center: ['50%', '70%'],
          startAngle: 190,
          endAngle: -10,
          radius: '100%',
          axisLine: {
            show: true,
            lineStyle: {
              width: 4,
              color: [
                [CPUThreshold1/100, '#60b515'],
                [CPUThreshold2/100, '#f58403'],
                [CPUThreshold3/100, '#c21d00']
              ]
            }
          },
          pointer: {
            show: false
          },
          axisTick: {
            show: false
          },
          splitLine: {
            show: true,
            distance: -12,
            length: 5,
            lineStyle: {
              width: 1,
              color: '#999'
            }
          },
          axisLabel: {
            show: true,
            distance: -12,
            color: '#999',
            fontSize: 10
          },
          anchor: {
            show: false
          },
          title: {
            show: false
          },
          detail: {
            valueAnimation: true,
            width: '60%',
            lineHeight: 40,
            borderRadius: 8,
            offsetCenter: [0, '-10%'],
            fontSize: 20,
            fontWeight: 'bolder',
            formatter: '{value}%',
            color: 'auto'
          },
          data: [
            {
              value: CPUUsagePercentage
            }
          ]
        },
        {
          type: 'gauge',
          center: ['50%', '70%'],
          startAngle: 190,
          endAngle: -10,
          radius: '92%',
          pointer: {
            show: false
          },
          itemStyle: {
            color: 'auto'
          },
          progress: {
            show: true,
            width: 20
          },
          axisLine: {
            show: false
          },
          axisTick: {
            show: false
          },
          splitLine: {
            show: false
          },
          axisLabel: {
            show: false
          },
          detail: {
            show: false
          },
          data: [
            {
              value: CPUUsagePercentage
            }
          ]
        }
      ]
    };
    if(CPUUsagePercentage <= CPUThreshold1){
      this.allTenantCPUOption.series[1].itemStyle.color = '#60b515'
    } else if (CPUUsagePercentage <= CPUThreshold2){
      this.allTenantCPUOption.series[1].itemStyle.color = '#f58403'
    } else {
      this.allTenantCPUOption.series[1].itemStyle.color = '#c21d00'
    }

  }

  allTenantCPUChartInit($event: any): void {
    const allTenantCPUChart = $event;
    window.onresize = function() {
      allTenantCPUChart.resize();
    };
  }

  drawAllTenantMemoryGauge(MemoryUsagePercentage: number): void {
    const MemoryThreshold1 = 60;
    const MemoryThreshold2 = 80;
    const MemoryThreshold3 = 100;
    this.allTenantMemoryOption = {
      series: [
        {
          type: 'gauge',
          center: ['50%', '70%'],
          startAngle: 190,
          endAngle: -10,
          radius: '100%',
          axisLine: {
            show: true,
            lineStyle: {
              width: 4,
              color: [
                [MemoryThreshold1/100, '#60b515'],
                [MemoryThreshold2/100, '#f58403'],
                [MemoryThreshold3/100, '#c21d00']
              ]
            }
          },
          pointer: {
            show: false
          },
          axisTick: {
            show: false
          },
          splitLine: {
            show: true,
            distance: -12,
            length: 5,
            lineStyle: {
              width: 1,
              color: '#999'
            }
          },
          axisLabel: {
            show: true,
            distance: -12,
            color: '#999',
            fontSize: 10
          },
          anchor: {
            show: false
          },
          title: {
            show: false
          },
          detail: {
            valueAnimation: true,
            width: '60%',
            lineHeight: 40,
            borderRadius: 8,
            offsetCenter: [0, '-10%'],
            fontSize: 20,
            fontWeight: 'bolder',
            formatter: '{value}%',
            color: 'auto'
          },
          data: [
            {
              value: MemoryUsagePercentage
            }
          ]
        },
        {
          type: 'gauge',
          center: ['50%', '70%'],
          startAngle: 190,
          endAngle: -10,
          radius: '92%',
          pointer: {
            show: false
          },
          itemStyle: {
            color: 'auto'
          },
          progress: {
            show: true,
            width: 20
          },
          axisLine: {
            show: false
          },
          axisTick: {
            show: false
          },
          splitLine: {
            show: false
          },
          axisLabel: {
            show: false
          },
          detail: {
            show: false
          },
          data: [
            {
              value: MemoryUsagePercentage
            }
          ]
        }
      ]
    };
    if(MemoryUsagePercentage <= MemoryThreshold1){
      this.allTenantMemoryOption.series[1].itemStyle.color = '#60b515'
    } else if (MemoryUsagePercentage <= MemoryThreshold2){
      this.allTenantMemoryOption.series[1].itemStyle.color = '#f58403'
    } else {
      this.allTenantMemoryOption.series[1].itemStyle.color = '#c21d00'
    }

  }

  allTenantMemoryChartInit($event: any): void {
    const allTenantMemoryChart = $event;
    window.onresize = function() {
      allTenantMemoryChart.resize();
    };
  }

  drawAllTenantStorageGauge(storageUsagePercentage: number): void {
    const storageThreshold1 = 70;
    const storageThreshold2 = 90;
    const storageThreshold3 = 100;
    this.allTenantStoageOption = {
      series: [
        {
          type: 'gauge',
          center: ['50%', '70%'],
          startAngle: 190,
          endAngle: -10,
          radius: '100%',
          axisLine: {
            show: true,
            lineStyle: {
              width: 4,
              color: [
                [storageThreshold1/100, '#60b515'],
                [storageThreshold2/100, '#f58403'],
                [storageThreshold3/100, '#c21d00']
              ]
            }
          },
          pointer: {
            show: false
          },
          axisTick: {
            show: false
          },
          splitLine: {
            show: true,
            distance: -12,
            length: 5,
            lineStyle: {
              width: 1,
              color: '#999'
            }
          },
          axisLabel: {
            show: true,
            distance: -12,
            color: '#999',
            fontSize: 10
          },
          anchor: {
            show: false
          },
          title: {
            show: false
          },
          detail: {
            valueAnimation: true,
            width: '60%',
            lineHeight: 40,
            borderRadius: 8,
            offsetCenter: [0, '-10%'],
            fontSize: 20,
            fontWeight: 'bolder',
            formatter: '{value}%',
            color: 'auto'
          },
          data: [
            {
              value: storageUsagePercentage
            }
          ]
        },
        {
          type: 'gauge',
          center: ['50%', '70%'],
          startAngle: 190,
          endAngle: -10,
          radius: '92%',
          pointer: {
            show: false
          },
          itemStyle: {
            color: 'auto'
          },
          progress: {
            show: true,
            width: 20
          },
          axisLine: {
            show: false
          },
          axisTick: {
            show: false
          },
          splitLine: {
            show: false
          },
          axisLabel: {
            show: false
          },
          detail: {
            show: false
          },
          data: [
            {
              value: storageUsagePercentage
            }
          ]
        }
      ]
    };
    if(storageUsagePercentage <= storageThreshold1){
      this.allTenantStoageOption.series[1].itemStyle.color = '#60b515'
    } else if (storageUsagePercentage <= storageThreshold2){
      this.allTenantStoageOption.series[1].itemStyle.color = '#f58403'
    } else {
      this.allTenantStoageOption.series[1].itemStyle.color = '#c21d00'
    }

  }

  allTenantStorageChartInit($event: any): void {
    const allTenantStorageChart = $event;
    window.onresize = function() {
      allTenantStorageChart.resize();
    };
  }

  drawAllTenantResourceUsageBar(resourceMetrics: OrgMetrics): void {
    this.allTenantResourceUsageOption = {
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow'
        }
      },
      legend: {
        data: [
          {name: this.translate.instant('OVERVIEW.CPUUSED')},
          {name: this.translate.instant('OVERVIEW.MEMUSED')},
          {name: this.translate.instant('OVERVIEW.STORAGEUSED')},
          {name: this.translate.instant('OVERVIEW.ALLOWANCE')}
        ]
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: [
        {
          type: 'category',
          data: resourceMetrics.name
        }
      ],
      yAxis: [
        {
          type: 'value',
          splitNumber: 10,
          name: 'Core | Gi'
        }
      ],
      series: [
        {
          name: this.translate.instant('OVERVIEW.CPUUSED'),
          type: 'bar',
          barWidth: 'auto',
          color: '#0072a3',
          stack: 'CPU',
          emphasis: {
            focus: 'series'
          },
          data: resourceMetrics.cpuUsed
        },
        {
          name: this.translate.instant('OVERVIEW.ALLOWANCE'),
          type: 'bar',
          stack: 'CPU',
          color: '#c2c2c2',
          emphasis: {
            focus: 'series'
          },
          data: resourceMetrics.cpuFree
        },
        {
          name: this.translate.instant('OVERVIEW.MEMUSED'),
          type: 'bar',
          barWidth: 'auto',
          color: '#60b515',
          stack: 'Memory',
          emphasis: {
            focus: 'series'
          },
          data: resourceMetrics.memoryUsed
        },
        {
          name: this.translate.instant('OVERVIEW.ALLOWANCE'),
          type: 'bar',
          stack: 'Memory',
          color: '#c2c2c2',
          emphasis: {
            focus: 'series'
          },
          data: resourceMetrics.memoryFree
        },
        {
          name: this.translate.instant('OVERVIEW.STORAGEUSED'),
          type: 'bar',
          barWidth: 'auto',
          color: '#f58403',
          stack: 'Storage',
          emphasis: {
            focus: 'series'
          },
          data: resourceMetrics.storageUsed
        },
        {
          name: this.translate.instant('OVERVIEW.ALLOWANCE'),
          type: 'bar',
          stack: 'Storage',
          color: '#c2c2c2',
          emphasis: {
            focus: 'series'
          },
          data: resourceMetrics.storageFree
        }
      ]
    };
  }

  allTenantResourceUsageChartInit($event: any): void {
    const allTenantResourceUsageChart = $event;
    window.onresize = function() {
      allTenantResourceUsageChart.resize();
    };
  }

  drawCPUUsageTop5Bar(orgNameList: string[], orgCpuUsed: number[]): void {
    this.CPUUsageTop5Option = {
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow'
        }
      },
      legend: {
        data: [
          {name: this.translate.instant('OVERVIEW.CPUUSED')}
        ]
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        data: orgNameList
      },
      yAxis: {
        type: 'value',
        name: 'Core'
      },
      series: [
        {
          name: this.translate.instant('OVERVIEW.CPUUSED'),
          barWidth: 'auto',
          color: '#0072a3',
          type: 'bar',
          emphasis: {
            focus: 'series'
          },
          data: orgCpuUsed,
        }
      ]
    };
  }

  cpuUsageTop5ChartInit($event: any): void {
    const CPUUsageTop5Chart = $event;
    window.onresize = function() {
      CPUUsageTop5Chart.resize();
    };
  }

  drawMemoryUsageTop5Bar(orgNameList: string[], orgMemUsed: number[]): void {
    this.MemoryUsageTop5Option = {
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow'
        }
      },
      legend: {
        data: [
          {name: this.translate.instant('OVERVIEW.MEMUSED')}
        ]
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        data: orgNameList
      },
      yAxis: {
        type: 'value',
        name: 'Gi'
      },
      series: [
        {
          name: this.translate.instant('OVERVIEW.MEMUSED'),
          barWidth: 'auto',
          color: '#60b515',
          type: 'bar',
          emphasis: {
            focus: 'series'
          },
          data: orgMemUsed
        }
      ]
    };
  }

  memoryUsageTop5ChartInit($event: any): void {
    const MemoryUsageTop5Chart = $event;
    window.onresize = function() {
      MemoryUsageTop5Chart.resize();
    };
  }

  drawStorageUsageTop5Bar(orgNameList: string[], orgStorageUsed: number[]): void {
    this.StorageUsageTop5Option = {
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow'
        }
      },
      legend: {
        data: [
          {name: this.translate.instant('OVERVIEW.STORAGEUSED')}
        ]
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        data: orgNameList
      },
      yAxis: {
        type: 'value',
        name: 'Gi'
      },
      series: [
        {
          name: this.translate.instant('OVERVIEW.STORAGEUSED'),
          barWidth: 'auto',
          color: '#f58403',
          type: 'bar',
          emphasis: {
            focus: 'series'
          },
          data: orgStorageUsed
        }
      ]
    };
  }

  storageUsageTop5ChartInit($event: any): void {
    const StorageUsageTop5Chart = $event;
    window.onresize = function() {
      StorageUsageTop5Chart.resize();
    };
  }

  showClusterDetails(clusterList: ClusterStatusMetrics[], clusterStatus: string): void {
    this.allCluster = clusterList;
    this.showClusterStatus = clusterStatus;
    this.showClusterDetailsModel = true;
  }

  showExtensionDetails(extensionList: Extension[], extensionName: string): void {
    this.allExtensions = extensionList;
    this.showExtensionName = extensionName;
    this.showExtensionDetailsModel = true;
  }

}
