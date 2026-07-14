import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ForceDeleteObjectComponent } from './force-delete-object.component';

describe('ForceDeleteObjectComponent', () => {
  let component: ForceDeleteObjectComponent;
  let fixture: ComponentFixture<ForceDeleteObjectComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ForceDeleteObjectComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ForceDeleteObjectComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
