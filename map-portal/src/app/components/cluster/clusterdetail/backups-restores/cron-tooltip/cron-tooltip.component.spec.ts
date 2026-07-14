import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CronTooltipComponent } from './cron-tooltip.component';

describe('CronTooltipComponent', () => {
  let component: CronTooltipComponent;
  let fixture: ComponentFixture<CronTooltipComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CronTooltipComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(CronTooltipComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
