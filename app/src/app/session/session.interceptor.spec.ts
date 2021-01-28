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

import { HttpRequest, HttpEvent, HttpHandler } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';

import { Observable, EMPTY } from 'rxjs';

import { SessionInterceptor } from './session.interceptor';
import { SessionService } from './session.service';

describe('SessionInterceptor', () => {
  beforeEach(() => {
    const serviceSpy = {
      sessionId: 'ABCD',
      logged: false,
    };

    TestBed.configureTestingModule({
    providers: [
      SessionInterceptor,
      { provide: SessionService, useValue: serviceSpy}
      ]
    });
  })

  function checkHasSession(sessionId: string): HttpHandler {
    return {
      handle: (req: HttpRequest<any>): Observable<HttpEvent<any>> => {
        expect(req.headers.get('X-CSRF')).toEqual(sessionId);
        return EMPTY;
      }
    };
  }

  let checkNoSession: HttpHandler = {
    handle: (req: HttpRequest<any>): Observable<HttpEvent<any>> => {
      expect(req.headers.has('X-CSRF')).toBeFalse();
      return EMPTY;
    }
  };

  it('should be created', () => {
    const interceptor: SessionInterceptor = TestBed.inject(SessionInterceptor);
    expect(interceptor).toBeTruthy();
  });

  it('adds session when needed', () => {
    const interceptor: SessionInterceptor = TestBed.inject(SessionInterceptor);
    
    let service = TestBed.inject(SessionService) as any;
    service.logged = true;

    interceptor.intercept(new HttpRequest('GET', '/foo'), checkHasSession('ABCD'));
    interceptor.intercept(new HttpRequest('POST', '/foo?t=bar', {}), checkHasSession('ABCD'));
  });

  it('does not add session when there is none', () => {
    const interceptor: SessionInterceptor = TestBed.inject(SessionInterceptor);
    
    let service = TestBed.inject(SessionService) as any;
    service.logged = false;

    interceptor.intercept(new HttpRequest('GET', '/foo'), checkNoSession);
    interceptor.intercept(new HttpRequest('POST', '/foo?t=bar', {}), checkNoSession);
  });
});
