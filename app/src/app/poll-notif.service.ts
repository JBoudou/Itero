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

import { Observable, Subject, Subscription } from 'rxjs';
import { filter, take, timeout } from 'rxjs/operators';

import { PollNotifAnswerEntry } from './api';
import { SessionInfo, SessionService, SessionState } from './session/session.service';

export type PollNotification = PollNotifAnswerEntry;

@Injectable({
  providedIn: 'root'
})
export class PollNotifService {

  static interval = 15000;

  private _lastUpdate = new Date();
  private _intervalId: number|undefined;

  private _events = new Subject<PollNotification>();
  get event$(): Observable<PollNotification> {
    return this._events;
  }

  private _subscriptions: Subscription[] = [];

  constructor(
    private session: SessionService,
    private http: HttpClient,
  ) {
    this._subscriptions.push(
      this.session.state$.subscribe({
        next: (state: SessionInfo) => {
          switch (state.state) {
          case SessionState.Unlogged:
            this.stop();
            break;
          case SessionState.Logged:
            this.start();
            break;
          }
        },
      }),
    );
  }

  destroy() {
    this._subscriptions.forEach((sub: Subscription) => sub.unsubscribe());
    this.stop();
  }

  private start(): void {
    if (this._intervalId !== undefined) { return; }
    const interval = PollNotifService.interval + Math.floor(Math.random() * 20);
    this._intervalId = window.setInterval(() => this.pull(), interval);
  }

  private stop(): void {
    window.clearInterval(this._intervalId);
    this._lastUpdate = new Date();
    this._intervalId = undefined;
  }

  private pull() {
    this.session.state$.pipe(
      filter((state: SessionInfo) => state.state === SessionState.Logged),
      take(1),
      timeout(PollNotifService.interval * 0.99)
    ).subscribe({
      next: () => {
        this.http.post('/a/pollnotif', {LastUpdate: this._lastUpdate}, {responseType: 'text'})
          .pipe(take(1)).subscribe({
            next: (body: string) =>
              PollNotifAnswerEntry.fromJSONList(body)
                .forEach((evt: PollNotifAnswerEntry) => this._events.next(evt)),
            // TODO handle errors somehow
          });
        this._lastUpdate = new Date();
      }
    });
  }

}
