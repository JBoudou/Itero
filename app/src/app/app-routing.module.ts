import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';

import { DeadRouteComponent } from './dead-route/dead-route.component';
import { HomeComponent }      from './home/home.component';
import { ListComponent }      from './list/list.component';
import { LoginComponent }     from './login/login.component';
import { PollComponent }      from './poll/poll.component';
import { SignupComponent }    from './signup/signup.component';

const routes: Routes = [
  { path: '', component: HomeComponent },
  { path: 'r/list', component: ListComponent },
  { path: 'r/login', component: LoginComponent },
  { path: 'r/signup', component: SignupComponent },
  { path: 'r/poll/:pollSegment', component: PollComponent },
  { path: '**', component: DeadRouteComponent },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
