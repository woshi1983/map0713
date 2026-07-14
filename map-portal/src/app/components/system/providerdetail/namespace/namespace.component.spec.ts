import { ComponentFixture, TestBed } from '@angular/core/testing';

import { SupervisorNamespaceComponent } from './namespace.component';

describe('SupervisorNamespaceComponent', () => {
  let component: SupervisorNamespaceComponent;
  let fixture: ComponentFixture<SupervisorNamespaceComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ SupervisorNamespaceComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(SupervisorNamespaceComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
