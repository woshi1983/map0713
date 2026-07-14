import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AppcomponentDetailsComponent } from './appcomponent-details.component';

describe('ComponentDetailsComponent', () => {
  let component: AppcomponentDetailsComponent;
  let fixture: ComponentFixture<AppcomponentDetailsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ AppcomponentDetailsComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(AppcomponentDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
