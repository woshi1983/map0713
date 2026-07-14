import { ComponentFixture, TestBed } from '@angular/core/testing';

import { UserNoOrgComponent } from './user-no-org.component';

describe('OrginitComponent', () => {
  let component: UserNoOrgComponent;
  let fixture: ComponentFixture<UserNoOrgComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ UserNoOrgComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(UserNoOrgComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
