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

import { FormGroup } from '@angular/forms';
import { ActivatedRoute, UrlSegment } from '@angular/router';

import { Observable, Subscription } from 'rxjs';
import { map, startWith, take } from 'rxjs/operators';

import { CreateService } from '../create/create.service';
import { CreateQuery } from '../api';

/** Base class to implement editing component using CreateService. */
export abstract class CreateSubComponentBase {

  /**
   * Form to edit parts of the query.
   * The name of the controls must match the names of properties of the query.
   */
  abstract form: FormGroup;

  protected subscriptions: Subscription[] = [];

  // Fields that are assumed to be injected by the subclass.
  protected abstract service: CreateService;
  protected abstract route: ActivatedRoute;

  constructor(
  ) { }

  /** Notifications are send each time the status of the form changes. */
  get validable$(): Observable<boolean> {
    return this.form.statusChanges.pipe(map(val => val == 'VALID'), startWith(this.form.valid));
  }

  /**
   * Connects the service with the form.
   * Must be called in some initializing method of the subclass, usually ngOnInit.
   */
  protected initModel(): void {
    this.route.url.pipe(take(1)).subscribe({
      next: (segments: UrlSegment[]) => {
        const stepSegment = segments[segments.length - 1].toString();
        this._initModel(stepSegment);
      }
    });
  }

  /**
   * Unsubscribe from subcriptions.
   * Must be called before the component is destroyed, usually ngOnDestroy.
   */
  protected unsubscribeAll(): void {
    for (const sub of this.subscriptions) {
      sub.unsubscribe();
    }
  }

  /**
   * Modify the partial query sent to the model.
   * Could be overloaded in the component. By default, return the parameter unchanged.
   * Beware that the properties of the original query may be references to value stored in some form
   * controls.
   */
  protected modifyQueryToSend(query: Partial<CreateQuery>): Partial<CreateQuery> {
    return query;
  }

  /** Callback called just after the form is refreshed with the query from the service. */
  protected afterQueryFetch(query: Partial<CreateQuery>): void {
  }
  
  private _initModel(stepSegment: string): void {
    this.subscriptions.push(
      this.service.query$.subscribe({
        next: (query: Partial<CreateQuery>) => this._synchronize(stepSegment, query),
      })
    );
  }

  private _first_synchronize_done: boolean = false;

  /** Synchronize the view from the model. */
  private _synchronize(stepSegment: string, query: Partial<CreateQuery>) {
    this.form.patchValue(query)
    this.afterQueryFetch(query)

    if (!this._first_synchronize_done) {
      let toSend: any = {};
      let somethingToSend: boolean = false;
      for (const prop in this.form.controls) {
        const val = this.form.controls[prop].value;
        if (query[prop] === undefined && val !== undefined) {
          toSend[prop] = val;
          somethingToSend = true;
        }
      }
      if (somethingToSend) {
        this.service.patchQuery(stepSegment, this.modifyQueryToSend(toSend), { defaultValues: true });
      }

      // Subscribe to valueChanges for each control in this.form.
      for (const prop in this.form.controls) {
        const control = this.form.controls[prop];
        this.subscriptions.push(
          control.valueChanges.subscribe({
            next: value => this.service.patchQuery(stepSegment, this.modifyQueryToSend({ [prop]: value })),
          })
        );
      };

      this._first_synchronize_done = true;
    }
  }

}
