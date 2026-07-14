import { ComponentFixture, TestBed } from '@angular/core/testing';

import { BinaryImageComponent } from './binary-image.component';

describe('BinaryBuildComponent', () => {
  let component: BinaryImageComponent;
  let fixture: ComponentFixture<BinaryImageComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ BinaryImageComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(BinaryImageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
