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

export class ScheduleOne {

  constructor() {
  }

  destroy(): void {
    this._cancel();
  }

  schedule(cb: () => void, timeout: number) {
    this._cancel();
    this._currentId = window.setTimeout(() => this._timeOut(cb), timeout);
  }


  // Implementation //

  private _currentId: number|undefined;

  private _cancel(): void {
    if (this._currentId !== undefined) {
      clearTimeout(this._currentId);
      this._currentId = undefined;
    }
  }

  private _timeOut(cb: () => void): void {
    this._currentId = undefined;
    cb();
  }

}
