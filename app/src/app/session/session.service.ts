// Itero - Online iterative vote application
// Copyright (C) 2020 Joseph Boudou
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

import { Injectable } from '@angular/core';
import { Router } from '@angular/router';

import { Subject } from 'rxjs';
import { map } from 'rxjs/operators';

export class SessionInfo {
  registered: boolean;
  user?: string;
}

/**
 * Manage the session.
 *
 * TODO
 */
@Injectable({
  providedIn: 'root'
})
export class SessionService {

  /**
   * Current session identifier.
   * May be anything when registered() returns false.
   */
  sessionId: string = '';

  /** Observable to subscribe to to be notified about session change. */
  observable: Subject<SessionInfo> = new Subject<SessionInfo>()

  currentState: SessionInfo = {registered: false};

  constructor(
    private router: Router,
  ) { }

  /** Whether a session is currently registered. */
  registered(): boolean {
    return this.currentState.registered;
  }

  get loginUrl(): string {
    return '/r/login';
  }

  private _loginRedirectionUrl : string | undefined;

  /**
   * Get the URL to go after a successful login.
   *
   * The call to getLoginRedirectionUrl() consumes the URL set by setLoginRedirectionUrl().
   * If no URL has previously been set, or the URL that have been set has already been consumed,
   * a default URL is returned.
   */
  getLoginRedirectionUrl(): string {
    if (this._loginRedirectionUrl === undefined) {
      return '/r/list';
    }
    const ret = this._loginRedirectionUrl;
    this._loginRedirectionUrl = undefined;
    return ret;
  }

  /** Set the URL to be returned by the next call to getLoginRedirectionUrl(). */
  setLoginRedirectionUrl(url: string): void {
    if (this._loginRedirectionUrl !== undefined) {
      console.error('setLoginRedirectionUrl while there already is a redirection URL');
    }
    this._loginRedirectionUrl = url;
  }

  /** Ask the user to log before returning to the current page. */
  logNow(): void {
    this.setLoginRedirectionUrl(this.router.url);
    this.router.navigateByUrl(this.loginUrl);
  }

  /**
   * Check if a session is available in the browser.
   *
   * If there is one, it is used. Otherwise, the service gets unregistered.
   * Notice that the synchronization is done only in the direction from the
   * browser to the service.
   *
   * This method should be called once at startup.
   */
  checkSession() {
    if (this.registered()) {
      if (!this.hasCookie) {
        this.logoff();
      }
      return;
    }
    if (!this.hasCookie()) {
      return;
    }

    let sessionId = localStorage.getItem("SessionId");
    if (!!sessionId) {
      this.register(sessionId, localStorage.getItem("User"));
    }
  }

  /**
   * Intercept session information.
   *
   * This operator must be piped in requests like login or signup that may
   * result in a new session.
   */
  httpOperator(user: string) {
    return map((data: string) => {
      this.register(data, user);
      localStorage.setItem("SessionId", this.sessionId);
      localStorage.setItem("User", user);
      return true;
    })
  }

  /** Close the current session (if any). */
  logoff() {
    localStorage.removeItem("SessionId");
    localStorage.removeItem("User");
    document.cookie = "s=; Path=/; Max-Age=0; Secure";
    this.currentState = {registered: false};
    this.observable.next(this.currentState);
  }

  private register(sessionId: string, user: string): void {
    this.sessionId = sessionId;
    this.currentState = {registered: true, user: user};
    this.observable.next(this.currentState);
  }

  private hasCookie(): boolean {
    let ca: Array<string> = document.cookie.split(/; */);
    let caLen: number = ca.length;
    for (let i: number = 0; i < caLen; i+=1) {
      if (ca[i].startsWith('s=')) {
        return true;
      }
    }
    return false;
  }
}
