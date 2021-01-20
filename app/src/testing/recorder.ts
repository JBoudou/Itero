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

import { Observable, Observer } from 'rxjs';

export function justRecordedFrom<T>(obs: Observable<T>): T[] {
  const recorder = new Recorder<T>();
  obs.subscribe(recorder);
  return recorder.record;
}

export class Recorder<T> implements Observer<T> {
  private _record: T[];
  private _errors: any[];

  constructor() {
    this._record = [];
    this._errors = [];
    this.closed = false;
  }

  get record(): T[] {
    return this._record;
  }

  get errors(): any[] {
    return this._errors;
  }

  /* Implements Observer */

  public closed: boolean;

  next(value: T): void {
    this._record.push(value);
  }

  error(err: any): void {
    this._errors.push(err);
  }

  complete(): void {
    this.closed = true;
  }

}
