import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { BackupsRestoresComponent } from './backups-restores.component';

describe('BackupsRestoresComponent', () => {
  let component: BackupsRestoresComponent;
  let fixture: ComponentFixture<BackupsRestoresComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ BackupsRestoresComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(BackupsRestoresComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
