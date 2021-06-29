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

export class SessionAnswer {
  SessionId: string;
  Expires: Date;
  Verified: boolean;

  static fromObject(obj: any): SessionAnswer {
    const ret = {} as SessionAnswer;
    if (typeof obj.SessionId === 'string') {
      ret.SessionId = obj.SessionId
    }
    if ('Expires' in obj) {
      if (typeof obj.Expires === 'string') {
        ret.Expires = new Date(obj.Expires)
      }
      if (obj.Expires instanceof Date) {
        ret.Expires = obj.Expires
      }
    }
    if ('Profile' in obj && 'Verified' in obj.Profile) {
      ret.Verified = obj.Profile.Verified
    }
    return ret
  }
}

export enum PollAction {
  Vote,
  Modi,
  Part,
  Term,
  Wait
}

export class ListAnswerEntry {
  Segment:      string;
  Title:        string;
  CurrentRound: number;
  MaxRound:     number;
  Deadline:     Date|undefined;
  Action:       PollAction;
  Deletable:    boolean;
}

export class ListAnswer {
  Public: ListAnswerEntry[];
  Own:    ListAnswerEntry[];

  static fromJSON(json: string): ListAnswer {
    const ret = JSON.parse(json, function(key: string, value: any) {
      if (key === 'Deadline') {
        return value === '⋅' ? undefined : new Date(value);
      }
      return value;
    });

    const normalize = function (entry: any) {
      if (!entry.hasOwnProperty('Deadline')) {
        entry.Deadline = undefined;
      }
      if (!entry.hasOwnProperty('Deletable')) {
        entry.Deletable = false;
      }
    }
    ret.Public.forEach(normalize);
    ret.Own   .forEach(normalize);

    return ret;
  }
}

export enum BallotType {
  // Front end types.
  None = -255,
  Blank,

  // API types.
  Closed = 0,
  Uninominal,
}

export enum InformationType {
  NoneYet,
  Counts,
}

export class PollAnswer {
  Title:            string;
  Description:      string;
  Admin:            string;
  CreationTime:     Date;
  CurrentRound:     number;
  Active:           boolean;
  State:            string;
  CarryForward:     boolean;
  Start:            Date;
  RoundDeadline:    Date;
  PollDeadline:     Date;
  MaxRoundDuration: number; // milliseconds
  MinNbRounds:      number;
  MaxNbRounds:      number;
  Ballot:           BallotType;
  Information:      InformationType;

  static fromJSON(raw: string): PollAnswer {
    return JSON.parse(raw, function(key: string, value: any): any{
      if ((typeof value === 'string' || typeof value === 'number') &&
          (key === 'CreationTime' ||
           key === 'Start' ||
           key === 'RoundDeadline' ||
           key === 'PollDeadline') )
      {
        return new Date(value)
      }
      return value
    })
  }
}

export interface PollAlternative {
  Id:   number;
  Name: string;
  Cost: number;
}

export interface UninominalBallotAnswer {
  Previous?: number;
  PreviousIsBlank?: boolean;
  Current?: number;
  CurrentIsBlank?: boolean;
  Alternatives: Array<PollAlternative>;
}

export interface UninominalVoteQuery {
  Blank?:       boolean;
  Alternative?: number;
  Round:        number;
}

export interface CountInfoEntry {
  Alternative: PollAlternative;
  Count: number;
}

export interface CountInfoAnswer {
  Result: Array<CountInfoEntry>;
}

export enum PollUserType {
  Simple,
}

export interface SimpleAlternative {
  Name: string;
  Cost: number;
}

export interface CreateQuery {
  UserType:         PollUserType;
  Title:            string;
  Description:      string;
  Hidden:           boolean;
  Start:            Date;
  Alternatives:     SimpleAlternative[];
  ReportVote:       boolean;
  MinNbRounds:      number;
  MaxNbRounds:      number;
  Deadline:         Date;
  MaxRoundDuration: number; // milliseconds
  RoundThreshold:   number;
}

export enum PollNotifAction {
  Start,
  Next,
  Term,
  Delete,
}

export class PollNotifAnswerEntry {
  Timestamp: Date;
  Segment:   string;
  Title:     string;
  Round:     number;
  Action:    PollNotifAction;

  static fromJSONList(json: string): PollNotifAnswerEntry[] {
    return JSON.parse(json, function(key: string, value: any) {
      if (key === 'Timestamp') { return new Date(value as string); }
      return value;
    });
  }
}

export interface ConfirmAnswer {
  Type: string
}
