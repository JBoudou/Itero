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


/**
 * Suspender allows to delay functions.
 * This class can be used to throttle, but it's more flexible.
 */
export class Suspender {

  /**
   * Constructor.
   * The action, if given, will be the default action used when do() is called without any argument.
   * If a delay is given, then suspend() will be called with that delay after each action is executed.
   */
  constructor(delay?: number);
  constructor(action: () => void, delay?: number);
  constructor(arg1?: number|(() => void), arg2?: number) {
    if (typeof arg1 === 'function') {
      this._action = arg1;
      this._delay = arg2;
    } else {
      this._delay = arg1;
    }
  }

  /**
   * Prevent any action to be executed for some time.
   *
   * The delay is either the number of miliseconds to wait or the date until actions can be run.
   * In the first case the timeout starts right now, when suspend() is called.
   *
   * If an action was already delayed while suspend() is called, it will be further delayed.
   * However, to give a date in the past to suspend() does not shorten the delay of the currently
   * waiting task.
   */
  suspend(delay: number|Date): void {
    let candidate: number;
    if (delay instanceof Date) {
      candidate = delay.getTime();
    } else {
      candidate = Date.now() + delay;
    }
    if (candidate > this._suspend) {
      this._suspend = candidate;
    }
  }

  /**
   * Ask for an action to be executed.
   * If no action is given, the default action given to the constructor is used.
   * If another action is already pending, the current action will not be executed.
   * If no action is already pending but actions are suspended, the action will be delayed.
   * Otherwise, the action is executed right now.
   * The function returns true only is the action is executed right now.
   */
  do(action?: () => void): boolean {
    if (this._timeout !== undefined) {
      return false;
    }
    return this._do(action ?? this._action);
  }

  /** Cancel any pending action. */
  cancel(): void {
    if (this._timeout) {
      clearTimeout(this._timeout);
      this._timeout = undefined;
    }
  }

  // Implementation //

  private _action: (() => void)|undefined;
  private _delay: number|undefined;
  private _suspend: number = 0;
  private _timeout: number|undefined;

  private _do(action: () => void): boolean {
    const now = Date.now();

    if (this._suspend <= now) {
      action();
      this._timeout = undefined;
      if (this._delay !== undefined) {
        this.suspend(this._delay);
      }
      return true;
    }

    this._timeout = window.setTimeout(() => this._do(action), this._suspend - now);
    return false;
  }
}

/**
 * Allows to have suspendable functions without explicit Suspender object.
 *
 * An implicit Suspender object is created with the given function and the given delay.
 * The returned value is equivalent to both the do() method of the implicit object and the implicit
 * object itself.
 *
 * This function is a workaround for the impossibility to have a TypeScript decorator, because
 * method decorators cannot change the type of methods.
 *
 * Beware that when used to have suspendable "method" the resulting value is enumerable and
 * writable.
 */
export function Suspendable(fct: () => void, delay?: number): {
  (): boolean;
  suspend(delay: number|Date): void;
  cancel(): void;
} {
  const suspender = new Suspender(delay);
  const ret = function (): boolean { return suspender.do(fct.bind(this)); };
  ret.suspend = suspender.suspend.bind(suspender);
  ret.cancel  = suspender. cancel.bind(suspender);
  return ret;
}
