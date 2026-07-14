import { Component, OnInit, Input, ChangeDetectionStrategy } from '@angular/core';

@Component({
  selector: 'app-chart-detail-values',
  templateUrl: './chart-detail-values.component.html',
  styleUrls: ['./chart-detail-values.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ChartDetailValuesComponent implements OnInit {
  @Input() values;
  @Input() yaml;

  // Default set to yaml file
  valueMode = false;
  valueHover = false;
  yamlHover = true;

  objKeys = Object.keys;

  constructor() { }

  ngOnInit(): void {
  }

  public get isValueMode() {
    return this.valueMode;
  }

  isHovering(view: string) {
    if (view === 'value') {
      return this.valueHover;
    } else {
      return this.yamlHover;
    }
  }

  showYamlFile(showYaml: boolean) {
    this.valueMode = !showYaml;
  }

  mouseEnter(mode: string) {
    if (mode === "value") {
      this.valueHover = true;
    } else {
      this.yamlHover = true;
    }
  }

  mouseLeave(mode: string) {
    if (mode === "value") {
      this.valueHover = false;
    } else {
      this.yamlHover = false;
    }
  }


}
