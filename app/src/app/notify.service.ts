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

import { Injectable } from '@angular/core';
import {Router} from '@angular/router';

import { BehaviorSubject, Observable, Subscription } from 'rxjs';
import { PollNotifAction } from './api';
import { PollNotifService, PollNotification } from './poll-notif.service';

@Injectable({
  providedIn: 'root'
})
export class NotifyService {

  private _permission = new BehaviorSubject<NotificationPermission>("default");
  get permission$(): Observable<NotificationPermission> {
    return this._permission;
  }

  private _subscriptions: Subscription[] = [];

  constructor(
    private pollNotifs: PollNotifService,
    private router: Router,
  ) {
    this.handlePermission(Notification.permission);
  }

  askAUthorization(): void {
    Notification.requestPermission().then((perm: NotificationPermission) => this.handlePermission(perm));
  }

  private handlePermission(perm: NotificationPermission): void {
    if (perm === 'granted') {
      this.startNotifications();
    } else {
      this.endNotifications();
    }
    this._permission.next(perm);
  }

  private startNotifications(): void {
    this.endNotifications();

    this._subscriptions.push(
      this.pollNotifs.event$.subscribe({
        next: (notif: PollNotification) => this.sendPollNotif(notif),
      }),
    );
  }

  private endNotifications(): void {
    this._subscriptions.forEach((sub: Subscription) => sub.unsubscribe());
    this._subscriptions = [];
  }

  private sendPollNotif(notif: PollNotification): void {
    let msg: string;
    switch (notif.Action) {
    case PollNotifAction.Next:
      msg = `Round ${notif.Round} has started for poll "${notif.Title}".`;
      break;
    case PollNotifAction.Term:
      msg = `Poll "${notif.Title}" has been closed.`;
      break;
    case PollNotifAction.Start:
      msg = `Poll "${notif.Title}" has been started.`;
      break;
    case PollNotifAction.Delete:
      msg = `Poll "${notif.Title}" has been deleted.`;
      break;
    }
    this.send(msg, notif.Segment, notif.Segment);
  }

  private send(message: string, tag: string, segment?: string): void {
    const notif = new Notification('Itero', { body: message, tag: tag });
    if (segment !== undefined && segment !== '') {
      notif.addEventListener<'click'>('click', () => {
        this.router.navigate(['/r/poll', segment]);
      });
    }
  }
  
}
