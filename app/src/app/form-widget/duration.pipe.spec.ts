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

import { DetailedDuration, DurationPipe } from './duration.pipe';

describe('DetailedDuration', () => {
  it('convert from milliseconds', () => {
    expect(DetailedDuration.fromMilliseconds(       60 * 1000).equals({ mins : 1 })).toBeTruthy();
    expect(DetailedDuration.fromMilliseconds(     3600 * 1000).equals({ hours: 1 })).toBeTruthy();
    expect(DetailedDuration.fromMilliseconds(24 * 3600 * 1000).equals({ days : 1 })).toBeTruthy();
  });

  it('is consistent', () => {
    for (const val of [3600 * 1000, 12 * 3600 * 1000, 60 * 1000]) {
      const detailed = DetailedDuration.fromMilliseconds(val);
      expect(detailed.toMilliseconds()).toBe(val);
    }
  });
});

describe('DurationPipe', () => {
  it('create an instance', () => {
    const pipe = new DurationPipe();
    expect(pipe).toBeTruthy();
  });
});
