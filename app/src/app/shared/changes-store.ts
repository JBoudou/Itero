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

import { SimpleChange, SimpleChanges } from '@angular/core';
import { cloneDeep } from 'lodash';

/** Interface to implement to receive changes notifications from ChangeStore.check(). */
export interface DynOnChanges {
  dynOnChanges(changes: SimpleChanges): void
}


/**
 * ChangeStore is a property map that record changes.
 * Changes are recorded between calls to check().
 *
 * Typical use is as a replacement for Angular's ngOnChange mechanism,
 * except that the component get notified about programmatical changes of input values.
 *
 * ```typescript
 * @Component()
 * class MyComponent implements DynOnChanges, DoCheck {
 *   private changeStore = new ChangeStore();
 *   @Input() @DynChanges('changeStore') value: string;
 *
 *   ngDoCheck(): void {
 *     // Check if there has been any change.
 *     this.changeStore.check(this);
 *   }
 *
 *   dynOnChanges(changes: SimpleChanges): void {
 *     // Update the component to reflect changes.
 *   }
 * }
 * ```
 */
export class ChangeStore {

  private changed: boolean = false
  private changes: SimpleChanges = {}

  set(key: string, val: any) {
    const change = this.changes[key]

    if (change === undefined) {
      this.changes[key] = new SimpleChange(undefined, val, true)
      this.changed = true
      return;
    }
    if (change.currentValue !== val) {
      change.currentValue = val
      this.changed = true
      return;
    }
  }

  get<T = any>(key: string): T|undefined {
    const change = this.changes[key]
    
    if (change === undefined) {
      return undefined
    }
    return change.currentValue as T
  }

  /**
   * Calls comp.dynOnChanges if and only if a monitored value has been changed since the last call
   * to check() (or the creation of the ChangeStore).
   * The excution of this method is guaranteed to be pretty fast, such that it can be called in
   * often executed methods like ngDoCheck().
   */
  check(comp: DynOnChanges): void {
    if (!this.changed) {
      return;
    }

    comp.dynOnChanges(cloneDeep(this.changes));
    this.changed = false;
    for (let key in this.changes) {
      const change = this.changes[key];
      if (change !== undefined) {
        change.previousValue = change.currentValue;
        change.firstChange = false;
      }
    }
  }

}


/**
 * Decorator thats mark a class field as being monitored by the ChangeStore field whose name is
 * given as parameter.
 * @see ChangeStore for typical usage.
 */
export function DynChanges(store: string|symbol): PropertyDecorator {
  return function(target: Object, key: string|symbol): void {
    if (typeof key === 'symbol') return;
    Object.defineProperty(target, key, {
      set: function(val: any)     { (this[store] as ChangeStore).set(key, val) },
      get: function(): any { return (this[store] as ChangeStore).get(key) },
    })
  }
}
