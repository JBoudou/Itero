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

import { Component, OnInit, ViewEncapsulation } from '@angular/core';

import { BehaviorSubject, Observable } from 'rxjs';

import { CreateService, CreateNextStatus } from './create.service';
import { NavStepStatus } from './navtree/navstep.status';

@Component({
  selector: 'app-create',
  templateUrl: './create.component.html',
  styleUrls: ['./create.component.sass'],
  encapsulation: ViewEncapsulation.None,
})
export class CreateComponent implements OnInit {

  private _stepStatus$ = new BehaviorSubject<NavStepStatus>(undefined);
  private _nextStatus$ = new BehaviorSubject<CreateNextStatus>(undefined);

  get stepStatus$(): Observable<NavStepStatus> {
    return this._stepStatus$;
  }

  get nextStatus$(): Observable<CreateNextStatus> {
    return this._nextStatus$;
  }

  constructor(
    private service: CreateService,
  ) {
  }

  ngOnInit(): void {
    this.service.createStepStatus$.subscribe(this._stepStatus$);
    this.service.createNextStatus$.subscribe(this._nextStatus$);
  }

  onJump(pos: number): void {
    if (!this._stepStatus$.value.isNavigable(pos, this._nextStatus$.value.validable)) {
      return;
    }
    const current = this._stepStatus$.value.current;
    if (pos < current) {
      this.service.back(current - pos);
    } else {
      this.service.next();
    }
  }

}
