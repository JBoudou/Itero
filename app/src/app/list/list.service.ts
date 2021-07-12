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

import { Component, Inject, Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Router } from '@angular/router';

import { BehaviorSubject, Observable, Subject, Subscription } from 'rxjs';
import { take } from 'rxjs/operators';
import { MatDialog, MAT_DIALOG_DATA } from '@angular/material/dialog';

import { ListAnswer, ListAnswerEntry } from '../api';
import { ServerError } from '../shared/server-error';
import { PollNotifService } from '../poll-notif.service';
import { Suspendable } from '../shared/suspender';

@Injectable({
  providedIn: 'root'
})
export class ListService {

  private _public = new BehaviorSubject<ListAnswerEntry[]>([]);
  private _own    = new BehaviorSubject<ListAnswerEntry[]>([]);

  get publicList$(): Observable<ListAnswerEntry[]> {
    return this._public;
  }
  get ownList$(): Observable<ListAnswerEntry[]> {
    return this._own;
  }

  private _error = new Subject<ServerError>();

  get error$(): Observable<ServerError> {
    return this._error;
  }

  go(poll: ListAnswerEntry): void {
    this.router.navigateByUrl('/r/poll/' + poll.Segment);
  }

  delete(poll: ListAnswerEntry): void {
    const ref = this.dialog.open<DeletePollDialog, ListAnswerEntry, boolean>(DeletePollDialog, { data: poll });
    ref.afterClosed().pipe(take(1)).subscribe({
      next: (result: boolean) => {
        if (result) {
          this.http.get('/a/delete/' + poll.Segment).pipe(take(1)).subscribe({
            next: () => this._refresh(),
            error: (err: HttpErrorResponse) => {
              this._error.next(new ServerError(err, 'deleting poll ' + poll.Segment));
              this._refresh();
            },
          });
        }
      }
    });
  }

  launch(poll: ListAnswerEntry): void {
    this.http.get('/a/launch/' + poll.Segment).pipe(take(1)).subscribe({
      next: () => this.go(poll),
      error: (err: HttpErrorResponse) => {
        this._error.next(new ServerError(err, 'launching poll ' + poll.Segment));
        this._refresh();
      },
    });
  }

  constructor(
    private http: HttpClient,
    private router: Router,
    private dialog: MatDialog,
    private pollNotif: PollNotifService,
  ) { }

  private _activation: number = 0;
  private _pollNotifSubscription: Subscription;

  /**
   * Inform ListService that some list is displayed.
   * A call to desactivate() is expected when the lists stop being displayed.
   */
  activate(): void {
    if (this._activation === 0) {
      this._error.next(new ServerError());
      this._refresh();
      this._pollNotifSubscription = this.pollNotif.event$.subscribe({
        next: () => this._refresh(),
      });
    }
    this._activation += 1;
  }

  /**
   * Inform ListService that some list are not displayed anymore.
   * Each call to desactivate() must correspond to one call to activate().
   */
  desactivate(): void {
    this._activation -= 1;
    if (this._activation === 0) {
      this._pollNotifSubscription.unsubscribe();
    }
  }

  private readonly _refresh = Suspendable(function(): void {
    this.http.get('/a/list', {responseType: 'text'}).pipe(take(1)).subscribe({
      next: (answerText: string) => {
        const answer = ListAnswer.fromJSON(answerText);
        this._public.next(answer.Public);
        this._own   .next(answer.Own   );
      },
      error: (err: HttpErrorResponse) =>
        this._error.next(new ServerError(err, 'fetching the list of polls')),
    });
  }, 2000);
}

@Component({
  selector: 'delete-poll-dialog',
  templateUrl: 'delete-poll.dialog.html',
  host: {class: 'dialog-box'},
})
export class DeletePollDialog {
  constructor(
    @Inject(MAT_DIALOG_DATA) public poll: ListAnswerEntry,
  ) { }
}
