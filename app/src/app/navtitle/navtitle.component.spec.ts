import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';

import { Subject } from 'rxjs';

import { SessionService, SessionInfo } from '../session.service';
import { NavtitleComponent } from './navtitle.component';

describe('NavtitleComponent', () => {
  let component: NavtitleComponent;
  let fixture: ComponentFixture<NavtitleComponent>;

  beforeEach(async () => {
    const sessionSpy = jasmine.createSpyObj('SessionService', ['checkSession', 'login']);
    const routerSpy  = jasmine.createSpyObj('Router', ['navigateByUrl']);

    await TestBed.configureTestingModule({
      declarations: [ NavtitleComponent ],
      providers: [
        {provide: SessionService, useValue: sessionSpy},
        {provide: Router, useValue: routerSpy}
      ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(NavtitleComponent);
    component = fixture.componentInstance;

    // Very harsh way of doing it...
    let router = fixture.debugElement.injector.get(SessionService);
    router.observable = new Subject<SessionInfo>();

    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
