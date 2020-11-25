// Itero - Online iterative vote application
// Copyright (C) 2020 Joseph Boudou
// 
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
// 
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
// 
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
