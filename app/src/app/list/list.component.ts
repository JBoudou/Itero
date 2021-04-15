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

import { Observable } from 'rxjs';

import { ListAnswerEntry } from '../api';
import { ListService } from './list.service';

/**
 * The list of polls.
 */
@Component({
  selector: 'app-list',
  templateUrl: './list.component.html',
  styleUrls: ['./list.component.sass']
})
export class ListComponent implements OnInit {

  get publicList$(): Observable<ListAnswerEntry[]> {
    return this.service.publicList$;
  }
  get ownList$(): Observable<ListAnswerEntry[]> {
    return this.service.ownList$;
  }

  constructor(
    private service: ListService
  ) { }

  ngOnInit(): void {
    // Retrieve the list of polls each time the component is displayed.
    this.service.refresh();
  }

}
