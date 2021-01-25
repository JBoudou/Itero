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
import { Component } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

import { ClipboardModule } from '@angular/cdk/clipboard';
import { MatIconModule } from '@angular/material/icon';

import { ActivatedRouteStub } from '../../../testing/activated-route-stub'

import { ResultComponent } from './result.component';

import { CreateService } from '../create.service';

@Component({ selector: 'app-poll', template: '' })
class PollStubComponent { }

describe('ResultComponent', () => {
  let component: ResultComponent;
  let fixture: ComponentFixture<ResultComponent>;
  let activatedRouteStub: ActivatedRouteStub;
  let serviceSpy: jasmine.SpyObj<CreateService>;

  beforeEach(async () => {
    activatedRouteStub = new ActivatedRouteStub();
    serviceSpy = jasmine.createSpyObj('CreateService', [], ['httpError']);

    await TestBed.configureTestingModule({
      declarations: [
        ResultComponent,
        PollStubComponent,
      ],
      imports: [ ClipboardModule, MatIconModule ],
      providers: [
        { provide: ActivatedRoute, useValue: activatedRouteStub },
        { provide: CreateService, useValue: serviceSpy },
      ],
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ResultComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
