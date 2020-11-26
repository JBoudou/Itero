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

import { TestBed } from '@angular/core/testing';
import { HttpErrorResponse } from '@angular/common/http';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';

import { throwError, of } from 'rxjs';

import { SessionInfo, SessionService } from './session.service';

describe('SessionService', () => {
  let service: SessionService;
  let httpTestingController: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [ HttpClientTestingModule ],
      providers: [SessionService]
    });
    service = TestBed.inject(SessionService);
    httpTestingController = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    // After every test, assert that there are no more pending requests.
    httpTestingController.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('does not have a session at startup', () => {
    expect(service.registered()).toBeFalse();
  })

  it('does not create a session on failed login', done => {
    throwError(new HttpErrorResponse({status: 403}))
      .pipe(service.httpOperator('foo'))
      .subscribe({
      error: () => {
        expect(service.registered()).toBeFalse();
        done();
      }
    });
  });

  it('creates a session on successful login', done => {
    service.observable.subscribe((notif: SessionInfo) => {
      expect(notif).toEqual({registered: true, user: 'foo'});
      expect(service.registered()).toBeTrue();
      expect(service.sessionId).toBe('ABCD');
      done();
    });

    of('ABCD')
      .pipe(service.httpOperator('foo'))
      .subscribe();
  });

  it('removes the session after logoff', done => {
    let count = 0;
    service.observable.subscribe((notif: SessionInfo) => {
      if (count == 0) {
        count = 1;
        return;
      }
      expect(notif.registered).toBeFalse();
      done();
    });

    of('ABCD')
      .pipe(service.httpOperator('foo'))
      .subscribe();
    service.logoff();
  });

});
