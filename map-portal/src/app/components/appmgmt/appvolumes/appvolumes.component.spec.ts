import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AppvolumesComponent } from './appvolumes.component';

describe('AppvolumesComponent', () => {
  let component: AppvolumesComponent;
  let fixture: ComponentFixture<AppvolumesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ AppvolumesComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(AppvolumesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
