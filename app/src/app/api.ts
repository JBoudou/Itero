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

export class LoginInfo {
  User: string;
  Passwd: string;
}

export class ListResponseEntry {
  s: string; // segment
  t: string; // title
  c: number; // current round
  m: number; // max nb rounds
  d: string; // deadline
  a: string; // action (abbreviated)
}
