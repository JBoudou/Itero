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

import {
  ChangeDetectionStrategy,
  Component,
  ElementRef,
  HostBinding,
  Input,
  OnDestroy,
  OnInit,
  Optional,
  Self,
} from '@angular/core';
import { ControlValueAccessor, NgControl, FormBuilder } from '@angular/forms';

import { Observable, Subject, Subscription } from 'rxjs';
import { MatFormFieldControl } from '@angular/material/form-field';
import { FocusMonitor, FocusOrigin } from '@angular/cdk/a11y';

import { DetailedDuration } from '../duration.pipe';

@Component({
  selector: 'app-day-hour-min-duration',
  templateUrl: './day-hour-min-duration.component.html',
  styleUrls: ['./day-hour-min-duration.component.sass'],
  providers: [{provide: MatFormFieldControl, useExisting: DayHourMinDurationComponent}],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class DayHourMinDurationComponent
          implements OnInit, OnDestroy, ControlValueAccessor, MatFormFieldControl<number> {

  @Input()
  get disabled(): boolean {
    return !this.form.enabled;
  }
  set disabled(val: boolean) {
    if (val) {
      this.form.disable();
    } else {
      this.form.enable();
    }
    this._stateChanges.next();
  }

  
  form = this.formBuilder.group({
    days:  2,
    hours: 0,
    mins:  0,
  });

  private _subscriptions: Subscription[] = [];

  constructor(
    @Self() @Optional() public ngControl: NgControl,
    private formBuilder: FormBuilder,
    private focusMonitor: FocusMonitor,
    private hostElementRef: ElementRef<HTMLElement>,
  ) {
    if (this.ngControl) {
      this.ngControl.valueAccessor = this;
    }

    this._subscriptions.push(focusMonitor.monitor(hostElementRef, true).subscribe({
      next: (origin: FocusOrigin) => {
        this.focused = !!origin;
        this._stateChanges.next();
      }
    }));
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
    this._stateChanges.complete();
    this.focusMonitor.stopMonitoring(this.hostElementRef);
  }

  onValueChange(): void {
    this.notifChange(this.value);
    this._stateChanges.next();
  }


  /** Implements MatFormFieldControl */

  get value(): number {
    const { days, hours, mins } = this.form.value;
    if (Number.isInteger(days) && Number.isInteger(hours) && Number.isInteger(mins)) {
      return new DetailedDuration(days, hours, mins).toMilliseconds();
    } else {
      return null;
    }
  }

  set value(newVal: number) {
    this.form.patchValue(DetailedDuration.fromMilliseconds(newVal));
  }

  private _stateChanges = new Subject<void>();
  get stateChanges(): Observable<void> {
    return this._stateChanges;
  }

  static _nextId: number = 0;
  @HostBinding() id = `app-day-hour-min-duration-${DayHourMinDurationComponent._nextId++}`;

  get placeholder(): string {
    return '';
  }

  focused: boolean = false;

  get empty(): boolean {
    const { days, hours, mins } = this.form.value;
    return !Number.isInteger(days) && !Number.isInteger(hours) && !Number.isInteger(mins);
  }

  @HostBinding('class.floating')
  get shouldLabelFloat(): boolean {
    return this.focused || !this.empty;
  }

  get required(): boolean {
    // TODO: To have a better behavior, there must be a way to display and set the undefined value.
    return false;
  }

  get errorState(): boolean {
    return this.form.invalid;
  }

  get controlType(): string {
    return 'day-hour-min-duration';
  }

  setDescribedByIds(ids: string[]): void {
    // TODO: Implements the whole accessibility stuff.
  }

  onContainerClick(event: MouseEvent): void {
    if ((event.target as Element).tagName.toLowerCase() != 'input') {
      this.hostElementRef.nativeElement.querySelector('input').focus();
    }
  }


  /** Implements ControlValueAccessor */

  writeValue(obj: any): void {
    if (!Number.isInteger(obj)) {
      console.warn("DateTimePickerComponent unknown value type " + typeof(obj));
      return;
    }
    this.value = obj;
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
