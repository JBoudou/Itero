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
import { FormBuilder, Validators } from '@angular/forms';

import { PollSubComponent, ServerError } from '../poll/common';
import { PollAlternative, UninomialBallotAnswer, UninomialVoteQuery } from '../api';

@Component({
  selector: 'app-uninominal-ballot',
  templateUrl: './uninominal-ballot.component.html',
  styleUrls: ['./uninominal-ballot.component.sass']
})
export class UninominalBallotComponent implements OnInit, PollSubComponent {

  @Input() pollSegment: string;
  @Output() errors = new EventEmitter<ServerError>();

  answer: UninomialBallotAnswer;

  form = this.formBuilder.group({
    Choice: ['', [Validators.required]],
  });

  lastVote: UninomialVoteQuery;

  constructor(
    private http: HttpClient,
    private formBuilder: FormBuilder,
  ) { }

  ngOnInit(): void {
    this.http.get<UninomialBallotAnswer>('/a/ballot/uninominal/' + this.pollSegment).subscribe({
      next: (answer: UninomialBallotAnswer) => {
        this.answer = answer;
      },
      error: (err: HttpErrorResponse) => {
        this.errors.emit({status: err.status, message: err.error.trim()});
      }
    });
  }

  onVote(): void {
    this.vote({Alternative: this.form.value.Choice})
  }

  onAbstain(): void {
    this.vote({Blank: true})
  }

  private vote(vote: UninomialVoteQuery): void {
    this.http.post('/a/vote/uninominal/' + this.pollSegment, vote)
      .subscribe({
        next: _ => this.lastVote = vote,
      });
  }

  hasPrevious(): boolean {
    return this.answer !== undefined && this.answer.Previous !== undefined;
  }

  previous(): string|null {
    return this.nameOf(this.answer.Previous);
  }

  hasCurrent(): boolean {
    return !this.hasJust() && this.answer != undefined && this.answer.Current !== undefined;
  }

  current(): string|null {
    return this.nameOf(this.answer.Current);
  }

  hasJust(): boolean {
    return !!this.lastVote;
  }

  just(): string|null {
    return this.nameOf(this.lastVote.Alternative);
  }

  private nameOf(id: number|undefined): string|null {
    if (id === undefined) {
      return null;
    }
    var alternative: PollAlternative;
    for (alternative of this.answer.Alternatives) {
      if (alternative.Id == id!) {
        return alternative.Name;
      }
    }
    return null;
  }

}
