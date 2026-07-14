import { Injectable, EventEmitter } from '@angular/core';

@Injectable({
  providedIn: 'root'
})

export class EventbusService {
  deleteObjectEmitter: EventEmitter<boolean> = new EventEmitter();

  // used for refresh orgs
  refreshUserOrgEmitter: EventEmitter<boolean> = new EventEmitter();
}
