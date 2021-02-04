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

import { Component, Injectable } from '@angular/core';

import {
  ActivatedRouteSnapshot,
  CanActivate,
  CanDeactivate,
  Router,
  RouterStateSnapshot,
  UrlTree,
} from '@angular/router';

import { Observable } from 'rxjs';
import { map }        from 'rxjs/operators';
import { MatDialog }  from '@angular/material/dialog';

import { CreateComponent }  from './create.component';
import { CreateService }    from './create.service';

/** Ask the user what to do when leaving create routes. */
@Injectable()
export class CreateGuard implements CanActivate, CanDeactivate<CreateComponent> {

  constructor(
    private dialog: MatDialog,
    private service: CreateService,
    private router: Router,
  ) { }

  canActivate(
    route: ActivatedRouteSnapshot,
    state: RouterStateSnapshot,
  ): Observable<boolean | UrlTree> | Promise<boolean | UrlTree> | boolean | UrlTree {
    const serviceUrl = this.service.currentUrl();
    if (state.url === serviceUrl) {
      return true;
    }
    return this.router.parseUrl(serviceUrl);
  }

  canDeactivate(
    component: CreateComponent,
    currentRoute: ActivatedRouteSnapshot,
    currentState: RouterStateSnapshot,
    nextState?: RouterStateSnapshot
  ): Observable<boolean|UrlTree>|Promise<boolean|UrlTree>|boolean|UrlTree {

    if (this.service.canLeave()) {
      return true;
    }

    const ref = this.dialog.open(LeaveCreateDialog, {
      disableClose: true,
    });
    return ref.afterClosed().pipe(map((result: string) => {
      switch (result) {
      case 'Reset':
        this.service.reset();
        return true;
      case 'Keep':
        return true;
      default:
        return false;
      }
    }));
  }
}

@Component({
  selector: 'leave-create-dialog',
  templateUrl: 'leave-create.dialog.html',
})
export class LeaveCreateDialog {}
