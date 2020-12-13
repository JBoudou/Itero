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

import { Component, Input, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';

import { PollSubComponent } from '../poll/common';
import { CountInfoEntry, CountInfoAnswer } from '../api';

@Component({
  selector: 'app-counts-information',
  templateUrl: './counts-information.component.html',
  styleUrls: ['./counts-information.component.sass']
})
export class CountsInformationComponent implements OnInit, PollSubComponent {

  @Input() pollSegment: string;

  data: Array<{name: string; value: number}>;

  constructor(
    private http: HttpClient,
  ) { }

  ngOnInit(): void {
    this.http.get<CountInfoAnswer>('/a/info/count/' + this.pollSegment).subscribe({
      next: (answer: CountInfoAnswer) => {
        this.data = [];
        for (let entry of answer.Result) {
          console.log(entry.Alternative.Name);
          this.data.push({"name": entry.Alternative.Name, "value": entry.Count});
        }
      }
    });
  }

}
