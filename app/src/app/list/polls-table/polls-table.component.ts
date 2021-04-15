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


import { Component, Input, OnInit } from '@angular/core';

import { ListAnswerEntry, PollAction } from '../../api';
import { ListService } from '../list.service';

@Component({
  selector: 'app-polls-table',
  templateUrl: './polls-table.component.html',
  styleUrls: ['./polls-table.component.sass']
})
export class PollsTableComponent implements OnInit {

  @Input() polls: ListAnswerEntry[];

  constructor(
    public service: ListService
  ) { }

  ngOnInit(): void {
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
    case PollAction.Wait:
      return 'Wait';
    }
  }

  terminated(poll: ListAnswerEntry): boolean {
    return poll.Action == PollAction.Term;
  }

}
