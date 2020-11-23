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

import {HttpErrorResponse} from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { FormBuilder, Validators } from '@angular/forms';
import { Router } from '@angular/router';

import { SessionService } from '../session.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.sass']
})
export class LoginComponent implements OnInit {

  form = this.formBuilder.group({
    User: ['', [Validators.required, Validators.minLength(3), Validators.pattern(/^[^\s].*[^\s]$/)]],
    Passwd: ['', [Validators.required, Validators.minLength(4)]]
  });


  errorType = 'None'
  errorMsg = ''

  constructor(private session: SessionService,
              private formBuilder: FormBuilder,
              private router: Router
             ) { }

  ngOnInit(): void {
  }

  onLogin() {
    this.session.login(this.form.value).subscribe({
      next: _ => {
        this.errorType = 'None';
        this.router.navigate(['r', 'list']);
      },
      error: (err: HttpErrorResponse) => {
        if (err.error instanceof ErrorEvent) {
          this.errorType = 'Local';
          this.errorMsg  = err.error.message;
        } else if (err.status == 403) {
          this.errorType = 'Wrong';
        } else {
          this.errorType = 'Server';
          this.errorMsg = err.statusText;
        }
      }
    });
  }

}
