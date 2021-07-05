// Itero - Online iterative vote application
// Copyright (C) 2021 Joseph Boudou
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

export function every<Elt>(it: Iterable<Elt>, predicate: (elt: Elt)=>boolean): boolean {
  for (let value of it) {
    if (!predicate(value)) {
      return false
    }
  }
  return true
}

export function setEqual<Elt>(a: Set<Elt>, b: Set<Elt>): boolean {
  return every(a, (elt: Elt) => b.has(elt)) &&
         every(b, (elt: Elt) => a.has(elt))
}
