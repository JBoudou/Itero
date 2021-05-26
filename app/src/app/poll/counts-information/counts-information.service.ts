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

import { Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';

import { take } from 'rxjs/operators';

import { CountInfoAnswer } from 'src/app/api';

@Injectable()
export class CountsInformationService {

  constructor(
    private http: HttpClient,
  ) { }

  // For now, the answer is just fetched from the middleware. In the future, caching may be
  // implemented.
  information(pollSegment: string, round?: number): Promise<CountInfoAnswer> {
    const url = round === undefined ? '/a/info/count/' + pollSegment :
                                      '/a/info/count/' + round + '/' + pollSegment;
    return new Promise((resolve: (value?: CountInfoAnswer | PromiseLike<CountInfoAnswer>) => void,
                        reject: (reason?: any) => void) =>
      this.http.get<CountInfoAnswer>(url)
        .pipe(take(1)).subscribe({
          next: (answer: CountInfoAnswer) => resolve(answer),
          error: (err: HttpErrorResponse) => reject({status: err.status, message: err.error.trim()})
        })
    );
  }
  
}
