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
import { Router } from '@angular/router';

import { throwError, of } from 'rxjs';
import { MatDialog } from '@angular/material/dialog';

import { SessionInfo, SessionService } from './session.service';

import { Recorder } from 'src/testing/recorder';
import { RouterStub } from 'src/testing/router.stub';

describe('SessionService', () => {
  let service: SessionService;
  let httpTestingController: HttpTestingController;
  let routerSpy: RouterStub;

  beforeEach(() => {
    routerSpy = new RouterStub();

    const dummyMatDialog = jasmine.createSpyObj('MatDialog', ['open'])

    TestBed.configureTestingModule({
      imports: [ HttpClientTestingModule ],
      providers: [
        SessionService,
        { provide: Router, useValue: routerSpy },
        { provide: MatDialog, useValue: dummyMatDialog },
      ],
    });
    service = TestBed.inject(SessionService);
    httpTestingController = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    // After every test, assert that there are no more pending requests.
    httpTestingController.verify();

    // Remove any session
    if (!!service) {
      service.logoff();
    }
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('does not have a session at startup', () => {
    expect(service.logged).toBeFalse();
  })

  it('does not create a session on failed login', done => {
    throwError(new HttpErrorResponse({status: 403}))
      .pipe(service.httpOperator('foo'))
      .subscribe({
      error: () => {
        expect(service.logged).toBeFalse();
        done();
      }
    });
  });

  it('creates a session on successful login', () => {
    const recorder = new Recorder<SessionInfo>();
    service.state$.subscribe(recorder);

    of({SessionId: 'ABCD', Expires: new Date(Date.now() + 600 * 1000)})
      .pipe(service.httpOperator('foo'))
      .subscribe();

    const last = recorder.record.length - 1;
    expect(last).toBeGreaterThanOrEqual(0);
    expect(recorder.record[last].logged).toBeTrue();
    expect(recorder.record[last].user).toBe('foo');
    expect(service.logged).toBeTrue();
    expect(service.sessionId).toBe('ABCD');
  });

  it('removes the session after logoff', () => {
    const recorder = new Recorder<SessionInfo>();
    service.state$.subscribe(recorder);

    of({SessionId: 'ABCD', Expires: new Date(Date.now() + 600 * 1000)})
      .pipe(service.httpOperator('foo'))
      .subscribe();
    service.logoff();

    const last = recorder.record.length - 1;
    expect(last).toBeGreaterThanOrEqual(0);
    expect(recorder.record[last].logged).toBeFalse();
    expect(recorder.record[last - 1].logged).toBeTrue();
    expect(recorder.record[last - 1].user).toBe('foo');
  });

});
