import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';

import { LoadbalancerComponent } from './loadbalancer.component';

describe('GlbComponent', () => {
  let component: LoadbalancerComponent;
  let fixture: ComponentFixture<LoadbalancerComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [ LoadbalancerComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(LoadbalancerComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
