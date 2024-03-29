// Itero - Online iterative vote applicationj
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

import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { FormBuilder, Validators } from '@angular/forms';
import { Router } from '@angular/router';

import { take } from 'rxjs/operators';
import {ServerError} from 'src/app/shared/server-error';

import { SessionService } from '../session.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.sass']
})
export class LoginComponent implements OnInit {

  form = this.formBuilder.group({
    User: ['',
      [Validators.required,
       Validators.minLength(3),
       Validators.pattern(/^[^\s].*[^\s]$/)
      ]],
    Passwd: ['',
      [Validators.required,
       Validators.minLength(4)
      ]]
  });

  showPassword: boolean = false;

  errorType = 'None'
  errorMsg = ''

  serverError: ServerError = new ServerError()
  forgotSent: boolean = false

  constructor(
    private session: SessionService,
    private http: HttpClient,
    private formBuilder: FormBuilder,
    private router: Router
  ) {}

  ngOnInit(): void {
  }

  onLogin(): void {
    this.http.post('/a/login', this.form.value)
      .pipe(this.session.httpOperator(this.form.value.User), take(1))
      .subscribe({
      next: _ => {
        this.reset()
        this.router.navigateByUrl(this.session.getLoginRedirectionUrl());
      },
      error: (err: HttpErrorResponse) => {
        this.reset()
        if (err.status == 403) {
          this.errorType = 'Wrong';
        } else {
          this.serverError = new ServerError(err, 'loging in')
        }
      }
    });
  }

  onForgot(): void {
    this.http.post('/a/forgot', this.form.value)
      .pipe(take(1))
      .subscribe({
      next: _ => {
        this.reset()
        console.log('sent')
        this.forgotSent = true
      },
      error: (err: HttpErrorResponse) => {
        this.reset()
        if (err.status == 403) {
          this.errorType = 'Unknown';
        } else if (err.status == 409) {
          this.errorType = 'Already'
        } else {
          this.serverError = new ServerError(err, 'requesting a password change')
        }
      }
    });
  }

  private reset(): void {
    this.errorType = 'None'
    this.serverError = new ServerError()
    this.forgotSent = false
  }

}
