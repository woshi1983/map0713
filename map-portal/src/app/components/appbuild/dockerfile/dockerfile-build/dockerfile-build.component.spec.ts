import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DockerfileBuildComponent } from './dockerfile-build.component';

describe('DockerfileBuildComponent', () => {
  let component: DockerfileBuildComponent;
  let fixture: ComponentFixture<DockerfileBuildComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ DockerfileBuildComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(DockerfileBuildComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
