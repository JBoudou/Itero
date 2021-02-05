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

import { Component, ChangeDetectionStrategy, OnInit, ViewEncapsulation, TemplateRef } from '@angular/core';

import { BehaviorSubject, Observable, Subscription, Subject } from 'rxjs';
import { delay } from 'rxjs/operators';

import { CreateService } from './create.service';
import { NavStepStatus } from './navtree/navstep.status';

@Component({
  selector: 'app-create',
  templateUrl: './create.component.html',
  styleUrls: ['./create.component.sass'],
  encapsulation: ViewEncapsulation.None,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class CreateComponent implements OnInit {

  private _stepStatus$ = new BehaviorSubject<NavStepStatus>(undefined);
  private _validable$  = new BehaviorSubject<boolean>(false);

  get stepStatus$(): Observable<NavStepStatus> {
    return this._stepStatus$;
  }

  get validable$(): Observable<boolean> {
    return this._validable$;
  }

  infoContext: any;

  infoTemplate$ = new Subject<TemplateRef<any>>();

  constructor(
    private service: CreateService,
  ) {
    this.service.stepStatus$.subscribe(this._stepStatus$);
    this.infoContext = { $implicit: true, query$: this.service.query$ };
  }

  ngOnInit(): void {
  }

  onJump(pos: number): void {
    if (!this._stepStatus$.value.isNavigable(pos, this._validable$.value)) {
      return;
    }
    const current = this._stepStatus$.value.current;
    if (pos < current) {
      this.service.back(current - pos);
    } else {
      this.service.next();
    }
  }

  private _validableSubscription: Subscription | undefined;

  onActivate(component: any): void {
    if ('validable$' in component) {
      if (this._validableSubscription !== undefined) {
        this._validableSubscription.unsubscribe();
      }
      // The first event needs to be delayed because otherwise it will happen during the same
      // rendering cycle but after its parent (the current CreateComponent), resulting in an error.
      this._validableSubscription = component.validable$.pipe(delay(0)).subscribe(this._validable$);
    }
    setTimeout(() => this.infoTemplate$.next(component.infoTemplate), 0);
  }

  onDesactivate(component: Object): void {
    if ("validable$" in component && this._validableSubscription !== undefined) {
      this._validableSubscription.unsubscribe();
      this._validableSubscription = undefined;
      this._validable$.next(false);
    }
  }

}
