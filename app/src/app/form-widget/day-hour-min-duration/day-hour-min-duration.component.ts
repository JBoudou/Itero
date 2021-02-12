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

import { Component, OnInit, Input, Self, Optional, OnDestroy } from '@angular/core';
import { ControlValueAccessor, NgControl, FormBuilder } from '@angular/forms';

import { Subscription } from 'rxjs';

import { DetailedDuration } from '../duration.pipe';

@Component({
  selector: 'app-day-hour-min-duration',
  templateUrl: './day-hour-min-duration.component.html',
  styleUrls: ['./day-hour-min-duration.component.sass']
})
export class DayHourMinDurationComponent implements OnInit, OnDestroy, ControlValueAccessor {

  @Input() disabled: boolean;
  
  form = this.formBuilder.group({
    days:  2,
    hours: 0,
    mins:  0,
  });

  private _subscriptions: Subscription[] = [];

  constructor(
    @Self() @Optional() private ngControl: NgControl,
    private formBuilder: FormBuilder,
  ) {
    if (this.ngControl) {
      this.ngControl.valueAccessor = this;
    }
  }

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
    const { days, hours, mins } = this.form.value;
    if (Number.isInteger(days) && Number.isInteger(hours) && Number.isInteger(mins)) {
      this.notifChange(new DetailedDuration(days, hours, mins).toMilliseconds());
    } else {
      this.notifChange(null);
    }
  }

  /** Implements ControlValueAccessor */

  writeValue(obj: any): void {
    if (!Number.isInteger(obj)) {
      console.warn("DateTimePickerComponent unknown value type " + typeof(obj));
      return;
    }
    
    this.form.patchValue(DetailedDuration.fromMilliseconds(obj));
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
