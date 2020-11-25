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
    const serviceSpy = jasmine.createSpyObj('SessionService', ['registered'], {sessionId: 'ABCD'});

    TestBed.configureTestingModule({
    providers: [
      SessionInterceptor,
      { provide: SessionService, useValue: serviceSpy}
      ]
    });
  })

  function checkUrlHandler(url: string): HttpHandler {
    return {
      handle: (req: HttpRequest<any>): Observable<HttpEvent<any>> => {
        expect(req.url).toBe(url);
        return EMPTY;
      }
    };
  }

  it('should be created', () => {
    const interceptor: SessionInterceptor = TestBed.inject(SessionInterceptor);
    expect(interceptor).toBeTruthy();
  });

  it('adds session when needed', () => {
    const interceptor: SessionInterceptor = TestBed.inject(SessionInterceptor);
    
    let service = TestBed.inject(SessionService) as jasmine.SpyObj<SessionService>;
    service.registered.and.returnValue(true);

    interceptor.intercept(new HttpRequest('GET', '/foo'), checkUrlHandler('/foo?s=ABCD'));
    interceptor.intercept(new HttpRequest('POST', '/foo?t=bar', {}), checkUrlHandler('/foo?t=bar&s=ABCD'));
  });

  it('does not change URL when there is no session', () => {
    const interceptor: SessionInterceptor = TestBed.inject(SessionInterceptor);
    
    let service = TestBed.inject(SessionService) as jasmine.SpyObj<SessionService>;
    service.registered.and.returnValue(false);

    interceptor.intercept(new HttpRequest('GET', '/foo'), checkUrlHandler('/foo'));
    interceptor.intercept(new HttpRequest('POST', '/foo?t=bar', {}), checkUrlHandler('/foo?t=bar'));
  });
});
