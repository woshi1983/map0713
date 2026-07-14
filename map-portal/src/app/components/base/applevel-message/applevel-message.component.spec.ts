import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';

import { ApplevelMessageComponent } from './applevel-message.component';

describe('GlobalMessageComponent', () => {
  let component: ApplevelMessageComponent;
  let fixture: ComponentFixture<ApplevelMessageComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [ ApplevelMessageComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ApplevelMessageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
