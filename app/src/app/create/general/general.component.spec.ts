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
import { ActivatedRoute } from '@angular/router';

import { MatSelectModule }  from '@angular/material/select'; 

import { GeneralComponent } from './general.component';

import { CreateService } from '../create.service';

import { ActivatedRouteStub } from '../../../testing/activated-route-stub'

describe('GeneralComponent', () => {
  let component: GeneralComponent;
  let fixture: ComponentFixture<GeneralComponent>;
  let serviceSpy: jasmine.SpyObj<CreateService>;
  let activatedRouteStub: ActivatedRouteStub;

  beforeEach(async () => {
    serviceSpy = jasmine.createSpyObj('CreateService', {register: {}});
    activatedRouteStub = new ActivatedRouteStub();
    
    await TestBed.configureTestingModule({
      declarations: [ GeneralComponent ],
      imports: [ ReactiveFormsModule, FormsModule, MatSelectModule ],
      providers: [
        FormBuilder,
        { provide: CreateService, useValue: serviceSpy },
        { provide: ActivatedRoute, useValue: activatedRouteStub },
      ],
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(GeneralComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});