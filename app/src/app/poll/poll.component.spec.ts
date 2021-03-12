// Itero - Online iterative vote application
// Copyright (C) 2020 Joseph Boudou
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
import { ActivatedRoute } from '@angular/router';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { ReactiveFormsModule, FormBuilder, FormsModule } from '@angular/forms';
import { EMPTY } from 'rxjs';

import { ActivatedRouteStub } from '../../testing/activated-route-stub'
import { DynamicComponentFactoryStub } from '../../testing/dynamic-component-factory-stub'

import { PollBallotDirective, PollInformationDirective, PollPreviousDirective, PollComponent } from './poll.component';

import { UninominalBallotComponent } from './uninominal-ballot/uninominal-ballot.component';
import { CountsInformationComponent } from './counts-information/counts-information.component';
import { DynamicComponentFactoryService } from '../dynamic-component-factory.service';
import { BallotType, InformationType } from '../api';
import { SessionService } from '../session/session.service';

describe('PollComponent', () => {
  let component: PollComponent;
  let fixture: ComponentFixture<PollComponent>;
  let httpControler: HttpTestingController;
  let activatedRouteStub: ActivatedRouteStub;
  let dynamicFactoryStub: DynamicComponentFactoryStub;
  let sessionSpy: jasmine.SpyObj<SessionService>;

  beforeEach(async () => {
    activatedRouteStub = new ActivatedRouteStub();
    dynamicFactoryStub = new DynamicComponentFactoryStub();
    sessionSpy = jasmine.createSpyObj('SessionService', ['logNow'], ['logged']);

    await TestBed.configureTestingModule({
      declarations: [
        PollComponent,
        PollBallotDirective,
        PollInformationDirective,
        PollPreviousDirective,
      ],
      imports: [
        FormsModule,
        HttpClientTestingModule,
        ReactiveFormsModule,
      ],
      providers: [
        FormBuilder,
        { provide: ActivatedRoute, useValue: activatedRouteStub },
        { provide: DynamicComponentFactoryService, useValue: dynamicFactoryStub },
        { provide: SessionService, useValue: sessionSpy },
      ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    activatedRouteStub.setParamMap({pollSegment: '123456789'});
    dynamicFactoryStub.reset();

    fixture = TestBed.createComponent(PollComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    httpControler = TestBed.inject(HttpTestingController);
    jasmine.clock().install();
  });

  afterEach(function() {
    jasmine.clock().uninstall();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('asks middleware about poll type and create sub components', () => {
    const dummySpy = jasmine.createSpyObj('SubComponent', {
    },{
      pollSegment: '',
      finalResult: false,
      errors: EMPTY,
      previousRoundBallot: EMPTY,
      currentRoundBallot: EMPTY,
      justVoteBallot: EMPTY,
    })
    dynamicFactoryStub.nextComponent(UninominalBallotComponent, dummySpy);
    dynamicFactoryStub.nextComponent(CountsInformationComponent, dummySpy);

    const req = httpControler.expectOne('/a/poll/123456789');
    expect(req.request.method).toEqual('GET');
    req.flush({
      Title: 'Test',
      Description: '',
      Admin: 'author',
      CreationTime: new Date('2021-01-01T00:00:00'),
      CurrentRound: 1,
      Active: true,
      Ballot: BallotType.Uninominal,
      Information: InformationType.Counts,
    });
    jasmine.clock().tick(1);

    expect(dynamicFactoryStub.calls(UninominalBallotComponent)).toBe(1);
    expect(dynamicFactoryStub.calls(CountsInformationComponent)).toBe(1);
  })
});
