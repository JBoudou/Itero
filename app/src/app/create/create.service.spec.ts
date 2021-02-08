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

import { of } from 'rxjs';

import { CREATE_TREE, CreateService } from './create.service';

import { PollAlternative } from '../api';
import { FinalNavTreeNode, LinearNavTreeNode, } from './navtree/navtree.node';
import { NavStepStatus } from './navtree/navstep.status';

import { Recorder, justRecordedFrom } from '../../testing/recorder';
import { RouterStub } from '../../testing/router.stub';

describe('CreateService', () => {
  const TEST_TREE =
    new LinearNavTreeNode('root', 'Root',
      new LinearNavTreeNode('middle', 'Middle',
        new FinalNavTreeNode('leaf', 'Leaf')
      )
    );

  const dummyComponent = jasmine.createSpyObj('CreateSubComponent', [], {
    handledFields: new Set<string>([]),
    validable$: of(true),
  });

  const simpleComponent = jasmine.createSpyObj('CreateSubComponent', [], {
    handledFields: new Set<string>(['MaxNbRounds', 'Alternatives']),
    validable$: of(true),
  });

  const simpleAlternative: PollAlternative = { Id: 0, Name: 'test', Cost: 1 };

  let service: CreateService;
  let httpControler: HttpTestingController;
  let routerSpy: RouterStub;

  beforeEach(() => {
    routerSpy = new RouterStub('/root');

    TestBed.configureTestingModule({
      imports: [ HttpClientTestingModule ],
      providers: [
        CreateService,
        { provide: CREATE_TREE, useValue: TEST_TREE },
        { provide: Router, useValue: routerSpy },
      ],
    });
    service = TestBed.inject(CreateService);
    httpControler = TestBed.inject(HttpTestingController);

    jasmine.clock().install();
  });

  afterEach(function() {
    jasmine.clock().uninstall();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  /*

  it('registers subComponent', () => {
    const recorder = new Recorder<CreateNextStatus>();
    service.createNextStatus$.subscribe(recorder);

    service.register(dummyComponent);
    jasmine.clock().tick(1);

    expect(recorder.record).toEqual([
        new CreateNextStatus(false, false),
        new CreateNextStatus(true, false),
    ]);
    expect(justRecordedFrom(service.createStepStatus$)).toEqual([ new NavStepStatus(0, ['Root', 'Middle', 'Leaf']) ]);
  });

  it('goes next', () => {
    service.register(dummyComponent);
    jasmine.clock().tick(1);

    const recorder = new Recorder<NavStepStatus>();
    service.createStepStatus$.subscribe(recorder);
    service.next();
    service.register(dummyComponent);
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
    service.register(dummyComponent);
    service.next();
    service.register(dummyComponent);
    jasmine.clock().tick(1);

    const recorder = new Recorder<NavStepStatus>();
    service.createStepStatus$.subscribe(recorder);
    service.back();
    service.register(dummyComponent);
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
    service.register(dummyComponent);
    service.next();
    service.register(dummyComponent);
    service.next();
    service.register(dummyComponent);
    jasmine.clock().tick(1);

    const recorder = new Recorder<NavStepStatus>();
    service.createStepStatus$.subscribe(recorder);
    service.back(2);
    service.register(dummyComponent);
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
    let query = service.register(simpleComponent);
    query.MaxNbRounds = 2;
    query.Alternatives = [];

    // middle
    service.next();
    query = service.register(simpleComponent);
    expect(query.MaxNbRounds).toBe(2);
    expect(query.Alternatives).toEqual([]);
  });

  it('reset fields when going back', () => {
    // root
    let query = service.register(simpleComponent);
    query.MaxNbRounds = 2;
    query.Alternatives = [];

    // middle
    service.next();
    query = service.register(simpleComponent);
    query.MaxNbRounds = 3;
    query.Alternatives.push(simpleAlternative);

    // back
    service.back();
    query = service.register(simpleComponent);
    expect(query.MaxNbRounds).toBe(2);
    expect(query.Alternatives).toEqual([]);
  });

  it('recalls fields when going next again', () => {
    // root
    let query = service.register(simpleComponent);
    query.MaxNbRounds = 2;
    query.Alternatives = [];

    // middle
    service.next();
    query = service.register(simpleComponent);
    query.MaxNbRounds = 3;
    query.Alternatives.push(simpleAlternative);

    // next again
    service.back();
    // We do NOT register a component here. Therefore there must be NO handledFields for root now.
    service.next();
    query = service.register(simpleComponent);
    expect(query.MaxNbRounds).toBe(3);
    expect(query.Alternatives).toEqual([simpleAlternative]);
  });

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
