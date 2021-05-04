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

import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

import { BehaviorSubject, Observable } from 'rxjs';
import { take } from 'rxjs/operators';

import { ScheduleOne } from './shared/schedule-one';

@Injectable({
  providedIn: 'root'
})
export class ConfigService {

  private _demoPollSegment = new BehaviorSubject<string>('');
  get demoPollSegment$(): Observable<string> {
    return this._demoPollSegment;
  }

  constructor(
    private http: HttpClient,
  ) {
    this._retrieve();
  }

  private _retrieve(): void {
    this.http.get('/a/config').pipe(take(1)).subscribe({
      next: (obj: any) => {
        const hasDemoPollSegment = typeof obj['DemoPollSegment'] === 'string';
        this._demoPollSegment.next(hasDemoPollSegment ? obj.DemoPollSegment : '');
        this._reshedule();
      },
      error: () => {
        this._demoPollSegment.next('');
        this._reshedule();
      },
    });
  }

  private _scheduler = new ScheduleOne();
  private _reshedule(): void {
    this._scheduler.schedule(() => this._retrieve(), 3600*1000);
  }

}
