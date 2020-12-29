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


import { Component, OnInit, Type, ViewChild, ViewContainerRef, ComponentFactoryResolver, ViewEncapsulation } from '@angular/core';
import { ActivatedRoute, ParamMap } from '@angular/router';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';

import { Subscription } from 'rxjs';

import { PollAnswer, BallotType, InformationType } from '../api';
import { PollBallotDirective, PollInformationDirective } from './directives';
import { PollSubComponent, ServerError } from './common';
import { UninominalBallotComponent } from '../uninominal-ballot/uninominal-ballot.component';
import { CountsInformationComponent } from '../counts-information/counts-information.component';

@Component({
  selector: 'app-poll',
  templateUrl: './poll.component.html',
  styleUrls: ['./poll.component.sass'],
  encapsulation: ViewEncapsulation.None,
})
export class PollComponent implements OnInit {

  segment: string;
  answer: PollAnswer;

  error: ServerError;
  errorSubscription: Subscription[] = [undefined, undefined];

  @ViewChild(PollBallotDirective, { static: true }) ballot: PollBallotDirective;
  @ViewChild(PollInformationDirective, { static: true }) information: PollInformationDirective;

  constructor(
    private route: ActivatedRoute,
    private http: HttpClient,
    private componentFactoryResolver: ComponentFactoryResolver,
  ) { }

  ngOnInit(): void {
    this.route.paramMap.subscribe((params: ParamMap) => {
      this.segment = params.get('pollSegment');
      this.retrieveTypes();
    });
  }

  hasAnswer(): boolean {
    return typeof this.error == 'undefined' &&
           typeof this.answer !== 'undefined';
  }

  noInformation(): boolean {
    return this.hasAnswer() &&
      this.answer.CurrentRound > 0 &&
      !PollComponent.informationMap.has(this.answer.Information);
  }

  private static ballotMap = new Map<BallotType, Type<any>>([
    [BallotType.Uninomial, UninominalBallotComponent]
  ]);

  private static informationMap = new Map<InformationType, Type<any>>([
    [InformationType.Counts, CountsInformationComponent]
  ]);

  private retrieveTypes(): void {
    this.http.get<PollAnswer>('/a/poll/' + this.segment).subscribe({
      next: (answer: PollAnswer) => {
        this.answer = answer;
        if (PollComponent.ballotMap.has(this.answer.Ballot)) {
          this.loadSubComponent(0, this.ballot.viewContainerRef,
                                PollComponent.ballotMap.get(this.answer.Ballot));
        }
        if (PollComponent.informationMap.has(this.answer.Information)) {
          this.loadSubComponent(1, this.information.viewContainerRef,
                                PollComponent.informationMap.get(this.answer.Information));
        }
      },
      error: (err: HttpErrorResponse) => {
        this.registerError({status: err.status, message: err.error.trim()});
      }
    });
  }

  private loadSubComponent(subcriptionIndex: number,
                           viewContainerRef: ViewContainerRef,
                           type: Type<any>): void {
    if (!!this.errorSubscription[subcriptionIndex]) {
      this.errorSubscription[subcriptionIndex].unsubscribe();
    }
    const componentFactory = this.componentFactoryResolver.resolveComponentFactory(type);
    viewContainerRef.clear();
    const componentRef = viewContainerRef.createComponent<PollSubComponent>(componentFactory);
    componentRef.instance.pollSegment = this.segment;
    this.errorSubscription[subcriptionIndex] = componentRef.instance.errors.subscribe({
      next: (err: ServerError) => {
        this.registerError(err);
      }
    });
  }

  private registerError(err: ServerError) {
    this.error = err;
    for (let i = 0; i < 2; i++) {
      this.errorSubscription[i] = undefined;
    }
    this.ballot.viewContainerRef.clear();
    this.information.viewContainerRef.clear();
  }

}
