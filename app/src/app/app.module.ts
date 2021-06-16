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
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HttpClientModule } from '@angular/common/http';

import { MatIconModule }    from '@angular/material/icon';
import { MatMenuModule }    from '@angular/material/menu';

import { AppRoutingModule } from './app-routing.module';
import { ListModule }       from './list/list.module';
import { PollModule }       from './poll/poll.module';
import { SessionModule }    from './session/session.module';

import { AppComponent } from './app.component';
import { NavtitleComponent } from './navtitle/navtitle.component';
import { HomeComponent } from './home/home.component';
import { DeadRouteComponent } from './dead-route/dead-route.component';

/**
 * Root module.
 * Must be kept as small as possible.
 */
@NgModule({
  declarations: [
    AppComponent,
    DeadRouteComponent,
    HomeComponent,
    NavtitleComponent
  ],
  imports: [
    BrowserAnimationsModule,
    BrowserModule,
    HttpClientModule,
    ListModule,
    MatIconModule,
    MatMenuModule,
    PollModule,
    SessionModule,
    AppRoutingModule // Must be last
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
