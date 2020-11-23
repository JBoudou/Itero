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

import { SessionService } from '../session.service';
import { ListResponseEntry } from '../api';

@Component({
  selector: 'app-list',
  templateUrl: './list.component.html',
  styleUrls: ['./list.component.sass']
})
export class ListComponent implements OnInit {

  polls: ListResponseEntry[]

  constructor(private session: SessionService
             ) {
    this.polls = [];
  }

  ngOnInit(): void {
    var url: string = this.session.makeURL('/a/list');
    this.session.http.get<ListResponseEntry[]>(url).subscribe({
      next: (values: ListResponseEntry[]) => this.polls = values,
      error: (_) => this.polls = []
    });
  }

}
