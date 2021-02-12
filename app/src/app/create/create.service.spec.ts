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
import { HttpErrorResponse } from '@angular/common/http';
import { Router } from '@angular/router';

import { CREATE_TREE, CreateService } from './create.service';

import { CreateQuery, PollAlternative } from '../api';
import { FinalNavTreeNode, LinearNavTreeNode, } from './navtree/navtree.node';
import { NavStepStatus } from './navtree/navstep.status';

import { Recorder, justRecordedFrom } from '../../testing/recorder';
import { RouterStub } from '../../testing/router.stub';

describe('CreateService', () => {

  const simpleAlternative: PollAlternative = { Id: 0, Name: 'test', Cost: 1 };

  let service: CreateService;
  let httpControler: HttpTestingController;
  let routerSpy: RouterStub;
  let stepRecorder: Recorder<NavStepStatus>;
  let queryRecorder: Recorder<Partial<CreateQuery>>;

  beforeEach(() => {
    const test_tree =
      new LinearNavTreeNode('root', 'Root',
        new LinearNavTreeNode('middle', 'Middle',
          new FinalNavTreeNode('leaf', 'Leaf')
        )
      );
    routerSpy = new RouterStub('/root');

    TestBed.configureTestingModule({
      imports: [ HttpClientTestingModule ],
      providers: [
        CreateService,
        { provide: CREATE_TREE, useValue: test_tree },
        { provide: Router, useValue: routerSpy },
      ],
    });
    service = TestBed.inject(CreateService);
    httpControler = TestBed.inject(HttpTestingController);

    jasmine.clock().install();
    stepRecorder  = new Recorder<NavStepStatus>();
    queryRecorder = new Recorder<Partial<CreateQuery>>();
  });

  afterEach(function() {
    stepRecorder .unsubscribe();
    queryRecorder.unsubscribe();
    jasmine.clock().uninstall();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('goes next', () => {
    stepRecorder.listen(service.stepStatus$);
    service.next();
    jasmine.clock().tick(1);

    expect(stepRecorder.record).toEqual([
        new NavStepStatus(0, ['Root', 'Middle', 'Leaf']),
        new NavStepStatus(1, ['Root', 'Middle', 'Leaf']),
    ]);

    expect(routerSpy.navigate.calls.count()).toBe(1);
    const segments = routerSpy.routerState.snapshot.url.split('/');
    expect(segments[segments.length - 1]).toBe('middle');
  });

  it('goes back', () => {
    service.next();
    jasmine.clock().tick(1);

    stepRecorder.listen(service.stepStatus$);
    service.back();
    jasmine.clock().tick(1);

    expect(stepRecorder.record).toEqual([
        new NavStepStatus(1, ['Root', 'Middle', 'Leaf']),
        new NavStepStatus(0, ['Root', 'Middle', 'Leaf']),
    ]);

    expect(routerSpy.navigate.calls.count()).toBe(2);
    const segments = routerSpy.routerState.snapshot.url.split('/');
    expect(segments[segments.length - 1]).toBe('root');
  });

  it('goes back twice', () => {
    service.next();
    service.next();
    jasmine.clock().tick(1);

    stepRecorder.listen(service.stepStatus$);
    service.back(2);
    jasmine.clock().tick(1);

    expect(stepRecorder.record).toEqual([
        new NavStepStatus(2, ['Root', 'Middle', 'Leaf']),
        new NavStepStatus(0, ['Root', 'Middle', 'Leaf']),
    ]);

    expect(routerSpy.navigate.calls.count()).toBe(3);
    const segments = routerSpy.routerState.snapshot.url.split('/');
    expect(segments[segments.length - 1]).toBe('root');
  });

  it('copies fields when going next', () => {
    // root
    service.patchQuery('root', { MaxNbRounds: 2, Alternatives: [] });

    // middle
    service.next();

    jasmine.clock().tick(1);
    const query = justRecordedFrom(service.query$)[0];
    expect(query.MaxNbRounds).toBe(2);
    expect(query.Alternatives).toEqual([]);
  });

  it('reset fields when going back', () => {
    // root
    service.patchQuery('root', { MaxNbRounds: 2, Alternatives: [] });

    // middle
    service.next();
    service.patchQuery('middle', { MaxNbRounds: 3, Alternatives: [simpleAlternative] });

    // back
    service.back();

    jasmine.clock().tick(1);
    const query = justRecordedFrom(service.query$)[0];
    expect(query.MaxNbRounds).toBe(2);
    expect(query.Alternatives).toEqual([]);
  });

  it('recalls fields when going next again', () => {
    // root
    service.patchQuery('root', { MaxNbRounds: 2 });

    // middle
    service.next();
    service.patchQuery('middle', { Alternatives: [simpleAlternative] });

    // next again
    service.back();
    service.next();

    jasmine.clock().tick(1);
    const query = justRecordedFrom(service.query$)[0];
    expect(query.MaxNbRounds).toBe(2);
    expect(query.Alternatives).toEqual([simpleAlternative]);
  });

  it('recalls fields after back-next-next', () => {
    // root
    service.patchQuery('root', { MaxNbRounds: 2 });

    // middle
    service.next();
    service.patchQuery('middle', { Alternatives: [simpleAlternative] });

    // next again
    service.back();
    service.next();
    service.next();

    jasmine.clock().tick(1);
    const query = justRecordedFrom(service.query$)[0];
    expect(query.MaxNbRounds).toBe(2);
    expect(query.Alternatives).toEqual([simpleAlternative]);
  });

  it('recalls fields after two back-next iterations', () => {
    // root
    service.patchQuery('root', { MaxNbRounds: 2 });

    // middle
    service.next();
    service.patchQuery('middle', { Alternatives: [simpleAlternative] });

    // next again
    service.back();
    service.next();
    service.back();
    service.next();

    jasmine.clock().tick(1);
    const query = justRecordedFrom(service.query$)[0];
    expect(query.MaxNbRounds).toBe(2);
    expect(query.Alternatives).toEqual([simpleAlternative]);
  });

  it('updates query$ on patchQuery', () => {
    queryRecorder.listen(service.query$);
    service.patchQuery('root', { MaxNbRounds: 2});
    
    expect(queryRecorder.record[0]?.MaxNbRounds).toBeUndefined();
    expect(queryRecorder.record[queryRecorder.record.length - 1].MaxNbRounds).toBe(2);
  });

  it('does not update when the wrong segment is given to patchQuery', () => {
    queryRecorder.listen(service.query$);
    service.patchQuery('middle', { MaxNbRounds: 2});
    
    expect(queryRecorder.record[queryRecorder.record.length - 1].MaxNbRounds).toBeUndefined();
  });

  it('reset everything on reset', () => {
    queryRecorder.listen(service.query$);
    service.patchQuery('root', { MinNbRounds: 2});
    service.next();
    service.patchQuery('middle', { MaxNbRounds: 27});
    service.next();
    service.reset();

    jasmine.clock().tick(1);
    const lastQuery = queryRecorder.record[queryRecorder.record.length - 1];
    expect(lastQuery?.MinNbRounds).toBeUndefined();
    expect(lastQuery?.MaxNbRounds).toBeUndefined();
  });

  it('sends a request on validation, then provides the query on getResult() and reset itself', () => {
    // Send a query
    service.patchQuery('root', { MinNbRounds: 2});
    service.next();
    service.patchQuery('middle', { Alternatives: [simpleAlternative] });
    service.next();
    service.patchQuery('leaf', { MaxNbRounds: 27});
    service.next();

    // Send a response
    const req = httpControler.expectOne('/a/create');
    expect(req.request.method).toBe('POST');
    req.flush('0segment0');
    httpControler.verify();

    // call getResult()
    const query = service.getResult() as Partial<CreateQuery>;
    jasmine.clock().tick(1);
    expect(query.MinNbRounds).toBe(2);
    expect(query.MaxNbRounds).toBe(27);
    expect(query.Alternatives).toEqual([simpleAlternative]);

    // Check reset
    const stored = justRecordedFrom(service.query$).pop();
    expect(stored?.MinNbRounds).toBeUndefined();
    expect(stored?.MaxNbRounds).toBeUndefined();
    expect(stored?.Alternatives).toBeUndefined();
    expect(service.currentUrl().split('/').pop()).toBe('root');
  });

  it('returns the HttpErrorResponse when calling getResult() after an error', () => {
    // Send a request
    service.next();
    service.next();
    service.next();

    // Send response
    const req = httpControler.expectOne('/a/create');
    expect(req.request.method).toBe('POST');
    req.flush('Argh', { status: 500, statusText: 'error' });
    httpControler.verify();

    // call getResult()
    const error = service.getResult() as HttpErrorResponse;
    expect(error.status).toBe(500);
    expect(error.error?.trim()).toBe('Argh');
  });

});
