import { ComponentFixture, TestBed } from '@angular/core/testing';

import { SourceBuildComponent } from './source-build.component';

describe('DockerfileBuildComponent', () => {
  let component: SourceBuildComponent;
  let fixture: ComponentFixture<SourceBuildComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ SourceBuildComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(SourceBuildComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
