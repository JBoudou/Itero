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
import { Component, Input, Output, EventEmitter, OnInit } from '@angular/core';

import { 
  BLANK_BALLOT,
  NONE_BALLOT,
  PollBallot,
  PollBallotComponent,
  ServerError,
} from '../poll/common';

import { PollAlternative, UninomialBallotAnswer, UninomialVoteQuery, BallotType } from '../api';


export class UninomialBallot implements PollBallot {
  constructor(readonly id: number, readonly name: string) { }

  get type(): BallotType { return BallotType.Uninomial; }
  get asString(): string { return this.name; }
}


@Component({
  selector: 'app-uninominal-ballot',
  templateUrl: './uninominal-ballot.component.html',
  styleUrls: ['./uninominal-ballot.component.sass']
})
export class UninominalBallotComponent implements OnInit, PollBallotComponent {

  @Input() pollSegment: string;

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
    this.http.get<UninomialBallotAnswer>('/a/ballot/uninominal/' + this.pollSegment).subscribe({
      next: (answer: UninomialBallotAnswer) => {
        this.alternatives = answer.Alternatives;
        this.previousRoundBallot.emit(this.ballotFromId(answer.Previous));
        this.currentRoundBallot .emit(this.ballotFromId(answer.Current ));
      },
      error: (err: HttpErrorResponse) => {
        this.errors.emit({status: err.status, message: err.error.trim()});
      }
    });
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
    this.vote({Alternative: this.selected})
  }

  onAbstain(): void {
    this.vote({Blank: true})
  }

  private vote(vote: UninomialVoteQuery): void {
    this.http.post('/a/vote/uninominal/' + this.pollSegment, vote)
      .subscribe({
        next: _ => this.justVoteBallot.emit(this.ballotFromQuery(vote)),
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

  private ballotFromId(vote: number|undefined): PollBallot {
    if (vote === undefined) {
      return NONE_BALLOT;
    }
    return new UninomialBallot(vote!, this.nameOf(vote)!);
  }

  private ballotFromQuery(vote: UninomialVoteQuery): PollBallot {
    if (vote.Blank) {
      return BLANK_BALLOT;
    }
    return new UninomialBallot(vote.Alternative, this.nameOf(vote.Alternative));
  }

}
