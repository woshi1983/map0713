import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AppcomponentsComponent } from './appcomponents.component';

describe('ComponentsComponent', () => {
  let component: AppcomponentsComponent;
  let fixture: ComponentFixture<AppcomponentsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ AppcomponentsComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(AppcomponentsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
