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

import { ConfirmationComponent } from './confirmation/confirmation.component';
import { LoginComponent }   from './login/login.component';
import { SessionGuard } from './session.guard';
import { SignupComponent }  from './signup/signup.component';

const routes: Routes = [
  { path: 'r/session', canDeactivate: [ SessionGuard ], children: [
    { path: 'login' , component: LoginComponent , data: {title: 'Log in' } },
    { path: 'signup', component: SignupComponent, data: {title: 'Sign up'} },
  ]},
  { path: 'r/confirm/:confirmSegment', component: ConfirmationComponent },
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class SessionRoutingModule { }
