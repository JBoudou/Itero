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

import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, ParamMap } from '@angular/router';

import { CreateService } from '../create.service';

@Component({
  selector: 'app-create-result',
  templateUrl: './result.component.html',
  styleUrls: [ './result.component.sass']
})
export class ResultComponent implements OnInit {

  private pollSegment: string;

  constructor(
    private route: ActivatedRoute,
    private service: CreateService,
  ) { }

  ngOnInit(): void {
    this.route.paramMap.subscribe((params: ParamMap) => {
      this.pollSegment = params.get('pollSegment');
    });
  }

  isSuccess(): boolean {
    return this.pollSegment != 'error';
  }

  link(): string {
    return window.location.protocol + '//' + window.location.host + '/r/poll/' + this.pollSegment;
  }

  error(): {status: number, message: string}|undefined {
    return this.service.httpError === undefined ? undefined : {
      status: this.service.httpError.status,
      message: this.service.httpError.error.trim(),
    };
  }

}
