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
import { Router } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { Subject, of } from 'rxjs';

import { RouterStub } from 'src/testing/router.stub';
import { Recorder } from 'src/testing/recorder';

import { ListService } from './list.service';

import { ListAnswerEntry, PollAction } from '../api';
import { ServerError } from '../shared/server-error';
import { PollNotification, PollNotifService } from '../poll-notif.service';
import { setSpyProperty } from 'src/testing/misc';

describe('ListService', () => {
  let service: ListService;
  let httpControler: HttpTestingController;
  let routerStub: RouterStub;
  let dialogSpy: jasmine.SpyObj<MatDialog>;
  let pollNotifEvent: Subject<PollNotification>;

  const makeDialogSay = function(what: any) {
    const ref = jasmine.createSpyObj('MatDialogRef', {'afterClosed': of(what) });
    dialogSpy.open.and.returnValue(ref);
  }

  beforeEach(() => {
    routerStub = new RouterStub('');
    dialogSpy = jasmine.createSpyObj('MatDialog', ['open']);

    const pollNotifSpy = jasmine.createSpyObj('PollNotifService', [], ['event$']);
    setSpyProperty(pollNotifSpy, 'event$', new Subject<PollNotification>());

    TestBed.configureTestingModule({
      imports: [
        HttpClientTestingModule,
      ],
      providers: [
        { provide: Router, useValue: routerStub },
        { provide: MatDialog, useValue: dialogSpy },
        { provide: PollNotifService, useValue: pollNotifSpy },
      ],
    });
    jasmine.clock().install();

    service = TestBed.inject(ListService);
    httpControler = TestBed.inject(HttpTestingController);
    pollNotifEvent = TestBed.inject(PollNotifService).event$ as Subject<PollNotification>;
  });

  afterEach(function() {
    pollNotifEvent.complete();
    jasmine.clock().uninstall();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('when activated', () => {

    beforeEach(() => { service.activate() });
    afterEach(() => { service.desactivate() });

    it('fetches lists', () => {
      const publicList = [{
        Segment: '123456789',
        Title: 'Public',
        CurrentRound: 1,
        MaxRound: 3,
        Deadline: new Date('2012-12-12T12:12:12Z'),
        Action: PollAction.Vote,
        Deletable: false,
      }];
      const ownList = [{
        Segment: '987654321',
        Title: 'Own',
        CurrentRound: 2,
        MaxRound: 32,
        Deadline: new Date('2012-12-21T12:12:12Z'),
        Action: PollAction.Modi,
        Deletable: true,
      }];

      const recPublic = new Recorder<ListAnswerEntry[]>();
      const recOwn    = new Recorder<ListAnswerEntry[]>();
      recPublic.listen(service.publicList$);
      recOwn   .listen(service   .ownList$);

      const req = httpControler.expectOne('/a/list');
      expect(req.request.method).toEqual('GET');
      req.flush({
        Public: publicList,
        Own:    ownList,
      });
      jasmine.clock().tick(1);

      expect(recPublic.record[recPublic.record.length - 1]).toEqual(publicList);
      expect(recOwn   .record[recOwn   .record.length - 1]).toEqual(   ownList);

      recPublic.unsubscribe();
      recOwn   .unsubscribe();
    });

    it('converts JSON', () => {
      const recPublic = new Recorder<ListAnswerEntry[]>();
      recPublic.listen(service.publicList$);

      const req = httpControler.expectOne('/a/list');
      expect(req.request.method).toEqual('GET');
      req.flush('{"Public":[{' +
                '"Segment":"123456789","Title":"Public","CurrentRound":1,"MaxRound":3,' +
                '"Deadline":"⋅","Action":3}],' +
                '"Own":[]}');
      jasmine.clock().tick(1);

      expect(recPublic.record[recPublic.record.length - 1]).toEqual([{
        Segment: '123456789',
        Title: 'Public',
        CurrentRound: 1,
        MaxRound: 3,
        Deadline: undefined,
        Action: PollAction.Term,
        Deletable: false,
      }]);

      recPublic.unsubscribe();
    });

  }); // "when activated" nested describe

  it('sends delete request', () => {
    makeDialogSay(true);
    service.delete({
      Segment: '123456789',
      Title: 'Public',
      CurrentRound: 1,
      MaxRound: 3,
      Deadline: undefined,
      Action: PollAction.Term,
      Deletable: false,
    });
    const req = httpControler.expectOne('/a/delete/123456789');
    expect(req.request.method).toBe('GET');
  });

  it('transmit http errors', () => {
    const rec = new Recorder<ServerError>();
    rec.listen(service.error$);

    makeDialogSay(true);
    service.delete({
      Segment: '123456789',
      Title: 'Public',
      CurrentRound: 1,
      MaxRound: 3,
      Deadline: undefined,
      Action: PollAction.Term,
      Deletable: false,
    });

    const req = httpControler.expectOne('/a/delete/123456789');
    expect(req.request.method).toBe('GET');
    req.flush('Some error', {status: 404, statusText: 'Oups'});
    jasmine.clock().tick(1);

    expect(rec.record.length).toBeGreaterThan(0);
    const lastError = rec.record[rec.record.length - 1];
    expect(lastError.status).toBe(404);
    expect(lastError.message).toBe('Some error');

    rec.unsubscribe();
  });
  
});
