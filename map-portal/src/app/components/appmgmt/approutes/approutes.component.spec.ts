import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ApproutesComponent } from './approutes.component';

describe('ApproutesComponent', () => {
  let component: ApproutesComponent;
  let fixture: ComponentFixture<ApproutesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ApproutesComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ApproutesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
