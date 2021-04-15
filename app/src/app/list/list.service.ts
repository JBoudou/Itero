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
import { Router } from '@angular/router';

import { BehaviorSubject, Observable, Subject } from 'rxjs';
import { take } from 'rxjs/operators';

import { ListAnswer, ListAnswerEntry } from '../api';

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

  private _error = new Subject<HttpErrorResponse>();

  get error$(): Observable<HttpErrorResponse> {
    return this._error;
  }

  refresh(): void {
    this.http.get('/a/list', {responseType: 'text'}).pipe(take(1)).subscribe({
      next: (answerText: string) => {
        const answer = ListAnswer.fromJSON(answerText);
        this._public.next(answer.Public);
        this._own   .next(answer.Own   );
      },
      error: (err: HttpErrorResponse) => this.handleError(err),
    });
  }

  go(poll: ListAnswerEntry): void {
    this.router.navigateByUrl('/r/poll/' + poll.Segment)
  }

  delete(poll: ListAnswerEntry): void {
    this.http.get('/a/delete/' + poll.Segment, {responseType: 'text'}).pipe(take(1)).subscribe({
      next: () => this.refresh(),
      error: (err: HttpErrorResponse) => this.handleError(err),
    });
  }

  constructor(
    private http: HttpClient,
    private router: Router,
  ) { }

  private handleError(err: HttpErrorResponse): void {
    this._error.next(err);
  }
}
