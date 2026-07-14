import { ALERTTYPE } from '../utils/static-entries/shared.const';

export class Message {
   type: ALERTTYPE = ALERTTYPE.INFO;
   text: string;

   public constructor (text: string, type?: ALERTTYPE) {
     this.type = type;
     this.text = text;
   }
}
