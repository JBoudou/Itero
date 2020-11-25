import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';

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
    expect(service.sessionId).toBe('');
  })

  it('adds session info when there is no query', () => {
    service.sessionId = 'ABCD';
    let url = service.makeURL('/foo');
    let params = new URL(url, 'http://localhost/').searchParams;
    expect(params.has('s')).toBeTrue();
    expect(params.get('s')).toBe(service.sessionId);
  });

  it('adds session info when there is a query', () => {
    service.sessionId = 'ABCD';
    let url = service.makeURL('/foo?bar=27');
    let params = new URL(url, 'http://localhost/').searchParams;
    expect(params.has('s')).toBeTrue();
    expect(params.get('s')).toBe(service.sessionId);
  });

  it('does not create a session on failed login', done => {
    service.login({User: 'foo', Passwd: 'bar'}).subscribe({
      error: () => {
        expect(service.sessionId).toBe('');
        done();
      }
    });
    const req = httpTestingController.expectOne('/a/login');
    expect(req.request.method).toEqual('POST');
    req.flush('Argh', { status: 403, statusText: 'Unauthorized' });
  });

  it('creates a session on successful login', done => {
    service.observable.subscribe((notif: SessionInfo) => {
      expect(notif).toEqual({registered: true, user: 'foo'});
      expect(service.sessionId).toBe('ABCD');
      done();
    });

    service.login({User: 'foo', Passwd: 'bar'}).subscribe();
    const req = httpTestingController.expectOne('/a/login');
    expect(req.request.method).toEqual('POST');
    req.flush('ABCD');
  });

  it('removes the session after logoff', done => {
    let count = 0;
    service.observable.subscribe((notif: SessionInfo) => {
      if (count == 0) {
        count = 1;
        return;
      }
      expect(notif).toEqual({registered: false, user: ''});
      expect(service.sessionId).toBe('');
      done();
    });

    service.login({User: 'foo', Passwd: 'bar'}).subscribe();
    const req = httpTestingController.expectOne('/a/login');
    expect(req.request.method).toEqual('POST');
    req.flush('ABCD');

    service.logoff();
  });

});
