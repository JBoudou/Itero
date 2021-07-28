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

import { HttpErrorResponse } from '@angular/common/http';

export class ServerError {
  readonly ok: boolean;
  readonly status: number;
  readonly message: string;
  readonly context: string;

  constructor();
  constructor(err: HttpErrorResponse, context: string);
  constructor(err?: HttpErrorResponse, context?: string) {
    this.ok = err === undefined || (err.status > 99 && err.status < 400)
    this.status = err && err.status || 200;
    this.context = err && context || '';
    this.message = typeof(err?.error) === 'string' ? err.error.trim() :
                   err?.error instanceof Error ? err.error.message :
                   ''
  }
}
