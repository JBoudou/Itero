import { ComponentFixture, TestBed } from '@angular/core/testing';

import { NavtreeComponent } from './navtree.component';

describe('NavtreeComponent', () => {
  let component: NavtreeComponent;
  let fixture: ComponentFixture<NavtreeComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ NavtreeComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(NavtreeComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
