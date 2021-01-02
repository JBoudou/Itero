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

/* This file lists classes used in the communication between the front end and
 * the middleware. */

export class LoginQuery {
  User: string;
  Passwd: string;
}

export class SignupQuery {
  Name: string;
  Email: string;
  Passwd: string;
}

// Null-able Date
export type NuDate = Date | '.'

export enum PollAction {
  Vote,
  Modi,
  Part,
  Term
}

export class ListAnswerEntry {
  Segment:      string;
  Title:        string;
  CurrentRound: number;
  MaxRound:     number;
  Deadline:     NuDate;
  Action:       PollAction;

  deadlineCategory?: string;
}

export enum BallotType {
  // Front end types.
  None = -255,
  Blank,

  // API types.
  Closed = 0,
  Uninomial,
}

export enum InformationType {
  NoneYet,
  Counts,
}

export class PollAnswer {
  Title:        string;
  Description:  string;
  Admin:        string;
  CreationTime: Date;
  CurrentRound: number;
  Active:       boolean;
  Ballot:       BallotType;
  Information:  InformationType;
}

export class PollAlternative {
  Id:   number;
  Name: string;
  Cost: number;
}

export interface UninomialBallotAnswer {
  Previous?: number;
  Current?: number;
  Alternatives: Array<PollAlternative>;
}

export class UninomialVoteQuery {
  Blank?: boolean;
  Alternative?: number;
}

export class CountInfoEntry {
  Alternative: PollAlternative;
  Count: number;
}

export class CountInfoAnswer {
  Result: Array<CountInfoEntry>;
}
