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

import {matcher} from 'd3';
import { ChangeStore, DynOnChanges } from './changes-store';

interface onChangeCallSpec {
  [key: string]: {currentValue: any, firstChange: true} | {currentValue: any, previousValue: any}
}

function changesLike(spec: onChangeCallSpec): jasmine.AsymmetricMatcher<any> {
  return {
    asymmetricMatch(other: any): boolean {
      for (let key in spec) {
        if (!(key in other)) {
          return false
        }
        const got = other[key]
        const exp = spec[key]
        if (got.currentValue !== exp.currentValue) {
          return false
        }
        if ('firstChange' in exp) {
          if (got.firstChange !== exp.firstChange) {
            return false
          }
        } else {
          if (got.previousValue !== exp.previousValue) {
            return false
          }
        }
      }
      return true
    },
    jasmineToString(): string {
      return jasmine.matchersUtil.pp(spec)
    }
  }
}

describe('ChangeStore', () => {
  let store: ChangeStore;
  let dynSpy: jasmine.SpyObj<DynOnChanges>;

  beforeEach(() => {
    store = new ChangeStore()
    dynSpy = jasmine.createSpyObj('ChangeStore', ['dynOnChanges'])
  })

  it('get undefined values', () => {
    expect(store.get('foo')).toBe(undefined)
  })

  it('get set value', () => {
    store.set('foo', 'bar')
    expect(store.get<string>('foo')).toBe('bar')
  })

  it('get twice set value', () => {
    store.set('foo', 'blu')
    store.set('foo', 'bar')
    expect(store.get<string>('foo')).toBe('bar')
  })

  it('check does nothing until set', () => {
    store.check(dynSpy)
    expect(dynSpy.dynOnChanges).not.toHaveBeenCalled()
  })

  it('check active after initial set', () => {
    store.set('foo', 'bar')
    store.check(dynSpy)
  })

  it('does nothing on second check', () => {
    store.set('foo', 'bar')
    store.check(dynSpy)
    store.check(dynSpy)
    expect(dynSpy.dynOnChanges).toHaveBeenCalledTimes(1);
  })

  it('check after two set', () => {
    store.set('foo', 'bar')
    store.check(dynSpy)
    store.set('foo', 'blu')
    store.check(dynSpy)
    expect(dynSpy.dynOnChanges).toHaveBeenCalledWith(changesLike({
      foo: {currentValue: 'blu', previousValue: 'bar'}
    }));
  })

  it('checks correctly after two set', () => {
    store.set('foo', 'bar')
    store.set('foo', 'blu')
    store.check(dynSpy)
    expect(dynSpy.dynOnChanges).toHaveBeenCalledWith(changesLike({
      foo: {currentValue: 'blu', firstChange: true}
    }));
  });

  it('checks correctly after two set after a check', () => {
    store.set('foo', 'one')
    store.check(dynSpy)
    expect(dynSpy.dynOnChanges).toHaveBeenCalled()
    store.set('foo', 'two')
    store.set('foo', 'three')
    store.check(dynSpy)
    expect(dynSpy.dynOnChanges).toHaveBeenCalledWith(changesLike({
      foo: {currentValue: 'three', previousValue: 'one'}
    }));
  });

})
