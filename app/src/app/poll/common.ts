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

export class ServerError {
  status: number;
  message: string;
}

export interface PollBallot {
  readonly type: BallotType;
  readonly asString: string;
}

export const NONE_BALLOT: PollBallot  = {
  get type(): BallotType { return BallotType.None; },
  get asString(): string { return ''; }
}

export const BLANK_BALLOT: PollBallot = {
  get type(): BallotType { return BallotType.Blank; },
  get asString(): string { return ''; }
}

export interface PollSubComponent {
  pollSegment: string;
  errors: Observable<ServerError>;
}

export interface PollBallotComponent extends PollSubComponent {
  previousRoundBallot: Observable<PollBallot>;
  currentRoundBallot : Observable<PollBallot>;
  justVoteBallot     : Observable<PollBallot>;
}

export interface PollInformationComponent extends PollSubComponent {
  finalResult: boolean;
}
