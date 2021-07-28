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

import { ServerError } from './server-error'

import { HttpErrorResponse } from '@angular/common/http';

describe('ServerError', () => {

  it('construct non-error', () => {
    const err = new ServerError()

    expect(err.ok).toBeTrue()
    expect(err.status).toBe(200)
  })

  it('construct HTTP error', () => {
    const http = new HttpErrorResponse({status: 404, error: 'Not found'})
    const err = new ServerError(http, 'testing')

    expect(err.ok).toBeFalse()
    expect(err.status).toBe(http.status)
    expect(err.message).toBe(http.error)
  })

  it('construct HTTP non-error', () => {
    const http = new HttpErrorResponse({status: 200})
    const err = new ServerError(http, 'testing')

    expect(err.ok).toBeTrue()
    expect(err.status).toBe(200)
  })

  it('construct non-HTTP error', () => {
    const http = new HttpErrorResponse({error: new Error('custom')})
    const err = new ServerError(http, 'testing')

    expect(err.ok).toBeFalse()
    expect(err.message).toBe('custom')
  })

})
