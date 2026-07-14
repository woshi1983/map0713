import { ComponentFixture, TestBed } from '@angular/core/testing';

import { BinaryBuildComponent } from './binary-build.component';

describe('BinaryBuildComponent', () => {
  let component: BinaryBuildComponent;
  let fixture: ComponentFixture<BinaryBuildComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ BinaryBuildComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(BinaryBuildComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
