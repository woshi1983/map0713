import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CatrustComponent } from './catrust.component';

describe('CatrustComponent', () => {
  let component: CatrustComponent;
  let fixture: ComponentFixture<CatrustComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CatrustComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(CatrustComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
