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

import { LoggedGuard } from './logged.guard';

import { CreateComponent }                  from './create/create.component';
import { CreateGeneralComponent }           from './create-general/create-general.component';
import { CreateSimpleAlternativesComponent }from './create-simple-alternatives/create-simple-alternatives.component';
import { CreateRoundComponent }             from './create-round/create-round.component';
import { CreateResultComponent }            from './create-result/create-result.component';
import { DeadRouteComponent }               from './dead-route/dead-route.component';
import { HomeComponent }                    from './home/home.component';
import { LoginComponent }                   from './login/login.component';
import { SignupComponent }                  from './signup/signup.component';

const routes: Routes = [
  { path: '', component: HomeComponent },
  { path: 'r/create', component: CreateComponent, canActivate: [LoggedGuard], children: [
    { path: 'general', component: CreateGeneralComponent },
    { path: 'simpleAlternatives', component: CreateSimpleAlternativesComponent },
    { path: 'round', component: CreateRoundComponent },
    { path: '', redirectTo: 'general', pathMatch: 'full' },
  ] },
  { path: 'r/create-result/:pollSegment', component: CreateResultComponent },
  { path: 'r/login', component: LoginComponent },
  { path: 'r/signup', component: SignupComponent },
  { path: '**', component: DeadRouteComponent },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
