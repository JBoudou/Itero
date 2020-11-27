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

import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, ValidationErrors, Validators } from '@angular/forms';

import { SignupQuery } from '../api';
import { SessionService } from '../session/session.service';

@Component({
  selector: 'app-signup',
  templateUrl: './signup.component.html',
  styleUrls: ['./signup.component.sass']
})
export class SignupComponent implements OnInit {

  form = this.formBuilder.group({
    Name: ['', [
      Validators.required,
      Validators.minLength(5)
    ]],
    Email: ['', [
      Validators.required,
      Validators.email,
    ]],
    Passwd: ['', [
      Validators.required,
      Validators.minLength(5)
    ]],
    pwdconfirm: [''],
  }, {
    validators: [ this.samePasswordValidator ],
  });

  serverError = ''
  createdUser = ''

  constructor(
    private session: SessionService,
    private http: HttpClient,
    private formBuilder: FormBuilder,
  ) {}

  ngOnInit(): void {
  }

  onSignup(): void {
    let toSend = this.form.value;
    delete toSend.pwdconfirm;
    toSend.Name = toSend.Name.trim()
    this.http.post('/a/signup', toSend as SignupQuery)
      .pipe(this.session.httpOperator(toSend.Name))
      .subscribe({
      next: () => {
        this.createdUser = toSend.Name;
      },
      error: (err: HttpErrorResponse) => {
        if (err.status == 400) {
          this.serverError = err.error.trim()
          switch (this.serverError) {
            case 'Name too short':
            case 'Name has spaces':
            case 'Already exists':
              this.form.get('Name').valueChanges.subscribe(this.resetServerError);
              break;
            case 'Email invalid':
              this.form.get('Email').valueChanges.subscribe(this.resetServerError);
              break;
            case 'Passwd too short':
              this.form.get('Email').valueChanges.subscribe(this.resetServerError);
              break;
          }
        } else {
          this.serverError = err.message
        }
      }
    });
  }

  private resetServerError = { next: () => { this.serverError = '' } }

  samePasswordValidator(grp: FormGroup): ValidationErrors | null {
    const first  = grp.get('Passwd');
    const second = grp.get('pwdconfirm');
    return first.value != second.value ? { passwordsDiffer: true } : null;
  }

  nameTooShort(): boolean {
    return true &&
      this.form.controls['Name'].dirty &&
      !this.form.controls['Name'].valid &&
      !!this.form.controls['Name'].errors['minlength'];
  }

  pwdTooShort(): boolean {
    return true &&
      this.form.controls['Passwd'].dirty &&
      !this.form.controls['Passwd'].valid &&
      !!this.form.controls['Passwd'].errors['minlength'];
  }

  emailWrong(): boolean {
    return true &&
      this.form.controls['Email'].dirty &&
      !this.form.controls['Email'].valid &&
      !!this.form.controls['Email'].errors['email'];
  }

  passwordsDiffer(): boolean {
    return true &&
      !!this.form.errors &&
      !!this.form.errors['passwordsDiffer'];
  }
}
