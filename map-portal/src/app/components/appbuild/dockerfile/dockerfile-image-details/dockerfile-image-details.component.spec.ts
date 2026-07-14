import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DockerfileImageDetailsComponent } from './dockerfile-image-details.component';

describe('BuildImageDetailsComponent', () => {
  let component: DockerfileImageDetailsComponent;
  let fixture: ComponentFixture<DockerfileImageDetailsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ DockerfileImageDetailsComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(DockerfileImageDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
