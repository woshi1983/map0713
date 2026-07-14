import { ComponentFixture, TestBed } from '@angular/core/testing';

import { SubscribeMapdLogsComponent } from './subscribe-mapd-logs.component';

describe('SubscribeMapdLogsComponent', () => {
  let component: SubscribeMapdLogsComponent;
  let fixture: ComponentFixture<SubscribeMapdLogsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ SubscribeMapdLogsComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(SubscribeMapdLogsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
