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

import {BreakpointObserver, BreakpointState} from '@angular/cdk/layout';
import { Injectable } from '@angular/core';

import { Observable, of } from 'rxjs';
import { map } from 'rxjs/operators';

export enum ResponsiveState {
  SmartPhone,
  Tablet,
  Laptop,
}

@Injectable({
  providedIn: 'root'
})
export class ResponsiveBreakpointService {

  state$: Observable<ResponsiveState>

  constructor(
    private breakpointObserver: BreakpointObserver,
  ) {
    const spy = document.getElementById('breakpoints-spy')
    if (spy instanceof Element) {
      const bpStyle = getComputedStyle(spy)
      const tablet = `(max-width: ${bpStyle.getPropertyValue('min-width')})`
      const laptop = `(max-width: ${bpStyle.getPropertyValue('max-width')})`
      this.state$ = breakpointObserver
        .observe([ tablet, laptop ])
        .pipe(map((st: BreakpointState): ResponsiveState => {
          const brkpts = st.breakpoints
          return brkpts[tablet] ? ResponsiveState.SmartPhone :
                brkpts[laptop] ? ResponsiveState.Tablet : ResponsiveState.Laptop
        }))
    } else {
      this.state$ = of(ResponsiveState.Laptop)
    }
  }
}
