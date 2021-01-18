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
    DateTimePickerComponent,
    DayHourMinDurationComponent,
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
