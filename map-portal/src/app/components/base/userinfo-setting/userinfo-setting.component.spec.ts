import { ComponentFixture, TestBed } from '@angular/core/testing';

import { UserinfoSettingComponent } from './userinfo-setting.component';

describe('UserinfoSettingComponent', () => {
  let component: UserinfoSettingComponent;
  let fixture: ComponentFixture<UserinfoSettingComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ UserinfoSettingComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(UserinfoSettingComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
