import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ChartDetailValuesComponent } from './chart-detail-values.component';

describe('ChartDetailValuesComponent', () => {
  let component: ChartDetailValuesComponent;
  let fixture: ComponentFixture<ChartDetailValuesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ChartDetailValuesComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ChartDetailValuesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
