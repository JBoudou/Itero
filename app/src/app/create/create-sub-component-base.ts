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

import { Observable } from 'rxjs';
import { map, startWith } from 'rxjs/operators';

import { CreateService, CreateSubComponent } from '../create/create.service';

/** Implementation of CreateSubComponent using a FormGroup. */
export abstract class CreateSubComponentBase implements CreateSubComponent {

  /**
   * Form to edit parts of the query.
   * The name of the controls must match the names of properties of the query.
   */
  abstract form: FormGroup;

  protected abstract service: CreateService;

  constructor(
  ) { }

  /** The handled fields are the names of the controls of this.form. */
  get handledFields(): Set<string> {
    return new Set<string>(Object.keys(this.form.controls));
  }

  /** Notifications are send each time the status of the form changes. */
  get validable$(): Observable<boolean> {
    return this.form.statusChanges.pipe(map(val => val == 'VALID'), startWith(this.form.valid));
  }

  private _started: boolean;

  isStarted(): boolean {
    return this._started || this.form.dirty;
  }

  /**
   * Connects the service with the form.
   * Must be called in some initializing method of the subclass, usually ngOnInit.
   */
  protected initModel(): void {
    this._started = false;
    const query = this.service.register(this);
    for (const prop in this.form.controls) {
      if (query[prop] !== undefined) {
        this._started = query[prop] != this.form.controls[prop].value;
        this.form.controls[prop].setValue(query[prop]);
      } else {
        query[prop] = this.form.controls[prop].value;
      }
      this.form.controls[prop].valueChanges.subscribe({
        next: value => query[prop] = value,
      });
    }
  }

}
