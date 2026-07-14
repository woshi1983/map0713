import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DockerfileImageComponent } from './dockerfile-image.component';

describe('DockerfileBuildComponent', () => {
  let component: DockerfileImageComponent;
  let fixture: ComponentFixture<DockerfileImageComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ DockerfileImageComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(DockerfileImageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
