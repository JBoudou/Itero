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

import { CREATE_TREE, CreateService } from './create.service';

import { PollAlternative } from '../api';
import { FinalNavTreeNode, LinearNavTreeNode, } from './navtree/navtree.node';
import { NavStepStatus } from './navtree/navstep.status';

import { Recorder, justRecordedFrom } from '../../testing/recorder';
import { RouterStub } from '../../testing/router.stub';

describe('CreateService', () => {

  const simpleAlternative: PollAlternative = { Id: 0, Name: 'test', Cost: 1 };

  let service: CreateService;
  let httpControler: HttpTestingController;
  let routerSpy: RouterStub;
  let recorder: Recorder<NavStepStatus>;

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
    recorder = new Recorder<NavStepStatus>();
  });

  afterEach(function() {
    recorder.unsubscribe();
    jasmine.clock().uninstall();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('goes next', () => {
    recorder.listen(service.stepStatus$);
    service.next();
    jasmine.clock().tick(1);

    expect(recorder.record).toEqual([
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

    recorder.listen(service.stepStatus$);
    service.back();
    jasmine.clock().tick(1);

    expect(recorder.record).toEqual([
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

    recorder.listen(service.stepStatus$);
    service.back(2);
    jasmine.clock().tick(1);

    expect(recorder.record).toEqual([
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
    service.patchQuery('root', { MaxNbRounds: 2, Alternatives: [] });

    // middle
    service.next();
    service.patchQuery('middle', { MaxNbRounds: 3, Alternatives: [simpleAlternative] });

    // next again
    service.back();
    service.next();

    jasmine.clock().tick(1);
    const query = justRecordedFrom(service.query$)[0];
    expect(query.MaxNbRounds).toBe(3);
    expect(query.Alternatives).toEqual([simpleAlternative]);
  });

  /*

  it('resets fields after a successfull validation', () => {
    // root
    service.register(dummyComponent);
    service.next();

    // middle
    let query = service.register(simpleComponent);
    query.MaxNbRounds = 2;
    query.Alternatives = [];

    // leaf
    service.next();
    query = service.register(simpleComponent);
    query.MaxNbRounds = 3;
    query.Alternatives.push(simpleAlternative);

    // validate
    service.next(); // validate
    query = service.register(dummyComponent);
    expect(query.MaxNbRounds).toBeUndefined();
    expect(query.Alternatives).toBeUndefined();
    service.next();
    query = service.register(dummyComponent);
    expect(query.MaxNbRounds).toBeUndefined();
    expect(query.Alternatives).toBeUndefined();
  });

  */

});
