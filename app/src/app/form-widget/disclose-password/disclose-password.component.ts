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

import { ChangeDetectionStrategy, Component, ElementRef, HostBinding, Input, OnDestroy, OnInit, Optional, Self } from '@angular/core';
import { ControlValueAccessor, FormControl, NgControl } from '@angular/forms';

import { Observable, Subject, Subscription } from 'rxjs';
import { MatFormFieldControl } from '@angular/material/form-field';
import { FocusMonitor, FocusOrigin } from '@angular/cdk/a11y';

@Component({
  selector: 'app-disclose-password',
  templateUrl: './disclose-password.component.html',
  styleUrls: ['./disclose-password.component.sass'],
  host: { role: 'group', class: 'merged-input' },
  providers: [{provide: MatFormFieldControl, useExisting: DisclosePasswordComponent}],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class DisclosePasswordComponent
    implements OnInit, OnDestroy, ControlValueAccessor, MatFormFieldControl<string> {

  control = new FormControl('')
  disclosed: boolean = false

  private _subscriptions: Subscription[] = []

  constructor(
    @Self() @Optional() public ngControl: NgControl,
    private focusMonitor: FocusMonitor,
    private hostElementRef: ElementRef<HTMLElement>,
  ) {
    if (this.ngControl) {
      this.ngControl.valueAccessor = this
    }

    this._subscriptions.push(focusMonitor.monitor(hostElementRef, true).subscribe({
      next: (origin: FocusOrigin) => {
        this.focused = !!origin
        this._stateChanges.next()
      }
    }));
  }

  ngOnInit(): void {
    this._subscriptions.push(this.control.valueChanges.subscribe({ next: _ => this.onValueChange() }))
    if (this.disabled) {
      this.control.disable({onlySelf: true})
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

  get value(): string {
    return this.control.value
  }
  set value(newVal: string) {
    this.control.setValue(newVal)
  }

  private _stateChanges = new Subject<void>();
  get stateChanges(): Observable<void> {
    return this._stateChanges;
  }

  static _nextId: number = 0;
  @HostBinding() id = `app-disclose-password-${DisclosePasswordComponent._nextId++}`;

  private _placeholder: string;
  @Input()
  get placeholder(): string {
    return this._placeholder;
  }
  set placeholder(placeholder: string) {
    this._placeholder = placeholder;
    this._stateChanges.next();
  }

  focused: boolean = false;

  get empty(): boolean {
    return typeof this.control.value !== 'string' || this.control.value === ''
  }

  @HostBinding('class.floating')
  get shouldLabelFloat(): boolean {
    return this.focused || !this.empty;
  }

  get required(): boolean {
    // TODO: To have a better behavior, there must be a way to display and set the undefined value.
    return false;
  }

  @Input()
  get disabled(): boolean {
    return !this.control.enabled;
  }
  set disabled(val: boolean) {
    if (val) {
      this.control.disable();
    } else {
      this.control.enable();
    }
    this._stateChanges.next();
  }

  get errorState(): boolean {
    return this.control.invalid;
  }

  get controlType(): string {
    return 'disclose-password';
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
    if (typeof obj !== 'string') {
      console.warn("DisclosePasswordComponent unknown value type " + typeof(obj));
      return;
    }
    this.value = obj;
  }

  setDisabledState(isDisabled: boolean): void {
    if (isDisabled) {
      this.control.disable({onlySelf: true});
    } else {
      this.control.enable({onlySelf: true});
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
