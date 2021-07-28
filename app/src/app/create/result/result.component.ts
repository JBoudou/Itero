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

import { take } from 'rxjs/operators';

import { CreateService } from '../create.service';
import { CreateQuery } from 'src/app/api';

type Result = {status: number, message: string} | Partial<CreateQuery>;

function isCreateQuery(result: Result): result is Partial<CreateQuery> {
  const asError = result as any;
  return asError.status === undefined && asError.message === undefined;
}


@Component({
  selector: 'app-create-result',
  templateUrl: './result.component.html',
  styleUrls: [ './result.component.sass']
})
export class ResultComponent implements OnInit {

  private pollSegment: string;
  private result: Result = { status: 999, message: 'undefined' };

  constructor(
    private route: ActivatedRoute,
    private service: CreateService,
  ) { }

  ngOnInit(): void {
    this.route.paramMap.pipe(take(1)).subscribe((params: ParamMap) => {
      this.pollSegment = params.get('pollSegment');
    });

    const result = this.service.getResult();
    if (result === undefined) {
      this.result = {
        status: 999,
        message: 'Front end error',
      };
    } else {
      this.result = result;
    }
  }

  isSuccess(): boolean {
    return isCreateQuery(this.result);
  }

  link(): string {
    const prefix : string = window.location.protocol + '//' + window.location.host
    return (isCreateQuery(this.result) && this.result.ShortURL) ?
      prefix + '/p/' + this.result.ShortURL :
      prefix + '/r/poll/' + this.pollSegment;
  }

  error(): {status: number, message: string} | undefined {
    return isCreateQuery(this.result) ? undefined : this.result;
  }

  query(): Partial<CreateQuery> | undefined {
    return isCreateQuery(this.result) ? this.result : undefined;
  }

}
