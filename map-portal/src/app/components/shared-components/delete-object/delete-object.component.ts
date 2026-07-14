import { Component, EventEmitter, Output } from '@angular/core';
import { DeleteConfirmationAction, DeleteObjectEvent } from '../../../classes';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { CONFIRMATIONSTATE, DELETEOBJECTBUTTON } from '../../../utils/static-entries/shared.const';

@Component({
  selector: 'app-delete-object',
  templateUrl: './delete-object.component.html',
  styleUrls: ['./delete-object.component.css']
})
export class DeleteObjectComponent {
  opened = false;
  event: DeleteObjectEvent;
  buttons: DELETEOBJECTBUTTON

  @Output() confirmAction = new EventEmitter<DeleteConfirmationAction>();
  @Output() cancelAction = new EventEmitter<DeleteConfirmationAction>();

  deleteConfirmForm = new FormGroup({
    confirm: new FormControl('', [Validators.required])
  });

  open (event: DeleteObjectEvent): void {
    this.buttons = event.button;
    this.event = event;
    this.opened = true;
  }

  close (): void {
    this.event = null;
    this.buttons = null;
    this.opened = false;
  }

  cancel (): void {
    const cancelMessage: DeleteConfirmationAction = {
      state: CONFIRMATIONSTATE.CANCEL,
      objectType: this.event?.objectType
    }
    this.cancelAction.emit(cancelMessage);
    this.deleteConfirmForm.reset()
    this.close();
  }

  confirm (): void {
    const confirmMessage: DeleteConfirmationAction = {
      state: CONFIRMATIONSTATE.CONFIRMED,
      objectType: this.event?.objectType
    }
    this.confirmAction.emit(confirmMessage);
    this.deleteConfirmForm.reset()
    this.close();
  }
}
