import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AppInstancesComponent } from './app-instances.component';

describe('AppInstancesComponent', () => {
  let component: AppInstancesComponent;
  let fixture: ComponentFixture<AppInstancesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ AppInstancesComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(AppInstancesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
