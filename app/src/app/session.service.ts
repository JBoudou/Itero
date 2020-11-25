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
import { HttpClient } from '@angular/common/http';
import { Observable, Subject } from 'rxjs';
import { map } from 'rxjs/operators';

import { LoginInfo } from './api'

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

  constructor(private http: HttpClient) {
  }

  /** Whether a session is currently registered. */
  registered(): boolean {
    return this.currentState.registered;
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
   * Attempt to log in with the given user.
   *
   * The returned Observable sends the function parameters on success, or
   * the error from HttpClient on failure.
   */
  login(info: LoginInfo): Observable<LoginInfo> {
    return this.http.post('/a/login', info).pipe(
      map((data: string) => {
        this.register(data, info.User);
        localStorage.setItem("SessionId", this.sessionId);
        localStorage.setItem("User", info.User);
        return info;
      })
    );
  }

  /** Close the current session (if any). */
  logoff() {
    localStorage.removeItem("SessionId");
    localStorage.removeItem("User");
    document.cookie = "s=; Path=/; Max-Age=-1";
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
