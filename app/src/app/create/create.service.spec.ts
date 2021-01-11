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
import { Router } from '@angular/router';

import { of } from 'rxjs';

import {
  CREATE_TREE,
  CreateNextStatus,
  CreateService,
  CreateStepStatus,
  FinalCreateTreeNode,
  LinearCreateTreeNode,
} from './create.service';

import { PollAlternative } from '../api';

import { Recorder, justRecordedFrom } from '../../testing/recorder';

describe('CreateTree', () => {
  it('may consist of a final node', () => {
    const final = new FinalCreateTreeNode('segment', 'title');
    
    expect(final.parent).toBeUndefined();
    expect(final.next()).toBeUndefined();
    expect(final.isFinal).toBeTrue();

    const status = final.makeStatus();
    expect(status.current).toBe(0);
    expect(status.steps).toEqual(['title']);
    expect(status.mayHaveMore).toBeFalse();
  });

  it('may contain a linear node', () => {
    const leaf = new FinalCreateTreeNode('final', 'Leaf');
    const root = new LinearCreateTreeNode('base', 'Root', leaf);

    expect(leaf.parent).toBe(root);
    expect(leaf.next()).toBeUndefined();
    expect(leaf.isFinal).toBeTrue();
    expect(leaf.makeStatus()).toEqual(new CreateStepStatus(1, ['Root', 'Leaf'], false));

    expect(root.parent).toBeUndefined();
    expect(root.next()).toBe(leaf);
    expect(root.isFinal).toBeFalse();
    expect(root.makeStatus()).toEqual(new CreateStepStatus(0, ['Root', 'Leaf'], false));
  });
});

class RouterStub {
  private url = '/root';
  navigate = jasmine.createSpy('navigate');

  get routerState() {
    return { snapshot: { url: this.url } };
  }

  constructor() {
    this.navigate.and.callFake((url: string[]) => {
      this.url = url.join('/');
      // Fast promise because we don't want to wait one cycle.
      return { then: (resolve: (v: boolean) => void) => resolve(true) };
    });
  }
}

describe('CreateService', () => {
  const TEST_TREE =
    new LinearCreateTreeNode('root', 'Root',
      new FinalCreateTreeNode('leaf', 'Leaf')
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
  let routerSpy: RouterStub;

  beforeEach(() => {
    routerSpy = new RouterStub();

    TestBed.configureTestingModule({
      providers: [
        { provide: CREATE_TREE, useValue: TEST_TREE },
        { provide: Router, useValue: routerSpy },
      ],
    });
    service = TestBed.inject(CreateService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('registers subComponent', () => {
    const recorder = new Recorder<CreateNextStatus>();
    service.createNextStatus$.subscribe(recorder);

    service.register(dummyComponent);

    expect(recorder.record).toEqual([
        new CreateNextStatus(false, false),
        new CreateNextStatus(true, false),
    ]);
    expect(justRecordedFrom(service.createStepStatus$)).toEqual([ new CreateStepStatus(0, ['Root', 'Leaf']) ]);
  });

  it('goes next', () => {
    service.register(dummyComponent);

    const recorder = new Recorder<CreateStepStatus>();
    service.createStepStatus$.subscribe(recorder);
    service.next();
    service.register(dummyComponent);
    expect(recorder.record).toEqual([
        new CreateStepStatus(0, ['Root', 'Leaf']),
        new CreateStepStatus(1, ['Root', 'Leaf']),
    ]);

    expect(routerSpy.navigate.calls.count()).toBe(1);
    const segments = routerSpy.routerState.snapshot.url.split('/');
    expect(segments[segments.length - 1]).toBe('leaf');
  });

  it('goes back', () => {
    service.register(dummyComponent);
    service.next();
    service.register(dummyComponent);

    const recorder = new Recorder<CreateStepStatus>();
    service.createStepStatus$.subscribe(recorder);
    service.back();
    service.register(dummyComponent);
    expect(recorder.record).toEqual([
        new CreateStepStatus(1, ['Root', 'Leaf']),
        new CreateStepStatus(0, ['Root', 'Leaf']),
    ]);

    expect(routerSpy.navigate.calls.count()).toBe(2);
    const segments = routerSpy.routerState.snapshot.url.split('/');
    expect(segments[segments.length - 1]).toBe('root');
  });

  it('copies fields when going next', () => {
    // root
    let query = service.register(simpleComponent);
    query.MaxNbRounds = 2;
    query.Alternatives = [];

    // leaf
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

    // leaf
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

  it('recalls fields when goinf next again', () => {
    // root
    let query = service.register(simpleComponent);
    query.MaxNbRounds = 2;
    query.Alternatives = [];

    // leaf
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

});
