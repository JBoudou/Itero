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

import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { HttpClientModule, HTTP_INTERCEPTORS } from '@angular/common/http';
import { ReactiveFormsModule, FormsModule } from '@angular/forms';

import { GoogleChartsModule } from 'angular-google-charts';
import { DragDropModule } from '@angular/cdk/drag-drop';
import { MatIconModule } from '@angular/material/icon'; 
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatButtonModule } from '@angular/material/button';
import { ClipboardModule } from '@angular/cdk/clipboard';

import { AppRoutingModule } from './app-routing.module';
import { FormWidgetModule } from './form-widget/form-widget.module';

import { SessionInterceptor } from './session/session.interceptor';
import { CREATE_TREE, APP_CREATE_TREE } from './create/create.service';

import { AppComponent } from './app.component';
import { NavtitleComponent } from './navtitle/navtitle.component';
import { ListComponent } from './list/list.component';
import { LoginComponent } from './login/login.component';
import { HomeComponent } from './home/home.component';
import { DeadRouteComponent } from './dead-route/dead-route.component';
import { SignupComponent } from './signup/signup.component';
import { PollComponent } from './poll/poll.component';
import { PollBallotDirective, PollInformationDirective } from './poll/directives';
import { UninominalBallotComponent } from './uninominal-ballot/uninominal-ballot.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { CountsInformationComponent } from './counts-information/counts-information.component';
import { CreateComponent } from './create/create.component';
import { CreateGeneralComponent } from './create-general/create-general.component';
import { CreateSimpleAlternativesComponent } from './create-simple-alternatives/create-simple-alternatives.component';
import { CreateRoundComponent } from './create-round/create-round.component';
import { CreateResultComponent } from './create-result/create-result.component';

@NgModule({
  declarations: [
    AppComponent,
    CountsInformationComponent,
    CreateComponent,
    CreateGeneralComponent,
    CreateResultComponent,
    CreateRoundComponent,
    CreateSimpleAlternativesComponent,
    DeadRouteComponent,
    HomeComponent,
    ListComponent,
    LoginComponent,
    NavtitleComponent,
    PollBallotDirective,
    PollComponent,
    PollInformationDirective,
    SignupComponent,
    UninominalBallotComponent,
  ],
  imports: [
    AppRoutingModule,
    BrowserAnimationsModule,
    BrowserModule,
    ClipboardModule,
    DragDropModule,
    FormsModule,
    FormWidgetModule,
    GoogleChartsModule,
    HttpClientModule,
    MatButtonModule,
    MatIconModule,
    MatTooltipModule,
    ReactiveFormsModule,
  ],
  providers: [
    { provide: HTTP_INTERCEPTORS, useClass: SessionInterceptor, multi: true },
    { provide: CREATE_TREE, useValue: APP_CREATE_TREE },
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
