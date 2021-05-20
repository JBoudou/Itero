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

import { Observable } from 'rxjs';

import { BallotType } from '../api';
import { ServerError } from '../shared/server-error';

/** 
 * Base interface for all ballots.
 * The methods are usualy implemented as getters.
 * TODO: Implement a decorator.
 */
export interface PollBallot {
  readonly type: BallotType;
  readonly asString: string;
}

/** Ballot to use when there is no ballot, for instance when a component is initialized. */
export const NONE_BALLOT: PollBallot  = {
  get type(): BallotType { return BallotType.None; },
  get asString(): string { return ''; }
}

export const BLANK_BALLOT: PollBallot = {
  get type(): BallotType { return BallotType.Blank; },
  get asString(): string { return ''; }
}

/**
 * Child component of the poll, namely either the ballot component and the information component.
 * Use dedicated interface PollBallotComponent and PollInformationComponent whenever possible.
 */
export interface PollSubComponent {
  pollSegment: string;
  round: number|undefined;
  errors: Observable<ServerError>;
}

/**
 * Child componenent of the poll that displays the ballot.
 * This component makes a request to the middleware to obtain the previous votes of the user (for the previous
 * and current round) as well as the list of alternatives. The previous votes are transfered (for instance to
 * the poll component) by means of Observables. This component is also responsible to send the vote to the
 * middleware. When that's done, the vote is transfered too.
 */
export interface PollBallotComponent extends PollSubComponent {
  previousRoundBallot: Observable<PollBallot>;
  currentRoundBallot : Observable<PollBallot>;
  justVoteBallot     : Observable<PollBallot>;
}

/**
 * Child component of the poll that displays information about the current state of the poll.
 * This component makes the request to the middleware itself. When the poll is closed,
 * the information are usualy displayed differently. Most notably, the winner of the poll must be made clear.
 */
export interface PollInformationComponent extends PollSubComponent {
  winner: Observable<string>;
}
