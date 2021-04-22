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

import { BehaviorSubject, Observable, Subscription } from 'rxjs';
import {PollNotifAction, PollNotifAnswerEntry} from './api';
import {PollNotifService} from './poll-notif.service';

class Notif {
  constructor(
    public message: string,
    public tag: string,
  ) { }

  send(): void {
    new Notification('Itero', { body: this.message, tag: this.tag });
  }
}

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
        next: (notif: PollNotifAnswerEntry) => this.sendPollNotif(notif),
      }),
    );
  }

  private endNotifications(): void {
    this._subscriptions.forEach((sub: Subscription) => sub.unsubscribe());
    this._subscriptions = [];
  }

  private sendPollNotif(notif: PollNotifAnswerEntry): void {
    let msg: string;
    switch (notif.Action) {
    case PollNotifAction.Next:
      msg = `New round started for poll ${notif.Segment}.`;
      break;
    case PollNotifAction.Term:
      msg = `Poll ${notif.Segment} has been closed.`;
      break;
    case PollNotifAction.Start:
      msg = `Poll ${notif.Segment} has been started.`;
      break;
    case PollNotifAction.Delete:
      msg = `Poll ${notif.Segment} has been deleted.`;
      break;
    }
    (new Notif(msg, notif.Segment)).send();
  }
  
}
