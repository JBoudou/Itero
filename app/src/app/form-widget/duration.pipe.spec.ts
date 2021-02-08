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
  const minute =     60 * 1000;
  const hour =     3600 * 1000;
  const day = 24 * 3600 * 1000;

  let pipe: DurationPipe;

  beforeEach(() => {
    pipe = new DurationPipe();
  });

  it('create an instance', () => {
    pipe = new DurationPipe();
    expect(pipe).toBeTruthy();
  });

  it('display minutes correctly', () => {
    expect(pipe.transform(    minute)).toBe('1 minute' );
    expect(pipe.transform(2 * minute)).toBe('2 minutes');
  });

  it('display hours correctly', () => {
    expect(pipe.transform(    hour)).toBe('1 hour' );
    expect(pipe.transform(2 * hour)).toBe('2 hours');
  });

  it('display days correctly', () => {
    expect(pipe.transform(    day)).toBe('1 day' );
    expect(pipe.transform(2 * day)).toBe('2 days');
  });

  it('display two values correctly', () => {
    expect(pipe.transform(day + hour  , 2)).toBe('1 day and 1 hour'  );
    expect(pipe.transform(day + minute, 2)).toBe('1 day and 1 minute');
  });

  it('display three values correctly', () => {
    expect(pipe.transform(day + hour + minute, 3)).toMatch('^1 day, 1 hour,? and 1 minute$');
  });

  it('display only one value', () => {
    expect(pipe.transform(day + hour + minute, 1)).toBe('1 day');
  });

  it('display only two values', () => {
    expect(pipe.transform(day + hour + minute, 2)).toBe('1 day and 1 hour');
  });

});
