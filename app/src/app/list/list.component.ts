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

import { Component, OnInit } from '@angular/core';
import {Router} from '@angular/router';
import {HttpClient} from '@angular/common/http';

import { ListAnswerEntry, PollAction } from '../api';

function mapListAnswerEntry(e: ListAnswerEntry): ListAnswerEntry {
  const today = new Date(Date.now());
  today.setHours(0, 0, 0);

  if (e.Deadline as string == '⋅') {
    e.Deadline = today;
    e.deadlineCategory = 'None';
    return e;
  }

  e.Deadline = new Date(e.Deadline);
  if (e.Deadline < today) {
    e.deadlineCategory = 'None';
    return e;
  }

  const tomorrow = new Date(today.valueOf() + (24 * 3600 * 1000));
  if (e.Deadline < tomorrow) {
    e.deadlineCategory = 'Today';
    return e;
  }

  const afterTomorrow = new Date(today.valueOf() + (2 * 24 * 3600 * 1000));
  if (e.Deadline < afterTomorrow) {
    e.deadlineCategory = 'Tomorrow';
    return e;
  }

  const nextWeek = new Date(today.valueOf() + (8 * 24 * 3600 * 1000));
  if (e.Deadline < nextWeek) {
    e.deadlineCategory = 'Week';
    return e;
  }

  const nextMonth =  new Date(today.valueOf() + (33 * 24 * 3600 * 1000));
  if (e.Deadline < nextMonth) {
    e.deadlineCategory = 'Month';
    return e;
  }

  e.deadlineCategory = 'Year';
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
export class ListComponent implements OnInit {

  polls: ListAnswerEntry[];

  constructor(
    private router: Router,
    private http: HttpClient
  ) {
    this.polls = [];
  }

  pollActionString(poll: ListAnswerEntry): String {
    switch (poll.Action) {
    case PollAction.Vote:
      return 'Vote';
    case PollAction.Modi:
      return 'Modi';
    case PollAction.Part:
      return 'Part';
    case PollAction.Term:
      return 'Term';
    }
  }

  terminated(poll: ListAnswerEntry): boolean {
    return poll.Action == PollAction.Term;
  }

  ngOnInit(): void {
    // Retrieve the list of polls each time the component is displayed.
    this.http.get<ListAnswerEntry[]>('/a/list').subscribe({
      next: (values: ListAnswerEntry[]) => this.polls = values.map(mapListAnswerEntry),
      error: (_) => this.polls = []
    });
  }

  /**
   * Receives click event on individual poll.
   * @Param {string} segment The identifier of the poll.
   */
  go(segment: string): void {
    this.router.navigateByUrl('/r/poll/' + segment)
  }

}
