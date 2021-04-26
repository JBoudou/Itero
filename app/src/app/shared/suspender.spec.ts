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

import { Suspendable, Suspender } from './suspender';

describe('Suspender', () => {

  let suspender : Suspender;

  beforeEach(() => {
    jasmine.clock().install();
    jasmine.clock().mockDate();
  });

  afterEach(() => {
      suspender.cancel();
    jasmine.clock().uninstall();
  });

  describe('without action and delay', () => {

    beforeEach(() => {
      suspender = new Suspender();
    });

    it('executes the given action right now', () => {
      const action = jasmine.createSpy('action');
      expect(suspender.do(action)).toBeTrue();
      expect(action).toHaveBeenCalledTimes(1);
    })

    it('delays the given action', () => {
      const action = jasmine.createSpy('action');
      suspender.suspend(1000);
      expect(suspender.do(action)).toBeFalse();
      expect(action).not.toHaveBeenCalled();
      jasmine.clock().tick(1000);
      expect(action).toHaveBeenCalledTimes(1);
    })

    it('discard the second action', () => {
      const action1 = jasmine.createSpy('action1');
      const action2 = jasmine.createSpy('action2');
      suspender.suspend(1000);
      expect(suspender.do(action1)).toBeFalse();
      expect(suspender.do(action2)).toBeFalse();
      jasmine.clock().tick(2000);
      expect(action1).toHaveBeenCalledTimes(1);
      expect(action2).toHaveBeenCalledTimes(0);
    })

    it('does not execute canceled action', () => {
      const action = jasmine.createSpy('action');
      suspender.suspend(1000);
      expect(suspender.do(action)).toBeFalse();
      expect(action).toHaveBeenCalledTimes(0);
      suspender.cancel();
      jasmine.clock().tick(2000);
      expect(action).toHaveBeenCalledTimes(0);
    });

  }); // without action and delay

  describe('with delay only', () => {

    const delay = 1000;

    beforeEach(() => {
      suspender = new Suspender(delay);
    });

    it('executes the first action right now', () => {
      const action = jasmine.createSpy('action');
      expect(suspender.do(action)).toBeTrue();
      expect(action).toHaveBeenCalledTimes(1);
    })

    it('automatically delays the second action', () => {
      const action1 = jasmine.createSpy('action1');
      const action2 = jasmine.createSpy('action2');
      expect(suspender.do(action1)).toBeTrue();
      expect(suspender.do(action2)).toBeFalse();
      expect(action2).toHaveBeenCalledTimes(0);
      jasmine.clock().tick(delay);
      expect(action2).toHaveBeenCalledTimes(1);
    })

  }); // with delay only

  describe('with default action only', () => {

    let defaultAction: jasmine.Spy<() => void>;

    beforeEach(() => {
      defaultAction = jasmine.createSpy<() => void>('defaultAction');
      suspender = new Suspender(defaultAction);
    });

    it('calls default if none given', () => {
      expect(suspender.do()).toBeTrue();
      expect(defaultAction).toHaveBeenCalledTimes(1);
    });

    it('calls the given action', () => {
      const action = jasmine.createSpy<() => void>('action');
      expect(suspender.do(action)).toBeTrue();
      expect(defaultAction).toHaveBeenCalledTimes(0);
      expect(action).toHaveBeenCalledTimes(1);
    });

  }); // with default action only

  describe('with both action and delay', () => {

    const delay = 1000;
    let defaultAction: jasmine.Spy<() => void>;

    beforeEach(() => {
      defaultAction = jasmine.createSpy<() => void>('defaultAction');
      suspender = new Suspender(defaultAction, delay);
    });

    it('does not do anything is do() is not called', () => {
      jasmine.clock().tick(2 * delay);
      expect(defaultAction).toHaveBeenCalledTimes(0);
    });

    it('execute action once if do() is called once', () => {
      expect(suspender.do()).toBeTrue();
      jasmine.clock().tick(2 * delay);
      expect(defaultAction).toHaveBeenCalledTimes(1);
    });

    it('wait for the second call to do()', () => {
      expect(suspender.do()).toBeTrue();
      expect(suspender.do()).toBeFalse();
      expect(defaultAction).toHaveBeenCalledTimes(1);
      jasmine.clock().tick(delay);
      expect(defaultAction).toHaveBeenCalledTimes(2);
    });

    it('discard the third call to do()', () => {
      expect(suspender.do()).toBeTrue();
      expect(suspender.do()).toBeFalse();
      expect(suspender.do()).toBeFalse();
      jasmine.clock().tick(delay);
      jasmine.clock().tick(delay);
      jasmine.clock().tick(delay);
      expect(defaultAction).toHaveBeenCalledTimes(2);
    });

  });

});


class testSuspended {
  called: number = 0;

  action = Suspendable(() => {
    this.called += 1;
  }, 1000);
}

describe('Suspendable', () => {

  let obj: testSuspended;

  beforeEach(() => {
    obj = new testSuspended();
    jasmine.clock().install();
    jasmine.clock().mockDate();
  });

  afterEach(() => {
    jasmine.clock().uninstall();
  });

  it('executes the first call right now', () => {
    expect(obj.called).toBe(0);
    obj.action();
    expect(obj.called).toBe(1);
  });

  it('delays the second call', () => {
    obj.action();
    obj.action();
    expect(obj.called).toBe(1);
    jasmine.clock().tick(1000);
    expect(obj.called).toBe(2);
  });

  it('has suspend method', () => {
    obj.action.suspend(100);
    obj.action();
    expect(obj.called).toBe(0);
    jasmine.clock().tick(100);
    expect(obj.called).toBe(1);
  });

  it('does not share anything', () => {
    const other = new testSuspended();
    obj.action();
    expect(other.called).toBe(0);
    other.action();
    expect(obj.called).toBe(1);
    expect(other.called).toBe(1);
  });

});
