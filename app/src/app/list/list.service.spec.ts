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
import { HttpErrorResponse } from '@angular/common/http';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { Router } from '@angular/router';

import { RouterStub } from '../../testing/router.stub';
import { Recorder } from '../../testing/recorder';

import { ListService } from './list.service';

import { ListAnswerEntry, PollAction } from '../api';

describe('ListService', () => {
  let service: ListService;
  let httpControler: HttpTestingController;
  let routerStub: RouterStub;

  beforeEach(() => {
    routerStub = new RouterStub('');

    TestBed.configureTestingModule({
      imports: [
        HttpClientTestingModule,
      ],
      providers: [
        { provide: Router, useValue: routerStub },
      ],
    });

    service = TestBed.inject(ListService);
    httpControler = TestBed.inject(HttpTestingController);
    jasmine.clock().install();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  afterEach(function() {
    jasmine.clock().uninstall();
  });

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

    service.refresh();

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

    service.refresh();

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

  it('sends delete request', () => {
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
    const rec = new Recorder<HttpErrorResponse>();
    rec.listen(service.error$);

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
    // TODO: Check the message. It seems that TestRequest does not allow body with error response.
    //expect(lastError.message.trim()).toBe('Some error');

    rec.unsubscribe();
  });
  
});
