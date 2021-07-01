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
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormsModule } from '@angular/forms';

import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';

import { DateTimePickerComponent } from './date-time-picker/date-time-picker.component';
import { DayHourMinDurationComponent } from './day-hour-min-duration/day-hour-min-duration.component';
import { DurationPipe } from './duration.pipe';
import { DisclosePasswordComponent } from './disclose-password/disclose-password.component';


@NgModule({
  declarations: [
    DateTimePickerComponent,
    DayHourMinDurationComponent,
    DurationPipe,
    DisclosePasswordComponent,
  ],
  imports: [
    CommonModule,
    FormsModule,
    MatButtonModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    ReactiveFormsModule,
  ],
  exports: [
    DateTimePickerComponent,
    DayHourMinDurationComponent,
    DurationPipe,
    DisclosePasswordComponent,
  ],
})
export class FormWidgetModule { }
