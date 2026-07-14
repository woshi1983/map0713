import { ComponentFixture, TestBed } from '@angular/core/testing';

import { SourceImageDetailsComponent } from './source-image-details.component';

describe('BuildImageDetailsComponent', () => {
  let component: SourceImageDetailsComponent;
  let fixture: ComponentFixture<SourceImageDetailsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ SourceImageDetailsComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(SourceImageDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
