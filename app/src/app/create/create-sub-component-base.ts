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

/** Implementation of CreateSubComponent using a FormGroup. */
export abstract class CreateSubComponentBase {

  /**
   * Form to edit parts of the query.
   * The name of the controls must match the names of properties of the query.
   */
  abstract form: FormGroup;

  protected abstract service: CreateService;
  protected abstract route: ActivatedRoute;

  protected subscriptions: Subscription[] = [];

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
    this.subscriptions.push(
      this.service.query$.subscribe({
        next: (query: Partial<CreateQuery>) => this.form.patchValue(query),
      })
    );

    // Once stepSegment is known, subscribe to valueChanges for each control in this.form.
    this.route.url.pipe(take(1)).subscribe({
      next: (segments: UrlSegment[]) => {
        const stepSegment = segments[segments.length - 1].toString();
        for (const prop in this.form.controls) {
          const control = this.form.controls[prop];
          this.subscriptions.push(
            control.valueChanges.subscribe({
              next: value => this.service.patchQuery(stepSegment, { [prop]: value }),
            })
          );
        };
      },
    });
  }

  protected unsubscribeAll(): void {
    for (const sub of this.subscriptions) {
      sub.unsubscribe();
    }
  }

}
