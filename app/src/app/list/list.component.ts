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

import { Observable } from 'rxjs';

import { ListAnswerEntry } from '../api';
import { ServerError } from '../shared/server-error';
import { ListService } from './list.service';

/**
 * The list of polls.
 */
@Component({
  selector: 'app-list',
  templateUrl: './list.component.html',
  styleUrls: ['./list.component.sass']
})
export class ListComponent implements OnInit, OnDestroy {

  get publicList$(): Observable<ListAnswerEntry[]> {
    return this.service.publicList$;
  }
  get ownList$(): Observable<ListAnswerEntry[]> {
    return this.service.ownList$;
  }
  get error$(): Observable<ServerError> {
    return this.service.error$;
  }

  constructor(
    private service: ListService
  ) { }

  ngOnInit(): void {
    this.service.activate();
  }

  ngOnDestroy(): void {
    this.service.desactivate();
  }

}
