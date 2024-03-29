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
import { ReactiveFormsModule, FormBuilder, FormsModule, FormGroup } from '@angular/forms';

import { RetypePasswordComponent } from './retype-password.component';

import { FormWidgetModule } from 'src/app/form-widget/form-widget.module';

describe('RetypePasswordComponent', () => {
  let component: RetypePasswordComponent;
  let fixture: ComponentFixture<RetypePasswordComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ RetypePasswordComponent ],
      imports: [
        FormsModule,
        FormWidgetModule,
        ReactiveFormsModule,
      ],
      providers: [
        FormBuilder,
      ],
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(RetypePasswordComponent);
    component = fixture.componentInstance;

    component.form = new FormGroup({})

    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
