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

import { Pipe, PipeTransform } from '@angular/core';

export class DetailedDuration {
  constructor(
    public days: number,
    public hours: number,
    public mins: number
  ) { }

  toMilliseconds(): number {
    return ((this.days * 24 + this.hours) * 60 + this.mins) * 60 * 1000;
  }

  equals(other: Partial<DetailedDuration>): boolean {
    return ((other.days  ?? 0) === this.days ) &&
           ((other.hours ?? 0) === this.hours) &&
           ((other.mins  ?? 0) === this.mins );
  }

  static fromMilliseconds(milli: number): DetailedDuration {
    return new DetailedDuration(
      Math.floor(milli / (24 * 3600 * 1000)),
      Math.floor(milli /      (3600 * 1000)) % 24,
      Math.floor(milli /        (60 * 1000)) % 60,
    );
  }
}

const durationOrder = [ 'days', 'hours', 'mins' ]

const durationName = {
  days : 'day',
  hours: 'hour',
  mins : 'minute',
};

@Pipe({
  name: 'duration'
})
export class DurationPipe implements PipeTransform {

  transform(value: number, nbTerms?: number): string {
    if (isNaN(nbTerms)) nbTerms = 2;
    const detailed = DetailedDuration.fromMilliseconds(value);

    // construct the list
    let list: string[] = [];
    const end = durationOrder.length;
    for (let pos = 0; list.length < nbTerms && pos < end; pos++) {
      const prop = durationOrder[pos];
      const qtt = detailed[prop];
      if (qtt !== 0) {
        list.push(qtt.toString() + ' ' + durationName[prop] + (qtt > 1 ? 's' : ''));
      }
    }

    return this.languageJoin(list);
  }

  languageJoin(list: string[]): string {
    const len = list.length;
    if (len === 0) {
      return '';
    }
    if (len === 1) {
      return list[0];
    }
    return list.slice(0, -1).join(', ') + ' and ' + list[list.length - 1];
  }

}
