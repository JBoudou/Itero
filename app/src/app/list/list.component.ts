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

import { take } from 'rxjs/operators';

import { ListAnswerEntry, PollAction } from '../api';

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
    this.http.get<ListAnswerEntry[]>('/a/list').pipe(take(1)).subscribe({
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
