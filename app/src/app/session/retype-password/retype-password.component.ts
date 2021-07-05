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

import { Component, EventEmitter, Input, OnDestroy, OnInit, Output } from '@angular/core';
import { FormBuilder, FormGroup, ValidationErrors, Validators } from '@angular/forms';

import { Subscription } from 'rxjs';

import { setEqual } from 'src/app/shared/collections';

function samePasswordValidator(grp: FormGroup): ValidationErrors | null {
  const first  = grp.get('Passwd');
  const second = grp.get('pwdconfirm');
  return first.value != second.value ? { passwordsDiffer: true } : null;
}

/**
 * RetypePasswordComponent displays two fields to enter a password.
 * It takes a FormGroup as input, adding controls 'Passwd' and 'pwdconfirm', connected to the two fields.
 * The component check the correctness of the password, outputing set of errors.
 */
@Component({
  selector: 'app-retype-password',
  templateUrl: './retype-password.component.html',
  styleUrls: ['./retype-password.component.sass']
})
export class RetypePasswordComponent implements OnInit, OnDestroy {

  @Input('controlGroup') form!: FormGroup
  @Input() passwdLabel: string = "Password"

  @Output() errors: EventEmitter<Set<string>> = new EventEmitter<Set<string>>()

  constructor(
    private formBuilder: FormBuilder,
  ) { }

  private _subscriptions : Subscription[] = []

  ngOnInit(): void {
    this.form.addControl('Passwd', this.formBuilder.control('', [
      Validators.required,
      Validators.minLength(5)
    ]))
    this.form.addControl('pwdconfirm', this.formBuilder.control(''))
    this.form.setValidators(Validators.compose([this.form.validator, samePasswordValidator]))

    this._subscriptions.push(this.form.valueChanges.subscribe(() => this.generateErrors()))
  }

  ngOnDestroy(): void {
    for (let subscription of this._subscriptions) {
      subscription.unsubscribe()
    }
  }

  private _lastErrors: Set<string> = new Set<string>()
  private generateErrors(): void {
    const newErrors: Set<string> = new Set<string>()
    if (  this.form.controls['Passwd'].dirty &&
         !this.form.controls['Passwd'].valid &&
        !!this.form.controls['Passwd'].errors['minlength']) {
      newErrors.add('pwdTooShort')
    }
    if (!!this.form.errors &&
        !!this.form.errors['passwordsDiffer']) {
      newErrors.add('passwordsDiffer')
    }

    if (!setEqual(newErrors, this._lastErrors)) {
      this._lastErrors = newErrors
      this.errors.emit(newErrors)
    }
  }

}
