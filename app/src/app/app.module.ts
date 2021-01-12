import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { HttpClientModule, HTTP_INTERCEPTORS } from '@angular/common/http';
import { ReactiveFormsModule, FormsModule } from '@angular/forms';

import { GoogleChartsModule } from 'angular-google-charts';
import { DragDropModule } from '@angular/cdk/drag-drop';

import { AppRoutingModule } from './app-routing.module';
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
import { DateTimePickerComponent } from './date-time-picker/date-time-picker.component';
import { DayHourMinDurationComponent } from './day-hour-min-duration/day-hour-min-duration.component';

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
    CountsInformationComponent,
    CreateComponent,
    CreateGeneralComponent,
    CreateSimpleAlternativesComponent,
    CreateRoundComponent,
    DateTimePickerComponent,
    DayHourMinDurationComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    HttpClientModule,
    ReactiveFormsModule,
    FormsModule,
    BrowserAnimationsModule,
    GoogleChartsModule,
    DragDropModule,
  ],
  providers: [
    { provide: HTTP_INTERCEPTORS, useClass: SessionInterceptor, multi: true },
    { provide: CREATE_TREE, useValue: APP_CREATE_TREE },
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
