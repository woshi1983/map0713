import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ClusterWorkloadsComponent } from './cluster-workloads.component';

describe('ClusterWorkloadsComponent', () => {
  let component: ClusterWorkloadsComponent;
  let fixture: ComponentFixture<ClusterWorkloadsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ClusterWorkloadsComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ClusterWorkloadsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
