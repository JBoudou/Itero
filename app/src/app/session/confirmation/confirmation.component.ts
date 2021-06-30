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

import { Component, OnInit, ChangeDetectionStrategy } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { ActivatedRoute, ParamMap } from '@angular/router';

import { take } from 'rxjs/operators';

import { ConfirmAnswer } from 'src/app/api';
import { ServerError } from 'src/app/shared/server-error';
import {BehaviorSubject, Observable} from 'rxjs';

interface ConfirmState {
  type: string
  data?: ServerError
}

@Component({
  selector: 'app-confirmation',
  templateUrl: './confirmation.component.html',
  styleUrls: ['./confirmation.component.sass'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ConfirmationComponent implements OnInit {

  private state = new BehaviorSubject<ConfirmState>({ type: 'loading' })
  get state$(): Observable<ConfirmState> {
    return this.state
  }

  constructor(
    private http: HttpClient,
    private route: ActivatedRoute,
  ) { }

  ngOnInit(): void {
    this.route.paramMap.pipe(take(1)).subscribe((params: ParamMap) => {
      this.http.get<ConfirmAnswer>('/a/confirm/' + params.get('confirmSegment')).pipe(take(1)).subscribe({
        next: (answer: ConfirmAnswer) => {
          this.state.next({ type: answer.Type })
        },
        error: (err: HttpErrorResponse) => {
          if (err.status == 404) {
            this.state.next({ type: 'notfound' })
          } else {
            this.state.next({ type: 'error', data: new ServerError(err, 'retrieving confirmation') })
          }
        },
      })
    })
  }

}
