import { Directive, OnInit, Input, TemplateRef, ViewContainerRef } from '@angular/core';
import { environment } from '../../../../environments/environment';

@Directive({
  selector: '[featureToggle]'
})
export class FeaturesToggleDirective implements OnInit {
  @Input() featureToggle: string;

  constructor (
    private templateRef: TemplateRef<any>,
    private viewContainer: ViewContainerRef
  ) {}

  ngOnInit (): void {
    if (this.isEnabled()) {
      this.viewContainer.createEmbeddedView(this.templateRef);
    } else {
      this.viewContainer.clear();
    }
  }

  isEnabled (): boolean {
    const features = environment.features;
    if (environment.features.hasOwnProperty(this.featureToggle)) {
      return features[this.featureToggle];
    }
    return true;
  }
}
