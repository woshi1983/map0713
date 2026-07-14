import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DeleteObjectComponent } from './delete-object.component';

describe('DeleteObjectComponent', () => {
  let component: DeleteObjectComponent;
  let fixture: ComponentFixture<DeleteObjectComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ DeleteObjectComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(DeleteObjectComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
