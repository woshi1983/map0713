import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';

import { RepositorysComponent } from './repositorys.component';

describe('RepositorysComponent', () => {
  let component: RepositorysComponent;
  let fixture: ComponentFixture<RepositorysComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [ RepositorysComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(RepositorysComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
