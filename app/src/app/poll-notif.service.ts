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
import { HttpClient, HttpErrorResponse } from '@angular/common/http';

import { Observable, Subject } from 'rxjs';
import { take } from 'rxjs/operators';

import { PollNotifAnswerEntry } from './api';

@Injectable({
  providedIn: 'root'
})
export class PollNotifService {

  private _lastUpdate = new Date();
  private _intervalId: number;

  private _events = new Subject<PollNotifAnswerEntry>();
  get event$(): Observable<PollNotifAnswerEntry> {
    return this._events;
  }

  constructor(
    private http: HttpClient,
  ) {
    this._intervalId = window.setInterval(() => this.pull(), 15000);
  }

  destroy() {
    window.clearInterval(this._intervalId);
  }

  private pull() {
    this.http.post('/a/pollnotif', {LastUpdate: this._lastUpdate}, {responseType: 'text'})
      .pipe(take(1)).subscribe({
        next: (body: string) =>
          PollNotifAnswerEntry.fromJSONList(body)
            .forEach((evt: PollNotifAnswerEntry) => this._events.next(evt)),
        // TODO handle errors somehow
      });
    this._lastUpdate = new Date();
  }

}
