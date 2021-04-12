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

import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';

import { LoggedGuard } from './session/logged.guard';

import { DeadRouteComponent }               from './dead-route/dead-route.component';
import { HomeComponent }                    from './home/home.component';

const routes: Routes = [
  { path: '', component: HomeComponent },
  {
    path: 'r/create',
    canLoad: [ LoggedGuard ],
    loadChildren: () => import('./create/create.module').then(m => m.CreateModule),
    data: {title: "Create a new poll"},
  },
  { path: '**', component: DeadRouteComponent },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
