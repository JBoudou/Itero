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
import { ActivatedRouteSnapshot, RouterStateSnapshot, UrlTree, CanDeactivate } from '@angular/router';

import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { MatDialog } from '@angular/material/dialog';

import { CreateComponent } from './create.component';
import { CreateService } from './create.service';

@Injectable()
export class LeaveCreateGuard implements CanDeactivate<CreateComponent> {

  constructor(
    private dialog: MatDialog,
    private service: CreateService,
  ) { }

  canDeactivate(
    component: CreateComponent,
    currentRoute: ActivatedRouteSnapshot,
    currentState: RouterStateSnapshot,
    nextState?: RouterStateSnapshot
  ): Observable<boolean|UrlTree>|Promise<boolean|UrlTree>|boolean|UrlTree {

    if (!this.service.isStarted()) {
      return true;
    }

    const ref = this.dialog.open(LeaveCreateDialog, {
//      width: '300px',
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
