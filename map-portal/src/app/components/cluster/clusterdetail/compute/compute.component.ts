import {Component, OnInit, ViewChild, Input} from '@angular/core';
import {AddNodeTaint, DrainNode, handleObject, Node, RemoveNodeTaint, UpdateNode, DrainNodeEvent, NodeResource} from '../../../../classes';
import { AuthService } from '../../../../services/auth.service';
import {ActivatedRoute} from "@angular/router";
import {ToastrService} from "ngx-toastr";
import {TranslateService} from "@ngx-translate/core";
import {NodesService} from "../../../../services/nodes.service";
import {ClrDatagridPagination} from "@clr/angular";
import {convertCPUUnit, convertMemUsageUnit, convertMemCapUnit} from '../../../../utils/operator';
import {OrganizationService} from "../../../../services/organization.service";
import {FormArray, FormControl, FormGroup, Validators} from "@angular/forms";
import {CustomValidatorDirective} from "../../../../utils/validators/custom-validator.directive";
import {Subscription} from "rxjs";
import {EChartsOption} from "echarts";

@Component({
  selector: 'app-cluster-compute',
  templateUrl: './compute.component.html',
  styleUrls: ['./compute.component.css']
})
export class ComputeComponent implements OnInit {

  //admin and user role can handle cluster resources
  get canWriteResources(): boolean {
    return this.authService.can(this.orgId, handleObject.CLUSTER_RESOURCE);
  }

  orgId: string;
  allNodes: Node[] = [];
  selected: Node;
  @Input() clusterId;

  showSearch = false;
  searchName: string;
  loading = false;
  showNodeCordonModel = false;
  @ViewChild('pagination') pagination: ClrDatagridPagination;

  nodeCPUOption: EChartsOption;
  nodeMemoryOption: EChartsOption;
  nodeCPURequestOption: EChartsOption;
  nodeMemoryRequestOption: EChartsOption;
  nodeCPULimitOption: EChartsOption;
  nodeMemoryLimitOption: EChartsOption;
  loadingNodeResource = false;
  showNodeResourceModel = false;
  nodeResource: NodeResource;

  readonly disAllowNodeLabels = CustomValidatorDirective.disAllowNodeLabels;
  updateLabelOpen = false;
  updateNodeLabels = new FormArray([]);
  updateNodeLabelForm = new FormGroup({
    labels: this.updateNodeLabels
  });

  readonly disAllowNodeTaint = CustomValidatorDirective.disAllowNodeTaint;
  removeNodeTaint: RemoveNodeTaint[] = [];
  updateTaintOpen = false;
  updateNodeTaints = new FormArray([]);
  updateNodeTaintForm = new FormGroup({
    taints: this.updateNodeTaints
  });

  subscription: Subscription;
  loadingNodeDrainEvent = false;
  DrainNodeEventList: DrainNodeEvent[] = [];
  drainNodeOpen = false;
  showDrainEvent = false;
  drainNodeForm = new FormGroup({
    timeout: new FormControl(15, [Validators.min(1)])
  });

  constructor(
    public authService: AuthService,
    private route: ActivatedRoute,
    private toast: ToastrService,
    private translate: TranslateService,
    private nodeService: NodesService,
    private orgService: OrganizationService,
    ) {
    this.orgId = this.orgService.getSelectedOrgId;
    }

  ngOnInit(): void {
  }

  refresh(): void {
    this.loading = true;
    this.allNodes = [];
    const filters = [];
    filters.push({ key: 'cluster', value: [this.clusterId] });
    if(!!this.searchName){
      filters.push({key: 'name', value: [this.searchName]});
    }
    this.nodeService.getAllNodes(filters).subscribe(data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.GET-NODES-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if(!!data?.data?.Nodes){
        const nodeList = data?.data?.Nodes;
        for(const node of nodeList){
          this.allNodes.push({
            name: node.name,
            labels: node.labels,
            spec: node.spec,
            status: {
              kernelVersion: node.status.kernelVersion,
              ready: node.status.ready,
              osImage: node.status.osImage,
              internalIP: node.status.internalIP,
              externalIP: node.status.externalIP
            },
            capacity: {
              cpu: convertCPUUnit(node.status.capacity.cpu_number, node.status.capacity.cpu_suffix),
              memory: convertMemCapUnit(node.status.capacity.memory_number, node.status.capacity.memory_suffix)
            },
            createdAt: node.createdAt,
            lastModified: node.lastModified,
          })
        }
      } else {
        this.allNodes = [];
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.GET-NODES-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => this.loading = false);
  }

  showNodeResource(nodeName: string): void {
    this.showNodeResourceModel = true;
    this.loadingNodeResource = true;
    this.nodeService.getNodeResource(this.clusterId, nodeName).subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.GETERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.GET-NODES-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if(!!data?.data?.Nodes){
        const nodeData = data.data.Nodes[0];
        this.nodeResource = {
          name: nodeName,
          cpu: {
            used: !!nodeData.status.usage ? convertCPUUnit(nodeData.status.usage.cpu_number, nodeData.status.usage.cpu_suffix) : 0,
            total: convertCPUUnit(nodeData.status.capacity.cpu_number, nodeData.status.capacity.cpu_suffix),
            allocatable: convertCPUUnit(nodeData.status.allocatable.cpu_number, nodeData.status.allocatable.cpu_suffix)
          },
          memory: {
            used: !!nodeData.status.usage ? convertMemUsageUnit(nodeData.status.usage.memory_number, nodeData.status.usage.memory_suffix) : 0,
            total: convertMemCapUnit(nodeData.status.capacity.memory_number, nodeData.status.capacity.memory_suffix),
            allocatable: convertMemCapUnit(nodeData.status.allocatable.memory_number, nodeData.status.allocatable.memory_suffix),
          }
        }
        if(!!nodeData.status.allocated?.requests){
          this.nodeResource.cpu.request = !!nodeData.status.allocated.requests?.cpu_number ? convertCPUUnit(nodeData.status.allocated.requests.cpu_number, nodeData.status.allocated.requests.cpu_suffix) : 0;
          this.nodeResource.memory.request = !!nodeData.status.allocated.requests?.memory_number ? convertMemUsageUnit(nodeData.status.allocated.requests.memory_number, nodeData.status.allocated.requests.memory_suffix) : 0;
        } else {
          this.nodeResource.cpu.request = 0;
          this.nodeResource.memory.request = 0;
        }
        if(!!nodeData.status.allocated?.limits){
          this.nodeResource.cpu.limit = !!nodeData.status.allocated.limits?.cpu_number ? convertCPUUnit(nodeData.status.allocated.limits.cpu_number, nodeData.status.allocated.limits.cpu_suffix) : 0;
          this.nodeResource.memory.limit = !!nodeData.status.allocated.limits?.memory_number ? convertMemUsageUnit(nodeData.status.allocated.limits.memory_number, nodeData.status.allocated.limits.memory_suffix) : 0;
        } else {
          this.nodeResource.cpu.limit = 0;
          this.nodeResource.memory.limit = 0;
        }
        const cpuUsagePercent = Math.round((Number((this.nodeResource.cpu.used / this.nodeResource.cpu.total)) * 100 + Number.EPSILON) * 100) /100;
        const memUsagePercent = Math.round((Number((this.nodeResource.memory.used / this.nodeResource.memory.total)) * 100 + Number.EPSILON) * 100) /100;
        const cpuRequestPercent = Math.round((Number((this.nodeResource.cpu.request / this.nodeResource.cpu.allocatable)) * 100 + Number.EPSILON) * 100) /100;
        const memRequestPercent = Math.round((Number((this.nodeResource.memory.request / this.nodeResource.memory.allocatable)) * 100 + Number.EPSILON) * 100) /100;
        const cpuLimitPercent = Math.round((Number((this.nodeResource.cpu.limit / this.nodeResource.cpu.allocatable)) * 100 + Number.EPSILON) * 100) /100;
        const memLimitPercent = Math.round((Number((this.nodeResource.memory.limit / this.nodeResource.memory.allocatable)) * 100 + Number.EPSILON) * 100) /100;
        this.drawNodeCpuGauge(cpuUsagePercent);
        this.drawNodeMemoryGauge(memUsagePercent);
        this.drawNodeCpuRequestGauge(cpuRequestPercent);
        this.drawNodeMemoryRequestGauge(memRequestPercent);
        this.drawNodeCpuLimitGauge(cpuLimitPercent);
        this.drawNodeMemoryLimitGauge(memLimitPercent);
        this.loadingNodeResource = false;
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.GET-NODES-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}))
  }

  doCloseNodeResource(): void {
    this.showNodeResourceModel = false;
  }

  drawNodeCpuGauge(CPUUsagePercentage: number): void {
    const CPUThreshold1 = 60;
    const CPUThreshold2 = 90;
    const CPUThreshold3 = 100;
    this.nodeCPUOption = {
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
            color: 'inherit'
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
            color: 'inherit'
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
      this.nodeCPUOption.series[1].itemStyle.color = '#60b515'
    } else if (CPUUsagePercentage <= CPUThreshold2){
      this.nodeCPUOption.series[1].itemStyle.color = '#f58403'
    } else {
      this.nodeCPUOption.series[1].itemStyle.color = '#c21d00'
    }
  }

  nodeCPUChartInit($event: any): void {
    const nodeCPUChart = $event;
    window.onresize = function() {
      nodeCPUChart.resize();
    };
  }

  drawNodeMemoryGauge(MemoryUsagePercentage: number): void {
    const MemoryThreshold1 = 60;
    const MemoryThreshold2 = 80;
    const MemoryThreshold3 = 100;
    this.nodeMemoryOption = {
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
            color: 'inherit'
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
            color: 'inherit'
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
      this.nodeMemoryOption.series[1].itemStyle.color = '#60b515'
    } else if (MemoryUsagePercentage <= MemoryThreshold2){
      this.nodeMemoryOption.series[1].itemStyle.color = '#f58403'
    } else {
      this.nodeMemoryOption.series[1].itemStyle.color = '#c21d00'
    }

  }

  nodeMemoryChartInit($event: any): void {
    const nodeMemoryChart = $event;
    window.onresize = function() {
      nodeMemoryChart.resize();
    };
  }

  drawNodeCpuRequestGauge(CPURequestPercentage: number): void {
    const CPUThreshold1 = 90;
    const CPUThreshold2 = 100;
    this.nodeCPURequestOption = {
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
                [CPUThreshold2/100, '#c21d00']
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
            color: 'inherit'
          },
          data: [
            {
              value: CPURequestPercentage
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
            color: 'inherit'
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
              value: CPURequestPercentage
            }
          ]
        }
      ]
    };
    if(CPURequestPercentage <= CPUThreshold1){
      this.nodeCPURequestOption.series[1].itemStyle.color = '#60b515'
    } else {
      this.nodeCPURequestOption.series[1].itemStyle.color = '#c21d00'
    }
  }

  nodeCPURequestChartInit($event: any): void {
    const nodeCPURequestChart = $event;
    window.onresize = function() {
      nodeCPURequestChart.resize();
    };
  }

  drawNodeMemoryRequestGauge(MemoryRequestPercentage: number): void {
    const MemThreshold1 = 90;
    const MemThreshold2 = 100;
    this.nodeMemoryRequestOption = {
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
                [MemThreshold1/100, '#60b515'],
                [MemThreshold2/100, '#c21d00']
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
            color: 'inherit'
          },
          data: [
            {
              value: MemoryRequestPercentage
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
            color: 'inherit'
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
              value: MemoryRequestPercentage
            }
          ]
        }
      ]
    };
    if(MemoryRequestPercentage <= MemThreshold1){
      this.nodeMemoryRequestOption.series[1].itemStyle.color = '#60b515'
    } else {
      this.nodeMemoryRequestOption.series[1].itemStyle.color = '#c21d00'
    }
  }

  nodeMemoryRequestChartInit($event: any): void {
    const nodeMemoryRequestChart = $event;
    window.onresize = function() {
      nodeMemoryRequestChart.resize();
    };
  }

  drawNodeCpuLimitGauge(CPULimitPercentage: number): void {
    const CPUThreshold1 = 100;
    this.nodeCPULimitOption = {
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
                [CPUThreshold1/100, '#60b515']
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
            color: 'inherit'
          },
          data: [
            {
              value: CPULimitPercentage
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
            color: 'inherit'
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
              value: CPULimitPercentage
            }
          ]
        }
      ]
    };
    if(CPULimitPercentage <= CPUThreshold1){
      this.nodeCPULimitOption.series[1].itemStyle.color = '#60b515'
    } else {
      this.nodeCPULimitOption.series[1].itemStyle.color = '#c21d00'
    }
  }

  nodeCPULimitChartInit($event: any): void {
    const nodeCPULimitChart = $event;
    window.onresize = function() {
      nodeCPULimitChart.resize();
    };
  }

  drawNodeMemoryLimitGauge(MemoryLimitPercentage: number): void {
    const MemThreshold1 = 100;
    this.nodeMemoryLimitOption = {
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
                [MemThreshold1/100, '#60b515']
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
            color: 'inherit'
          },
          data: [
            {
              value: MemoryLimitPercentage
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
            color: 'inherit'
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
              value: MemoryLimitPercentage
            }
          ]
        }
      ]
    };
    if(MemoryLimitPercentage <= MemThreshold1){
      this.nodeMemoryLimitOption.series[1].itemStyle.color = '#60b515'
    } else {
      this.nodeMemoryLimitOption.series[1].itemStyle.color = '#c21d00'
    }
  }

  nodeMemoryLimitChartInit($event: any): void {
    const nodeMemoryLimitChart = $event;
    window.onresize = function() {
      nodeMemoryLimitChart.resize();
    };
  }


  showSetNodeCrodon(): void {
    this.showNodeCordonModel = true;
  }

  //TODO: node.unschedulable is error.
  updateNodeCordon(node: Node): void {
    const input: UpdateNode = {
      name: node.name,
      spec: {
        unschedulable: !node.spec.unschedulable
      },
      cluster: this.clusterId
    }
    this.nodeService.updateNode(input).subscribe( data => {
      if (!!data.errors){
        for (const each of data.errors) {
          if (!!each.message) {
            this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.UPDATE-NODES-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
          }
        }
      }
      if(!!data.data?.Node){
        this.toast.show(`${data.data.Node.name} `+this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.UPDATE-NODES-MSG'), this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.UPDATE-NODES-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.UPDATE-NODES-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.refresh();
      this.showNodeCordonModel = false;
      this.selected = null;
    })
  }

  showUpdateNodeLabels(node: Node): void {
    this.updateLabelOpen = true;
    if (!!node.labels) {
      const labels = node.labels;
      for (const key of Object.keys(labels)) {
        this.updateNodeLabels.push(new FormGroup({
          key: new FormControl(key),
          value: new FormControl(labels[key])
        }));
      }
    }
  }

  addupdatelabel(): void {
    this.updateNodeLabels.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl('')
    }));
  }

  deleteupdatelable(index: number): void{
    this.updateNodeLabels.removeAt(index);
  }

  updateNodeLables(node: Node): void {
   const nodeParam: UpdateNode = {
      name: node.name,
      cluster: this.clusterId
    };
    const labels: JSON = JSON.parse(JSON.stringify({}));
    if (!!this.updateNodeLabelForm.value.labels) {
      for (const each of this.updateNodeLabelForm.value.labels) {
        labels[each.key] = each.value;
      }
      nodeParam.labels = labels;
    }
    this.nodeService.updateNode(nodeParam).subscribe((data) => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.UPDATE-NODES-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if(data?.data?.UpdateNode != null){
          this.toast.show(`${data?.data.UpdateNode.name} `+this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.UPDATE-NODES-MSG'), this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.UPDATE-NODES-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.UPDATE-NODES-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doUpdateLabelsReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doUpdateLabelsReset(): void {
    this.updateNodeLabelForm.reset();
    this.updateLabelOpen = false;
    while (this.updateNodeLabels.length !== 0) {
      this.updateNodeLabels.removeAt(0);
    }
  }

  showUpdateNodeTaints(node: Node): void {
    this.updateTaintOpen = true;
    if (!!node.spec.taints) {
      for (const each of node.spec.taints) {
        this.updateNodeTaints.push(new FormGroup({
          key: new FormControl(each.key),
          value: new FormControl(each.value),
          effect: new FormControl(each.effect)
        }));
        this.removeNodeTaint.push({
          key: each.key
        })
      }
    }
  }

  addUpdateTaint(): void {
    this.updateNodeTaints.push(new FormGroup({
      key: new FormControl('', Validators.required),
      value: new FormControl(''),
      effect: new FormControl('', Validators.required)
    }));
  }

  deleteUpdateTaint(index: number): void{
    this.updateNodeTaints.removeAt(index);
  }

  doUpdateNodeTaints(node: Node): void {
    const addTaintList: AddNodeTaint[] = [];
    if (!!this.updateNodeTaintForm.value.taints) {
      for (const each of this.updateNodeTaintForm.value.taints) {
        addTaintList.push({
          key: each.key,
          value: each.value,
          effect: each.effect
        });
      }
    }
    const nodeParam: UpdateNode = {
      name: node.name,
      cluster: this.clusterId,
      spec: {
        addTaints: addTaintList,
        removeTaints: this.removeNodeTaint
      }
    };
    this.nodeService.updateNode(nodeParam).subscribe((data) => {
      if (!!data.errors){
        for (const each of data.errors){
           if(!!each.message){
             this.toast.show( this.translate.instant('MESSAGE.SHARED.UPDATEERROR')+` ${each.message} ${!!each.extensions?.reason ? ' : ' + each.extensions?.reason : '' }`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.UPDATE-NODES-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0})
           }
        }
      }
      if(data?.data?.UpdateNode != null){
          this.toast.show(`${data?.data.UpdateNode.name} `+this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.UPDATE-NODES-MSG'), this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.UPDATE-NODES-SUCCESS'), {toastClass: 'alert-success', closeButton: true, extendedTimeOut: 10000, timeOut: 10000})
      }
    }, (e) => this.toast.show( this.translate.instant('MESSAGE.SHARED.UNGET')+` ${e}`, this.translate.instant('MESSAGE.CLUSTER-DETAILD.COMPUTE.UPDATE-NODES-FAILED'), {toastClass: 'alert-danger', closeButton: true, extendedTimeOut: 0, timeOut: 0}), () => {
      this.doUpdateTaintsReset();
      if (this.pagination.currentPage === 1) {
        this.refresh();
      } else {
        this.pagination.currentPage = 1;
      }
    });
  }

  doUpdateTaintsReset(): void {
    this.updateNodeTaintForm.reset();
    this.updateTaintOpen = false;
    while (this.updateNodeTaints.length !== 0) {
      this.updateNodeTaints.removeAt(0);
    }
  }

  showDrainNode(): void {
    this.drainNodeOpen = true;
  }

  doDrainNode(node: Node): void {
    this.showDrainEvent = true;
    this.loadingNodeDrainEvent = true;
    const input: DrainNode = {
      name: node.name,
      cluster: this.clusterId,
      timeout: !!this.drainNodeForm.value.timeout ? 'PT'+this.drainNodeForm.value.timeout+'M' : 'PT15M'
    }
    this.subscription = this.nodeService.listenDrainNode(input).subscribe(data => {
      this.loadingNodeDrainEvent = false;
      this.DrainNodeEventList.push(data.data.DrainNode);
    }, (e) => {
      console.log(e)
    }, () => {})
  }

  doDrainNodeCancel(): void {
    this.drainNodeOpen = false;
    this.drainNodeForm.reset();
    this.drainNodeForm.patchValue({
      timeout: 15
    })
  }

  stopDrainNode(): void {
    this.doDrainNodeCancel();
    this.showDrainEvent = false;
    this.DrainNodeEventList = [];
    this.subscription.unsubscribe();
    this.refresh();
  }

}
