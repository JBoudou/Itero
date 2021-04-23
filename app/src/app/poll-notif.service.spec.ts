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
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { RouterTestingModule } from '@angular/router/testing';

import { BehaviorSubject } from 'rxjs';

import { PollNotifService } from './poll-notif.service';
import { Recorder } from 'src/testing/recorder';
import { PollNotifAction, PollNotifAnswerEntry } from './api';
import { SessionInfo, SessionService, SessionState } from './session/session.service';
import { setSpyProperty } from 'src/testing/misc';

describe('PollNotifService', () => {
  let service: PollNotifService;
  let httpControler: HttpTestingController;
  let sessionState: BehaviorSubject<SessionInfo>;

  beforeEach(() => {
    jasmine.clock().install();

    const state$ = new BehaviorSubject<SessionInfo>(new SessionInfo(SessionState.Unlogged));
    const sessionSpy = jasmine.createSpyObj('SessionService', [], ['state$']);
    setSpyProperty(sessionSpy, 'state$', state$);

    TestBed.configureTestingModule({
      imports: [
        HttpClientTestingModule,
        RouterTestingModule, // Argh!
      ],
      providers: [
        { provide: SessionService, UseValue: sessionSpy },
      ],
    });
    service = TestBed.inject(PollNotifService);
    httpControler = TestBed.inject(HttpTestingController);
    sessionState = TestBed.inject(SessionService).state$ as BehaviorSubject<SessionInfo>;
  });

  afterEach(function() {
    jasmine.clock().uninstall();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('forward events from the middleware', () => {
    sessionState.next(new SessionInfo(SessionState.Logged));

    const recorder = new Recorder<PollNotifAnswerEntry>();
    recorder.listen(service.event$);
    jasmine.clock().tick(15021);

    const expected: PollNotifAnswerEntry[] = [
      {
        Timestamp: new Date('2012-12-12T12:12:12Z'),
        Segment: '123456789',
        Title: 'First',
        Round: 3,
        Action: PollNotifAction.Next,
      },
      {
        Timestamp: new Date('2020-02-20T02:20:02Z'),
        Segment: '987654321',
        Title: 'Second',
        Round: 7,
        Action: PollNotifAction.Term,
      },
    ];

    const req = httpControler.expectOne('/a/pollnotif');
    expect(req.request.method).toBe('POST');
    req.flush(JSON.stringify(expected));
    jasmine.clock().tick(1);

    expect(recorder.record.length).toBeGreaterThan(0);
    expect(recorder.record).toEqual(expected);
  });

});
