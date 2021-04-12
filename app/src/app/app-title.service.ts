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
import { Title } from '@angular/platform-browser';
import { Router, NavigationEnd } from '@angular/router';

import { filter } from 'rxjs/operators';

/**
 * Service responsible of the title of the pages.
 *
 * The title can be set manually by calling setTitle, or automatically by
 * giving a title entry to the data field in the config of the router.
 */
@Injectable({
  providedIn: 'root'
})
export class AppTitleService {

  private _baseTitle: string;

  constructor(
    private router: Router,
    private title: Title,
  ) {
    this._baseTitle = this.title.getTitle();
    this.router.events.pipe(
      filter(event => event instanceof NavigationEnd)
    ).subscribe(
      () => this.updateTitle()
    );
  }

  getTitle(): string {
    return this.title.getTitle();
  }

  setTitle(title: string|string[]|undefined) {
    var array: string[] = title === undefined ? [] :
                          Array.isArray(title) ? title : [title];
    array.push(this._baseTitle);
    this.title.setTitle(array.join(' - '));
  }

  private updateTitle() {
    var current = this.router.routerState.snapshot.root;
    var title: string|undefined;
    while (true) {
      title = current.data?.title;
      if (title !== undefined) {
        break;
      }
      if (current.children.length != 1) {
        break;
      }
      current = current.firstChild;
    }
    this.setTitle(title);
  }
}
