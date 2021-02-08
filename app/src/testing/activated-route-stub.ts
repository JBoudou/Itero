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

import { convertToParamMap, ParamMap, Params, UrlSegment } from '@angular/router';

import { ReplaySubject } from 'rxjs';

/**
 * An ActivateRoute test double with a `paramMap` observable.
 * Use the `setParamMap()` method to add the next `paramMap` value.
 */
export class ActivatedRouteStub {
  private _paramMap$ = new ReplaySubject<ParamMap>();

  /** The mock paramMap observable */
  readonly paramMap = this._paramMap$.asObservable();

  private _url$ = new ReplaySubject<UrlSegment[]>();

  readonly url = this._url$.asObservable();

  constructor(initialParams?: Params) {
    if (initialParams !== undefined) {
      this.setParamMap(initialParams);
    }
  }

  /** Set the paramMap observables's next value */
  setParamMap(params?: Params) {
    this._paramMap$.next(convertToParamMap(params));
  }

  nextUrlFromString(url: string) {
    this._url$.next(url.split('/').map((segment: string) => new UrlSegment(segment, {})));
  }

}
