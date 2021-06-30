// Itero - Online iterative vote application
// Copyright (C) 2020 Joseph Boudou, David Gomez Prieto
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

import { Component, Inject, Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';

import { Observable, BehaviorSubject, pipe, UnaryFunction } from 'rxjs';
import { map, take, tap } from 'rxjs/operators';
import { MatDialog, MAT_DIALOG_DATA } from '@angular/material/dialog';

import { SessionAnswer } from '../api';
import { ScheduleOne } from '../shared/schedule-one';

export enum SessionState {
  Unlogged,
  Logged,
  Refreshing,
}

export class UserProfile {
  constructor(
    readonly name: string,
    readonly verified?: boolean
  ) {
    if (this.verified === undefined) {
      this.verified = false
    }
  }
}

export class SessionInfo {
  constructor(
    readonly state: SessionState,
    readonly profile?: UserProfile,
  ) { }

  get logged(): boolean {
    return this.state !== SessionState.Unlogged;
  }
  get user(): string {
    if (this.profile === undefined) {
      return undefined
    }
    return this.profile.name
  }
  get verified(): boolean {
    return this.logged && this.profile !== undefined && this.profile.verified
  }
}

const minRefreshTime = 15 * 1000;

/**
 * Manage the session.
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

  private _state = new BehaviorSubject<SessionInfo>(new SessionInfo(SessionState.Unlogged));

  /** Observable to be notified about session change. */
  get state$(): Observable<SessionInfo> {
    return this._state;
  }

  /** Whether a session is currently registered. */
  get logged(): boolean {
    return this._state.value.logged;
  }

  get loginUrl(): string {
    return '/r/session/login';
  }

  private _loginRedirectionUrl : string | undefined;

  private _scheduler = new ScheduleOne();

  constructor(
    private router: Router,
    private http: HttpClient,
    private dialog: MatDialog,
  ) {
    this.checkSession();
  }

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
    this.clearLoginRedirectionUrl();
    return ret;
  }

  /** Set the URL to be returned by the next call to getLoginRedirectionUrl(). */
  setLoginRedirectionUrl(url: string): void {
    if (this._loginRedirectionUrl !== undefined) {
      console.error('setLoginRedirectionUrl while there already is a redirection URL');
    }
    this._loginRedirectionUrl = url;
  }

  clearLoginRedirectionUrl(): void {
    this._loginRedirectionUrl = undefined;
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
  httpOperator(user: string): UnaryFunction<Observable<SessionAnswer>, Observable<boolean>> {
    return pipe(
      map((obj: any) => {
        const data = SessionAnswer.fromObject(obj);

        this.register(data.SessionId, new UserProfile(user, data.Verified));
        localStorage.setItem("SessionId", this.sessionId);
        localStorage.setItem("User", user);

        data.Expires = new Date(data.Expires);
        let diff = (data.Expires.getTime() - Date.now()) * 0.75;
        if (diff < minRefreshTime) {
          console.warn(`Too short refresh time: ${diff} sec.`);
          diff = minRefreshTime;
        }
        this._scheduler.schedule(() => { this.refresh(); }, (data.Expires.getTime() - Date.now()) * 0.75);

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

  verifyEmail(): void {
    this.http.get('/a/reverify')
      .pipe(take(1))
      .subscribe({
      next: () => {
        this.dialog.open<EmailVerificationDialog, boolean>(EmailVerificationDialog, {data: true});
      },
      error: (err: HttpErrorResponse) => {
        if (err.status == 409) {
          this.dialog.open<EmailVerificationDialog, boolean>(EmailVerificationDialog, {data: false});
        }
      }
    });
  }

  /** Close the current session (if any). */
  logoff() {
    localStorage.removeItem("SessionId");
    localStorage.removeItem("User");
    document.cookie = "s=; Path=/; Max-Age=0; Secure";
    document.cookie = "s=; Path=/; Max-Age=0; Secure; Domain=." + location.host;
    this._state.next(new SessionInfo(SessionState.Unlogged));
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
      this._scheduler.schedule(() => { this.refresh(); }, 1000);
    }
  }

  private register(sessionId: string, user: string|UserProfile): void {
    this.sessionId = sessionId;
    const profile: UserProfile = user instanceof UserProfile ? user : new UserProfile(user)
    this._state.next(new SessionInfo(SessionState.Logged, profile));
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
    this._state.next(new SessionInfo(SessionState.Refreshing));
    this.http
      .post('/a/refresh', user)
      .pipe(this.httpOperator(user), take(1))
      .subscribe();
  }
}

@Component({
  selector: 'email-verification-dialog',
  templateUrl: 'email-verification.dialog.html',
  host: {class: 'dialog-box'},
})
export class EmailVerificationDialog {
  constructor(
    @Inject(MAT_DIALOG_DATA) public emailSent: boolean,
  ) { }
}
