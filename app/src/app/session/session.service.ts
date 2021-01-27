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
import { HttpClient, HttpErrorResponse } from '@angular/common/http';

import { Observable, BehaviorSubject, pipe } from 'rxjs';
import { map, tap } from 'rxjs/operators';

import { SessionAnswer } from '../api';

export class SessionInfo {
  readonly logged: boolean;
  readonly user?: string;
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
  _state = new BehaviorSubject<SessionInfo>({logged: false})

  get state$(): Observable<SessionInfo> {
    return this._state;
  }

  constructor(
    private router: Router,
    private http: HttpClient,
  ) {
    this.checkSession();
  }

  /** Whether a session is currently registered. */
  get logged(): boolean {
    return this._state.value.logged;
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
   * Intercept session information.
   *
   * This operator must be piped in requests like login or signup that may
   * result in a new session.
   */
  httpOperator(user: string) {
    return pipe(
      map((data: SessionAnswer) => {
        this.register(data.SessionId, user);
        localStorage.setItem("SessionId", this.sessionId);
        localStorage.setItem("User", user);
        data.Expires = new Date(data.Expires);
        setTimeout(() => { this.refresh(); }, (data.Expires.getTime() - Date.now()) * 0.75);
        return true;
      }),
      tap({
        error: (err: HttpErrorResponse) => {
          console.log(`Intercepted error ${err.status}: ${err.message}`);
          this.logoff();
        }
      }),
    );
  }

  /** Close the current session (if any). */
  logoff() {
    localStorage.removeItem("SessionId");
    localStorage.removeItem("User");
    document.cookie = "s=; Path=/; Max-Age=0; Secure";
    this._state.next({logged: false});
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
  private checkSession() {
    if (this.logged) {
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
      setTimeout(() => { this.refresh(); }, 15 * 1000);
    }
  }

  private register(sessionId: string, user: string): void {
    this.sessionId = sessionId;
    this._state.next({logged: true, user: user});
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

  private refresh(): void {
    if (!this.logged) {
      return;
    }

    const user = this._state.value.user;
    this.http
      .post('/a/refresh', user)
      .pipe(this.httpOperator(user))
      .subscribe();
  }
}
