// Itero - Online iterative vote application
// Copyright (C) 2021 Joseph Boudou
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

import { TestBed } from '@angular/core/testing';
import { ActivatedRouteSnapshot, RouterStateSnapshot, Router } from '@angular/router';

import { LoggedGuard } from './logged.guard';

import { SessionService } from './session.service';

describe('LoggedGuard', () => {
  let guard: LoggedGuard;
  let sessionSpy: jasmine.SpyObj<SessionService>;
  let routerSpy : jasmine.SpyObj<Router>;

  beforeEach(() => {
    sessionSpy = jasmine.createSpyObj('SessionService', ['registered']);
    routerSpy  = jasmine.createSpyObj('Router', ['navigateByUrl']);

    TestBed.configureTestingModule({
      providers: [
        { provide: SessionService, useValue: sessionSpy },
        { provide: Router, useValue: routerSpy },
      ],
    });
    guard = TestBed.inject(LoggedGuard);
  });

  it('should be created', () => {
    expect(guard).toBeTruthy();
  });

  it('accepts when logged', () => {
    sessionSpy.registered.and.returnValue(true);
    expect(guard.canActivate({} as ActivatedRouteSnapshot, {} as RouterStateSnapshot)).toBe(true);
  });

  it('redirect to r/login when not logged', () => {
    sessionSpy.registered.and.returnValue(false);
    expect(guard.canActivate({} as ActivatedRouteSnapshot, {} as RouterStateSnapshot)).toBe(false);
    expect(routerSpy.navigateByUrl).toHaveBeenCalledWith('r/login');
  });

});
