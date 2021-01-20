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

import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ReactiveFormsModule, FormBuilder, FormsModule } from '@angular/forms';

import { CreateRoundComponent } from './create-round.component';

import { CreateService } from '../create/create.service';
import { DateTimePickerComponent } from '../date-time-picker/date-time-picker.component';
import { DayHourMinDurationComponent } from '../day-hour-min-duration/day-hour-min-duration.component';

describe('CreateRoundComponent', () => {
  let component: CreateRoundComponent;
  let fixture: ComponentFixture<CreateRoundComponent>;
  let serviceSpy: jasmine.SpyObj<CreateService>;

  beforeEach(async () => {
    serviceSpy = jasmine.createSpyObj('CreateService', {register: {}});
    await TestBed.configureTestingModule({
      declarations: [
        CreateRoundComponent,
        DateTimePickerComponent,
        DayHourMinDurationComponent,
      ],
      imports: [ ReactiveFormsModule, FormsModule ],
      providers: [
        FormBuilder,
        { provide: CreateService, useValue: serviceSpy },
      ],
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(CreateRoundComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
