import { ComponentFixture, TestBed } from '@angular/core/testing';

import { BinaryImageDetailsComponent } from './binary-image-details.component';

describe('BuildImageDetailsComponent', () => {
  let component: BinaryImageDetailsComponent;
  let fixture: ComponentFixture<BinaryImageDetailsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ BinaryImageDetailsComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(BinaryImageDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
