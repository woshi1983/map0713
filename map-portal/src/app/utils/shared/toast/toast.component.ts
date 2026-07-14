import {
  animate,
  keyframes,
  state,
  style,
  transition,
  trigger
} from '@angular/animations';
import { Component } from '@angular/core';

import { Toast, ToastrService, ToastPackage } from 'ngx-toastr';

// options.toastClass
//      class     --   icon              ---  color
// 1) alert-danger -- exclamation-circle --  red
// 2）alert-warning -- exclamation-triangle --   yellow
// 3）alert-info -- info-circle              --- blue
// 4）alert-success -- check-circle          -- green

@Component({
  selector: 'map-toast-component [map-toast-component]',
  styles: [`
    :host {
      position: relative;
      overflow: hidden;
      margin: 10!important;
      padding: 0!important;
      pointer-events: all;
      cursor: pointer;
      min-width: 120px;
    }
  `],
  template: `
  <div class="alert" [class]="options.toastClass" role="alert" [style.display]="state.value === 'inactive' ? 'none' : ''" style="margin: 0;">
    <div class="alert-items">
      <div class="alert-item static">
        <div class="alert-icon-wrapper">
          <cds-icon class="alert-icon" shape="exclamation-circle" *ngIf="options.toastClass === 'alert-danger'"></cds-icon>
          <cds-icon class="alert-icon" shape="exclamation-triangle" *ngIf="options.toastClass === 'alert-warning'"></cds-icon>
          <cds-icon class="alert-icon" shape="info-circle" *ngIf="options.toastClass === 'alert-info'"></cds-icon>
          <cds-icon class="alert-icon" shape="check-circle" *ngIf="options.toastClass === 'alert-success'"></cds-icon>
        </div>
        <div class="alert-text">
          <span style="width: 100%; float: left;" *ngIf="title"><b>{{ title }}</b></span>
          <span *ngIf="message" style="word-break: normal;width: auto; display: block;">{{ message }}</span>
        </div>
      </div>
    </div>
    <button *ngIf="options.closeButton" type="button" class="close" aria-label="Close" (click)="remove()">
      <clr-icon aria-hidden="true" shape="close"></clr-icon>
    </button>
  </div>
  `,
  animations: [
    trigger('flyInOut', [
      state('inactive', style({
        opacity: 0
      })),
      transition('inactive => active', animate('400ms ease-out', keyframes([
        style({
          transform: 'translate3d(100%, 0, 0)',
          opacity: 0
        }),
        style({
          transform: 'none',
          opacity: 1
        })
      ]))),
      transition('active => removed', animate('400ms ease-out', keyframes([
        style({
          opacity: 1
        }),
        style({
          transform: 'translate3d(100%, 0, 0)',
          opacity: 0
        })
      ])))
    ])
  ],
  preserveWhitespaces: false
})
export class MapToast extends Toast {
  // used for demo purposes
  undoString = 'undo';

  // constructor is only necessary when not using AoT
  constructor (
    protected toastrService: ToastrService,
    public toastPackage: ToastPackage
  ) {
    super(toastrService, toastPackage);
  }

  action (event: Event): boolean {
    event.stopPropagation();
    this.undoString = 'undid';
    this.toastPackage.triggerAction();
    return false;
  }
}
