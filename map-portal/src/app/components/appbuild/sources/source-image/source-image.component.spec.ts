import { ComponentFixture, TestBed } from '@angular/core/testing';

import { SourceImageComponent } from './source-image.component';

describe('SourceBuildComponent', () => {
  let component: SourceImageComponent;
  let fixture: ComponentFixture<SourceImageComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ SourceImageComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(SourceImageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
