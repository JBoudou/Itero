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

import { LoggedGuard } from '../logged.guard';

import { CreateComponent }              from './create.component';
import { GeneralComponent }             from './general/general.component';
import { SimpleAlternativesComponent }  from './simple-alternatives/simple-alternatives.component';
import { RoundComponent }               from './round/round.component';
import { ResultComponent }              from './result/result.component';

const routes: Routes = [
  {
    path: '',
    component: CreateComponent,
    canActivate: [LoggedGuard],
    children: [
    { path: 'general', component: GeneralComponent },
    { path: 'simpleAlternatives', component: SimpleAlternativesComponent },
    { path: 'round', component: RoundComponent },
    { path: '', redirectTo: 'general', pathMatch: 'full' },
  ] },
  { path: 'result/:pollSegment', component: ResultComponent },
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class CreateRoutingModule { }
