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

import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Component, Input, Output, EventEmitter, OnInit, OnDestroy } from '@angular/core';

import { take } from 'rxjs/operators';

import { 
  BLANK_BALLOT,
  NONE_BALLOT,
  PollBallot,
  PollBallotComponent,
} from '../common';

import { PollAlternative, UninominalBallotAnswer, UninominalVoteQuery, BallotType } from 'src/app/api';
import { ServerError } from 'src/app/shared/server-error';


export class UninominalBallot implements PollBallot {
  constructor(readonly id: number, readonly name: string) { }

  get type(): BallotType { return BallotType.Uninominal; }
  get asString(): string { return this.name; }
}


@Component({
  selector: 'app-uninominal-ballot',
  templateUrl: './uninominal-ballot.component.html',
  styleUrls: ['./uninominal-ballot.component.sass']
})
export class UninominalBallotComponent implements OnInit, OnDestroy, PollBallotComponent {

  @Input() pollSegment: string;
  @Input() round: number|undefined;

  @Output() errors = new EventEmitter<ServerError>();
  @Output() previousRoundBallot = new EventEmitter<PollBallot>();
  @Output() currentRoundBallot  = new EventEmitter<PollBallot>();
  @Output() justVoteBallot      = new EventEmitter<PollBallot>();

  alternatives: PollAlternative[];

  private selected: number;

  constructor(
    private http: HttpClient,
  ) { }

  ngOnInit(): void {
    this.http.get<UninominalBallotAnswer>('/a/ballot/uninominal/' + this.pollSegment)
      .pipe(take(1)).subscribe({
      next: (answer: UninominalBallotAnswer) => {
        this.alternatives = answer.Alternatives;
        this.previousRoundBallot.emit(this.ballotFromId(answer, 'Previous'));
        this.currentRoundBallot .emit(this.ballotFromId(answer, 'Current' ));
      },
      error: (err: HttpErrorResponse) => {
        this.errors.emit(new ServerError(err, 'retrieving uninominal ballot'));
      }
    });
  }

  ngOnDestroy(): void {
    this.errors.complete();
    this.previousRoundBallot.complete();
    this.currentRoundBallot.complete();
    this.justVoteBallot.complete();
  }

  onSelect(id: number): void {
    this.selected = id;
  }

  isSelected(id: number): boolean {
    return this.selected == id;
  }

  isValid(): boolean {
    return typeof this.selected != 'undefined';
  }

  onVote(): void {
    this.vote({Alternative: this.selected, Round: this.round})
  }

  onAbstain(): void {
    this.vote({Blank: true, Round: this.round})
  }

  private vote(vote: UninominalVoteQuery): void {
    this.http.post('/a/vote/uninominal/' + this.pollSegment, vote)
      .pipe(take(1)).subscribe({
        next: _ => this.justVoteBallot.emit(this.ballotFromQuery(vote)),
        error: (err: HttpErrorResponse) => {
          this.errors.emit(new ServerError(err, 'sending vote'));
        }
      });
  }

  /** The name of an alternative. */
  private nameOf(id: number|undefined): string|null {
    if (id === undefined) {
      return null;
    }
    var alternative: PollAlternative;
    for (alternative of this.alternatives) {
      if (alternative.Id == id!) {
        return alternative.Name;
      }
    }
    return null;
  }

  private ballotFromId(answer: UninominalBallotAnswer, prop: string): PollBallot {
    if (answer[prop] === undefined) {
      return !!answer[prop + 'IsBlank'] ? BLANK_BALLOT : NONE_BALLOT;
    }
    const vote = answer[prop] as number;
    return new UninominalBallot(vote!, this.nameOf(vote)!);
  }

  private ballotFromQuery(vote: UninominalVoteQuery): PollBallot {
    if (vote.Blank) {
      return BLANK_BALLOT;
    }
    return new UninominalBallot(vote.Alternative, this.nameOf(vote.Alternative));
  }

}
