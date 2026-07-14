import { DELETEOBJECTTYPE, DELETEOBJECTBUTTON, CONFIRMATIONSTATE, ALERTTYPE } from '../utils/static-entries/shared.const';

// modalNames are 'USER_PROFILE' / 'CHANGE_PWD' / 'ABOUT'
export class ModalEvent {
  modalName: string;
  modalFlag: boolean; // true for open, false for close
}

// modalNames are 'USER_PROFILE' / 'CHANGE_PWD' / 'ABOUT'
export class DeleteAlert {
  type: ALERTTYPE;
  message: string;
}

export class DeleteObjectEvent {
  title: string;
  message: string;
  formInputLabel?: string;
  alerts?: DeleteAlert[];
  objectName?: string;
  objectType: DELETEOBJECTTYPE;
  button: DELETEOBJECTBUTTON;
}

export class DeleteConfirmationAction {
  state: CONFIRMATIONSTATE = CONFIRMATIONSTATE.NA;
  objectType: DELETEOBJECTTYPE;
}
