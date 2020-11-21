import { ComponentFixture, TestBed } from '@angular/core/testing';

import { NavtitleComponent } from './navtitle.component';

describe('NavtitleComponent', () => {
  let component: NavtitleComponent;
  let fixture: ComponentFixture<NavtitleComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ NavtitleComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(NavtitleComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
