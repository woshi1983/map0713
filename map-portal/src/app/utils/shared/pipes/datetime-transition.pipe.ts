import { Pipe, PipeTransform } from '@angular/core';
import { DatePipe } from '@angular/common';
import { LANG_LOCALSTORAGE_KEY, DeFaultLang } from '../../static-entries/shared.const';

const baseTimeLine: Date = new Date('1970-1-1');

@Pipe({
  name: 'datetimeTransition',
  pure: false
})
export class DatetimeTransitionPipe implements PipeTransform {
  // can set format like: <date> | datetimeTransition: 'medium'  or  <date> | datetimeTransition:'MM/dd/y HH:mm:ss'
  // mode refer: https://angular.io/api/common/DatePipe#pre-defined-format-options
  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  transform (value: any, format?: string): string {
    let lang: string = DeFaultLang;
    if (localStorage && localStorage.getItem(LANG_LOCALSTORAGE_KEY)) {
      lang = localStorage.getItem(LANG_LOCALSTORAGE_KEY);
    }
    if (value && value <= baseTimeLine) { // invalid date
      return '-';
    }
    // default format short
    return new DatePipe(lang).transform(value, format || 'short');
  }
}
