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
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatRadioModule } from '@angular/material/radio';

import { BehaviorSubject } from 'rxjs';

import { RoundComponent } from './round.component';

import { CreateService } from '../create.service';
import { CreateQuery } from 'src/app/api';
import { FormWidgetModule } from '../../form-widget/form-widget.module';

import { ActivatedRouteStub } from '../../../testing/activated-route-stub'

describe('RoundComponent', () => {
  let component: RoundComponent;
  let fixture: ComponentFixture<RoundComponent>;
  let activatedRouteStub: ActivatedRouteStub;
  let query$: BehaviorSubject<Partial<CreateQuery>>;
  let serviceSpy: jasmine.SpyObj<CreateService>;

  beforeEach(async () => {
    activatedRouteStub = new ActivatedRouteStub();
    query$ = new BehaviorSubject<Partial<CreateQuery>>({});
    serviceSpy = jasmine.createSpyObj('CreateService', {}, { query$: query$ });
    
    await TestBed.configureTestingModule({
      declarations: [
        RoundComponent,
      ],
      imports: [
        ReactiveFormsModule,
        FormsModule,
        FormWidgetModule,
        MatFormFieldModule,
        MatInputModule,
        MatRadioModule,
        NoopAnimationsModule,
      ],
      providers: [
        FormBuilder,
        { provide: CreateService, useValue: serviceSpy },
        { provide: ActivatedRoute, useValue: activatedRouteStub },
      ],
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(RoundComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
    jasmine.clock().install();
  });

  afterEach(() => {
    jasmine.clock().uninstall();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('display values from the service', async () => {
    const query = {
      Start:    new Date('2010-10-10T10:10:10Z'),
      Deadline: new Date('2011-11-11T11:11:11Z'),
      MinNbRounds: 3,
      MaxNbRounds: 4,
      MaxRoundDuration: 1234000,
      ReportVote: false
    };
    query$.next(query);

    jasmine.clock().tick(1);
    await fixture.whenStable();
    
    expect(component.form.value).toEqual(query);
  });

});
