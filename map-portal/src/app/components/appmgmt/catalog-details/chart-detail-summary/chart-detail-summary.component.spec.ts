import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ChartDetailSummaryComponent } from './chart-detail-summary.component';

describe('ChartDetailSummaryComponent', () => {
  let component: ChartDetailSummaryComponent;
  let fixture: ComponentFixture<ChartDetailSummaryComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ChartDetailSummaryComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ChartDetailSummaryComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
