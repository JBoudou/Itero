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
import {FormGroup} from '@angular/forms';

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

  private _state = new BehaviorSubject<ConfirmState>({ type: 'loading' })
  get state$(): Observable<ConfirmState> {
    return this._state
  }

  constructor(
    private http: HttpClient,
    private route: ActivatedRoute,
  ) { }

  private _segment: string
  ngOnInit(): void {
    this.route.paramMap.pipe(take(1)).subscribe((params: ParamMap) => {
      this._segment = params.get('confirmSegment')
      this.http.get<ConfirmAnswer>('/a/confirm/' + this._segment).pipe(take(1)).subscribe({
        next: (answer: ConfirmAnswer) => {
          this._state.next({ type: answer.Type })
        },
        error: (err: HttpErrorResponse) => {
          if (err.status == 404) {
            this._state.next({ type: 'notfound' })
          } else {
            this._state.next({ type: 'error', data: new ServerError(err, 'retrieving confirmation') })
          }
        },
      })
    })
  }

  // Change password //

  passwdForm: FormGroup = new FormGroup({})
  passwdErrors: Set<string> = new Set<string>()

  onPwdErrors(evt: Set<string>): void {
    this.passwdErrors = evt
  }

  onChangePassword():void {
    let toSend = this.passwdForm.value;
    delete toSend.pwdconfirm;
    this.http.post('/a/passwd/' + this._segment, toSend).pipe(take(1)).subscribe({
      next: () =>
        this._state.next({ type: 'passwd changed' }),
      error: (err: HttpErrorResponse) =>
        this._state.next({ type: 'error', data: new ServerError(err, 'changing password') }),
    })
  }

}
