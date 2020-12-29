import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { HttpClientModule, HTTP_INTERCEPTORS } from '@angular/common/http';
import { ReactiveFormsModule } from '@angular/forms';

import { MatListModule } from '@angular/material/list';
import { GoogleChartsModule } from 'angular-google-charts';

import { AppRoutingModule } from './app-routing.module';
import { SessionInterceptor } from './session/session.interceptor';

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

@NgModule({
  declarations: [
    AppComponent,
    NavtitleComponent,
    ListComponent,
    LoginComponent,
    HomeComponent,
    DeadRouteComponent,
    SignupComponent,
    PollComponent,
    PollBallotDirective,
    PollInformationDirective,
    UninominalBallotComponent,
    CountsInformationComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    HttpClientModule,
    ReactiveFormsModule,
    BrowserAnimationsModule,
    GoogleChartsModule,
    MatListModule
  ],
  providers: [
    { provide: HTTP_INTERCEPTORS, useClass: SessionInterceptor, multi: true }
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
