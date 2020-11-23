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

import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

import { LoginInfo } from './api'

@Injectable({
  providedIn: 'root'
})
export class SessionService {

  sessionId: string;

  login(info: LoginInfo): Observable<LoginInfo> {
    return this.http.post('/a/login', info).pipe(
      map((data: string) => {
        this.sessionId = data;
        console.log("New session id " + this.sessionId);
        return info;
      })
    );
  }

  constructor(private http: HttpClient) { }
}
