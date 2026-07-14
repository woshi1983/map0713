import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';

import { ClusterdetailComponent } from './clusterdetail.component';

describe('ClusterdetailComponent', () => {
  let component: ClusterdetailComponent;
  let fixture: ComponentFixture<ClusterdetailComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [ ClusterdetailComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ClusterdetailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
