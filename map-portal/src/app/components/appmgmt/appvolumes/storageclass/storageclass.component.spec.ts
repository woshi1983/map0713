import { ComponentFixture, TestBed } from '@angular/core/testing';

import { StorageclassComponent } from './storageclass.component';

describe('StorageclassComponent', () => {
  let component: StorageclassComponent;
  let fixture: ComponentFixture<StorageclassComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ StorageclassComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(StorageclassComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
