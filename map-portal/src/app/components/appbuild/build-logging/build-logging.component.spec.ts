import { ComponentFixture, TestBed } from '@angular/core/testing';

import { BuildLoggingComponent } from './build-logging.component';

describe('LoggingComponent', () => {
  let component: BuildLoggingComponent;
  let fixture: ComponentFixture<BuildLoggingComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [BuildLoggingComponent]
    })
      .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(BuildLoggingComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
