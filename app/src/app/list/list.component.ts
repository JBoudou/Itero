// Itero - Online iterative vote application
// Copyright (C) 2020 Joseph Boudou
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

import { Component, OnDestroy, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';

import { BehaviorSubject, Observable } from 'rxjs';
import { take } from 'rxjs/operators';

import { ListAnswer, ListAnswerEntry } from '../api';

function mapListAnswerEntry(e: ListAnswerEntry): ListAnswerEntry {
  e.Deadline = e.Deadline as string == '⋅' ? undefined : new Date(e.Deadline);
  return e;
}

/**
 * The list of polls.
 */
@Component({
  selector: 'app-list',
  templateUrl: './list.component.html',
  styleUrls: ['./list.component.sass']
})
export class ListComponent implements OnInit, OnDestroy {

  private _answer = new BehaviorSubject<ListAnswer>({Public: [], Own: []});

  get answer$(): Observable<ListAnswer> {
    return this._answer;
  }

  constructor(
    private http: HttpClient
  ) { }

  ngOnInit(): void {
    // Retrieve the list of polls each time the component is displayed.
    this.http.get<ListAnswer>('/a/list').pipe(take(1)).subscribe({
      next: (answer: ListAnswer) =>
        // TODO: Use more standard solution (fromJSON ?).
        this._answer.next({Public: answer.Public.map(mapListAnswerEntry), Own: answer.Own.map(mapListAnswerEntry)}),
    });
  }

  ngOnDestroy(): void {
    this._answer.complete();
  }

}
