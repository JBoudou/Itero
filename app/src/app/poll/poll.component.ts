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


import {
  Component,
  OnInit,
  Type,
  ViewChild,
  ViewContainerRef,
  ViewEncapsulation
} from '@angular/core';

import { ActivatedRoute, ParamMap } from '@angular/router';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';

import { Subscription } from 'rxjs';

import {
  NONE_BALLOT,
  PollBallot,
  PollBallotComponent,
  PollInformationComponent,
  PollSubComponent,
  ServerError
} from './common';

import { PollAnswer, BallotType, InformationType } from '../api';
import { PollBallotDirective, PollInformationDirective } from './directives';
import { DynamicComponentFactoryService } from '../dynamic-component-factory.service';
import { SessionService } from '../session/session.service';

import { UninominalBallotComponent } from './uninominal-ballot/uninominal-ballot.component';
import { CountsInformationComponent } from './counts-information/counts-information.component';

// Indexes for the sub component.
const enum SubComponentId {
  Ballot = 0,
  Information
}

/**
 * This component displays the current state of the poll and allow the user to vote.
 * 
 * To manage the diversity of polls, this component delegates tasks to dynamic child components
 * called sub-component. There are currently two sub-components: one to display the ballot and
 * to vote, one to display the current informations about the poll.
 *
 * PollComponent ask the type of the poll to the middleware then creates the sub-components.
 */
@Component({
  selector: 'app-poll',
  templateUrl: './poll.component.html',
  styleUrls: ['./poll.component.sass'],
  encapsulation: ViewEncapsulation.None,
})
export class PollComponent implements OnInit {

  // Anchors to insert the dynamic sub-component into.
  @ViewChild(PollBallotDirective, { static: false }) ballot: PollBallotDirective;
  @ViewChild(PollInformationDirective, { static: false }) information: PollInformationDirective;

  segment: string;
  answer: PollAnswer;

  error: ServerError;

  previousRoundBallot: PollBallot = NONE_BALLOT;
  currentRoundBallot : PollBallot = NONE_BALLOT;
  justVoteBallot     : PollBallot = NONE_BALLOT;

  // Make it visible from template.
  // TODO: Implements a decorator for PollBallot that provides methods for that.
  BallotType = BallotType;

  /** Subscription for the sub component. The first index must be a SubComponentId. */
  private subscriptions: Subscription[][] = [];

  constructor(
    private route: ActivatedRoute,
    private http: HttpClient,
    private dynamicComponentFactory: DynamicComponentFactoryService,
    private session: SessionService,
  ) { }

  ngOnInit(): void {
    this.route.paramMap.subscribe((params: ParamMap) => {
      this.segment = params.get('pollSegment');
      this.retrieveTypes();
    });
  }

  /** Whether the response from the middleware has been received. */
  hasAnswer(): boolean {
    return typeof this.error == 'undefined' &&
           typeof this.answer !== 'undefined';
  }

  hasState(): boolean {
    return this.hasAnswer() &&
           ( this.answer.Information != InformationType.NoneYet ||
             this.hasCurrentRoundBallot() ||
             this.hasPreviousRoundBallot() ||
             this.hasJustVoteBallot() );
  }

  hasCurrentRoundBallot(): boolean {
    return this.currentRoundBallot.type != BallotType.None &&
           this.justVoteBallot.type == BallotType.None;
  }

  hasPreviousRoundBallot(): boolean {
    return this.previousRoundBallot.type != BallotType.None;
  }

  hasJustVoteBallot(): boolean {
    return this.justVoteBallot.type != BallotType.None;
  }

  private static ballotMap = new Map<BallotType, Type<PollBallotComponent>>([
    [BallotType.Uninominal, UninominalBallotComponent]
  ]);

  private static informationMap = new Map<InformationType, Type<PollSubComponent>>([
    [InformationType.Counts, CountsInformationComponent]
  ]);

  /** Ask the type of the poll to the middleware and creates the sub-components accordingly. */
  private retrieveTypes(): void {
    this.http.get<PollAnswer>('/a/poll/' + this.segment).subscribe({
      next: (answer: PollAnswer) => {

        this.answer = answer;

        // Update Ballot component
        if (this.answer.Active && PollComponent.ballotMap.has(this.answer.Ballot)) {
          const type = PollComponent.ballotMap.get(this.answer.Ballot);
          const comp = this.loadSubComponent(SubComponentId.Ballot, type) as PollBallotComponent;

          this.subscriptions[SubComponentId.Ballot].push(
            comp.previousRoundBallot.subscribe({
              next: (ballot: PollBallot) => this.previousRoundBallot = ballot,
            }),
            comp.currentRoundBallot.subscribe({
              next: (ballot: PollBallot) => this.currentRoundBallot = ballot,
            }),
            comp.justVoteBallot.subscribe({
              next: (ballot: PollBallot) => {
                this.justVoteBallot = ballot;
                this.clearSubComponent(SubComponentId.Ballot);
              }
            }),
          )
        }

        // Update Information component
        if (PollComponent.informationMap.has(this.answer.Information)) {
          const type = PollComponent.informationMap.get(this.answer.Information);
          const comp =
            this.loadSubComponent(SubComponentId.Information, type) as PollInformationComponent;
          comp.finalResult = !this.answer.Active;
        }

      },
      error: (err: HttpErrorResponse) => {
        if (err.status == 403 && !this.session.logged) {
          this.session.logNow();
        } else {
          this.registerError({status: err.status, message: err.error.trim()});
        }
      }
    });
  }

  /** Disconnect then remove a sub-component. */
  private clearSubComponent(componentIndex: number): void {
    if (!!this.subscriptions[componentIndex]) {
      for (let subscription of this.subscriptions[componentIndex]) {
        subscription.unsubscribe();
      }
    }
    this.viewContainerRef(componentIndex).clear();
  }

  /**
   * Create, connect and insert a sub-component.
   * The returned value is guaranteed to be of type type.
   */
  private loadSubComponent(componentIndex: number,
                           type: Type<PollSubComponent>): PollSubComponent {

    this.clearSubComponent(componentIndex);

    const viewContainerRef = this.viewContainerRef(componentIndex);
    const instance =
      this.dynamicComponentFactory.createComponent<PollSubComponent>(viewContainerRef, type);
    instance.pollSegment = this.segment;

    this.subscriptions[componentIndex] = [instance.errors.subscribe({
      next: (err: ServerError) => {
        this.registerError(err);
      }
    })];

    return instance;
  }

  private viewContainerRef(componentIndex: number): ViewContainerRef {
    switch (componentIndex) {
      case SubComponentId.Ballot:
        if (this.ballot === undefined) {
          return undefined;
        }
        return this.ballot.viewContainerRef;
      case SubComponentId.Information:
        if (this.information === undefined) {
          return undefined;
        }
        return this.information.viewContainerRef;
    }
    return undefined;
  }

  /**
   * Receive an error from a sub-component.
   * This results in all sub-components being cleared.
   */
  private registerError(err: ServerError) {
    this.error = err;
    for (let i = 0, end = this.subscriptions.length; i < end; i++) {
      this.clearSubComponent(i);
    }
    this.subscriptions = [];
  }

}
