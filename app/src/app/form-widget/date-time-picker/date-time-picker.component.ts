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

import { Component, OnInit, OnDestroy, Input, Self, Optional } from '@angular/core';
import { ControlValueAccessor, NgControl, FormBuilder } from '@angular/forms';

import { Subscription } from 'rxjs';

@Component({
  selector: 'app-date-time-picker',
  templateUrl: './date-time-picker.component.html',
  styleUrls: ['./date-time-picker.component.sass']
})
export class DateTimePickerComponent implements OnInit, OnDestroy, ControlValueAccessor {

  @Input() disabled: boolean;
  
  form = this.formBuilder.group({
    date: '',
    time: '',
  });

  constructor(
    @Self() @Optional() private ngControl: NgControl,
    private formBuilder: FormBuilder,
  ) {
    if (this.ngControl) {
      this.ngControl.valueAccessor = this;
    }
  }

  private _subscriptions: Subscription[] = [];

  ngOnInit(): void {
    this._subscriptions.push(this.form.valueChanges.subscribe({ next: _ => this.onValueChange() }));
    if (this.disabled) {
      this.form.disable({onlySelf: true});
    }
  }

  ngOnDestroy(): void {
    for (const sub of this._subscriptions) {
      sub.unsubscribe();
    }
  }

  onValueChange(): void {
    const date = this.form.value.date.split('-').map((str: string) => parseInt(str));
    const time = this.form.value.time.split(':').map((str: string) => parseInt(str));
    // No spread syntax yet :(
    const value = new Date(date[0], date[1] - 1, date[2], time[0], time[1]);
    this.notifChange(value);
  }

  /** Implements ControlValueAccessor */

  writeValue(obj: any): void {
    if (!(obj instanceof Date)) {
      console.warn("DateTimePickerComponent unknown value type " + typeof(obj));
      return;
    }

    const date = obj.getFullYear() + '-' +
                 (obj.getMonth() + 1).toString().padStart(2, '0') + '-' +
                 obj.getDate()   .toString().padStart(2, '0');
    const time = obj.getHours()  .toString().padStart(2, '0') + ':' +
                 obj.getMinutes().toString().padStart(2, '0');
    this.form.setValue({date: date, time: time});
  }

  setDisabledState(isDisabled: boolean): void {
    if (isDisabled) {
      this.form.disable({onlySelf: true});
    } else {
      this.form.enable({onlySelf: true});
    }
  }

  private notifChange(_: any) {}

  registerOnChange(fn: (_: any) => void): void {
    this.notifChange = fn;
  }

  // TODO: Use this function
  private notifTouch: any;

  registerOnTouched(fn: any): void {
    this.notifTouch = fn;
  }
}
