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

import {cloneDeep} from 'lodash';
import { Observable, Observer, BehaviorSubject, Subscription } from 'rxjs';
import { filter, mapTo, take } from 'rxjs/operators';

/** Returns the events sent by the Observable on subscribe(). */
export function justRecordedFrom<T>(obs: Observable<T>): T[] {
  const recorder = new Recorder<T>();
  obs.subscribe(recorder).unsubscribe();
  return recorder.record;
}

/** A simple observer saving events and errors in arrays. */
export class Recorder<T> implements Observer<T> {
  private _record: T[];
  private _errors: any[];
  private _recordLength: BehaviorSubject<number>;
  private _subscription: Subscription;

  constructor() {
    this._record = [];
    this._errors = [];
    this.closed = false;
    this._recordLength = new BehaviorSubject<number>(0);
  }

  get record(): T[] {
    return this._record;
  }

  get errors(): any[] {
    return this._errors;
  }

  listen(obs: Observable<T>): Recorder<T> {
    this.closed = false;
    this._subscription = obs.subscribe(this);
    return this;
  }

  unsubscribe(): boolean {
    if (this._subscription !== undefined && !this.closed) {
      this._subscription.unsubscribe();
      this._subscription = undefined;
      this.closed = true;
      return true;
    }
    return false;
  }

  get recordLength$(): Observable<number> {
    return this._recordLength;
  }

  waitLength(minLength: number, timeOut?: number): Promise<boolean> {
    const length =
      this.recordLength$.pipe(filter((len: number) => len >= minLength), mapTo(true), take(1)).toPromise();
    const wait =
      new Promise((resolve: (value?: boolean | PromiseLike<boolean>) => void) => setTimeout(() => resolve(false), timeOut));
    return Promise.race([length, wait]);
  }

  /* Implements Observer */

  public closed: boolean;

  next(value: T): void {
    this._record.push(cloneDeep(value));
    this._recordLength.next(this._record.length);
  }

  error(err: any): void {
    this._errors.push(err);
  }

  complete(): void {
    this.closed = true;
  }

}
