// Itero - Online iterative vote application
// Copyright (C) 2020 Joseph Boudou, Yifan Zeng, Wan JIN
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
  ComponentRef,
  Directive,
  OnDestroy,
  OnInit,
  Type,
  ViewChild,
  ViewContainerRef,
  ViewEncapsulation,
} from '@angular/core';

import { ActivatedRoute, ParamMap } from '@angular/router';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { FormBuilder } from '@angular/forms';

import { Observable, Subscription } from 'rxjs';
import { map, take } from 'rxjs/operators';

import {
  NONE_BALLOT,
  PollBallot,
  PollBallotComponent,
  PollInformationComponent,
  PollSubComponent,
} from './common';

import { PollAnswer, BallotType, InformationType, PollNotifAnswerEntry } from '../api';
import { DynamicComponentFactoryService } from '../dynamic-component-factory.service';
import { SessionService } from '../session/session.service';
import { AppTitleService } from '../app-title.service';
import { PollNotifService } from '../poll-notif.service';
import { Suspendable } from '../shared/suspender';
import { ServerError } from '../shared/server-error';

import { UninominalBallotComponent } from './uninominal-ballot/uninominal-ballot.component';
import { CountsInformationComponent } from './counts-information/counts-information.component';
import { ResponsiveBreakpointService, ResponsiveState } from '../responsive-breakpoint.service';

@Directive({
  selector: '[PollBallot]',
})
export class PollBallotDirective {
  constructor(public viewContainerRef: ViewContainerRef) { }
}

@Directive({
  selector: '[PollInformation]',
})
export class PollInformationDirective {
  constructor(public viewContainerRef: ViewContainerRef) { }
}

@Directive({
  selector: '[PollPrevious]',
})
export class PollPreviousDirective {
  constructor(public viewContainerRef: ViewContainerRef) { }
}

// Indexes for the sub component.
const enum SubComponentId {
  Ballot = 0,
  Information,
  Previous
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
export class PollComponent implements OnInit, OnDestroy {

  // Anchors to insert the dynamic sub-component into.
  @ViewChild(PollBallotDirective, { static: false }) ballot: PollBallotDirective;
  @ViewChild(PollInformationDirective, { static: false }) information: PollInformationDirective;
  @ViewChild(PollPreviousDirective, { static: false }) previousInformation: PollPreviousDirective;

  previousForm = this.formBuilder.group({
    round: [1]
  });

  segment: string;
  answer: PollAnswer;
  winner$: Observable<string>;
  displayedResult: number|undefined;

  error: ServerError;
  localError : false | 'next-round' | 'unverified' = false;

  previousRoundBallot: PollBallot = NONE_BALLOT;
  currentRoundBallot : PollBallot = NONE_BALLOT;
  justVoteBallot     : PollBallot = NONE_BALLOT;

  // Make it visible from template.
  // TODO: Implements a decorator for PollBallot that provides methods for that.
  BallotType = BallotType;

  // Whether the information must be displayed inside an InfoPanelComponent.
  infoOnPanel$ : Observable<boolean>

  /** Subscription for the sub component. The first index must be a SubComponentId. */
  private subsubscriptions: Subscription[][] = [];

  /** Other subscriptions, not corresponding to any subcomponent. */
  private subscriptions: Subscription[] = [];

  /** Currently loaded components, indexed by SubComponentId. **/
  private components: ComponentRef<PollSubComponent>[] = [];
  
  constructor(
    private dynamicComponentFactory: DynamicComponentFactoryService,
    private formBuilder: FormBuilder,
    private http: HttpClient,
    private notif: PollNotifService,
    private route: ActivatedRoute,
    private session: SessionService,
    private title: AppTitleService,
    private responsive: ResponsiveBreakpointService,
  ) {
    this.infoOnPanel$ = this.responsive.state$
      .pipe(map((st: ResponsiveState): boolean => st !== ResponsiveState.Laptop))
  }

  ngOnInit(): void {
    this.route.paramMap.pipe(take(1)).subscribe((params: ParamMap) => {
      this.segment = params.get('pollSegment');
      this.retrieveTypes();
    });
    this.subscriptions.push(
      this.notif.event$.subscribe({
        next: (evt: PollNotifAnswerEntry) => this.handleEvent(evt),
      }),
    )
  }

  ngOnDestroy(): void {
    this.subscriptions.forEach((sub: Subscription) => sub.unsubscribe());
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

  roundDeadlinePassed(): boolean {
    return this.answer.RoundDeadline.getTime() < Date.now();
  }

  pollDeadlinePassed(): boolean {
    return this.answer.PollDeadline.getTime() < Date.now();
  }

  displayPreviousResults(): boolean {
    return this.answer.CurrentRound >= 2 &&
           PollComponent.informationMap.has(this.answer.Information);
  }

  lastDisplayRound(): number {
    return !!this.answer.Active ? this.answer.CurrentRound - 1 : this.answer.CurrentRound;
  }

  pollEndCase(): string {
    if (this.pollDeadlinePassed()) {
      return this.answer.CurrentRound + 1 === this.answer.MinNbRounds ? 'current' : 'deadlinePassed';
    }
    if (this.answer.CurrentRound + 1 === this.answer.MaxNbRounds &&
        this.answer.RoundDeadline.getTime() < this.answer.PollDeadline.getTime()) {
      return 'current';
    }
    if (this.answer.MinNbRounds === this.answer.MaxNbRounds) {
      return 'deadlinePassed'
    }
    return this.answer.CurrentRound >= this.answer.MinNbRounds ? 'minExceeded' : 'full';
  }

  onPreviousResult(): void {
    const round = this.previousForm.value.round;
    if (!PollComponent.informationMap.has(this.answer.Information)) {
      this.displayedResult = undefined;
      console.warn('Wrong inforamtion type');
      return;
    }
    if (!Number.isInteger(round)) {
      this.displayedResult = undefined;
      console.warn('No value selected');
      return;
    }

    const type = PollComponent.informationMap.get(this.answer.Information);
    const {comp} = this.loadSubComponent(SubComponentId.Previous, type);
    comp.round = round - 1;
    this.displayedResult = round;
  }

  private static ballotMap = new Map<BallotType, Type<PollBallotComponent>>([
    [BallotType.Uninominal, UninominalBallotComponent]
  ]);

  private static informationMap = new Map<InformationType, Type<PollSubComponent>>([
    [InformationType.Counts, CountsInformationComponent]
  ]);

  /** Ask the type of the poll to the middleware and creates the sub-components accordingly. */
  private retrieveTypes(): void {
    this.http.get('/a/poll/' + this.segment, {responseType: 'text'}).pipe(take(1)).subscribe({
      next: (body: string) => {
        this.answer = PollAnswer.fromJSON(body)
        this.title.setTitle(this.answer.Title)
        // We need the ViewChilds to appear before inserting components in them.
        setTimeout(() => this.synchronizeSubComponents(), 0);
      },
      error: (err: HttpErrorResponse) => {
        if (err.status === 403 && err.error.trim() === 'Unlogged' && !this.session.logged) {
          this.session.logNow();
        } else if (err.status === 403 && err.error.trim() === 'Unverified' && this.session.logged) {
          this.localError = 'unverified'
        } else {
          this.registerError(new ServerError(err, 'retrieving poll information'));
        }
      }
    });
  }

  private synchronizeSubComponents(): void {
    // Update Ballot component
    if (this.answer.Active && PollComponent.ballotMap.has(this.answer.Ballot)) {
      const type = PollComponent.ballotMap.get(this.answer.Ballot);
      const load = this.loadSubComponent(SubComponentId.Ballot, type)
      const comp = load.comp as PollBallotComponent;
      comp.round = this.answer.CurrentRound;

      if (!load.alreadyThere) {
        this.subsubscriptions[SubComponentId.Ballot].push(
          comp.previousRoundBallot.subscribe({
            next: (ballot: PollBallot) => this.previousRoundBallot = ballot,
          }),
          comp.currentRoundBallot.subscribe({
            next: (ballot: PollBallot) => this.currentRoundBallot = ballot,
          }),
          comp.justVoteBallot.subscribe({
            next: (ballot: PollBallot) => {
              this.justVoteBallot = ballot;
              this.localError = false;
              this.clearSubComponent(SubComponentId.Ballot);
              this.refresh.suspend(5000);
            }
          }),
        )
      }
    }

    // Update Information component
    if (PollComponent.informationMap.has(this.answer.Information)) {
      const type = PollComponent.informationMap.get(this.answer.Information);
      const comp =
        this.loadSubComponent(SubComponentId.Information, type).comp as PollInformationComponent;
      comp.round = this.answer.CurrentRound - 1;
      this.winner$ = comp.winner;
    }
  }

  /** Disconnect then remove a sub-component. */
  private clearSubComponent(componentIndex: number): void {
    if (!!this.subsubscriptions[componentIndex]) {
      for (let subscription of this.subsubscriptions[componentIndex]) {
        subscription.unsubscribe();
      }
    }
    this.viewContainerRef(componentIndex)?.clear();
    if (this.components[componentIndex] !== undefined) {
      this.components[componentIndex].destroy();
      this.components[componentIndex] = undefined;
    }
  }

  /**
   * Ensures that the given component has the given type.
   * If it is already the case, the current component is returned and `alreadyThere` is set to true.
   * Otherwise, the previous component is detached and destroyed, a new one is created, inserted,
   * and returned, and `alreadyThere` is set to false.
   */
  private loadSubComponent(componentIndex: number,
                           type: Type<PollSubComponent>): {comp: PollSubComponent, alreadyThere: boolean} {

    if (this.components[componentIndex] !== undefined &&
        this.components[componentIndex].componentType === type) {
      return {comp: this.components[componentIndex].instance, alreadyThere: true}
    }

    this.clearSubComponent(componentIndex);

    const viewContainerRef = this.viewContainerRef(componentIndex);
    this.components[componentIndex] =
      this.dynamicComponentFactory.createComponent<PollSubComponent>(viewContainerRef, type);
    const instance = this.components[componentIndex].instance
    instance.pollSegment = this.segment;

    this.subsubscriptions[componentIndex] = [instance.errors.subscribe({
      next: (err: ServerError) => {
        this.registerError(err);
      }
    })];

    return {comp: instance, alreadyThere: false}
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
      case SubComponentId.Previous:
        if (this.previousInformation === undefined) {
          return undefined;
        }
        return this.previousInformation.viewContainerRef;
    }
    return undefined;
  }

  /**
   * Receive an error from a sub-component.
   * This results in all sub-components being cleared.
   */
  private registerError(err: ServerError) {
    if (err.status == 423 && err.message == "Next round") {
      this.localError = 'next-round';
      this.refresh();
      return;
    }
    this.localError = false;
    this.error = err;
    for (let i = 0, end = this.subsubscriptions.length; i < end; i++) {
      this.clearSubComponent(i);
    }
    this.subsubscriptions = [];
  }

  private handleEvent(evt: PollNotifAnswerEntry): void {
    if (evt.Segment != this.segment) {
      return
    }
    this.refresh();
  }

  private readonly refresh = Suspendable(function(): void {
    this.previousRoundBallot = NONE_BALLOT;
    this.currentRoundBallot  = NONE_BALLOT;
    this.justVoteBallot      = NONE_BALLOT;
    this.retrieveTypes();
  });

}
