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

import { NearDatePipe } from './near-date.pipe';

describe('NearDatePipe', () => {
  let pipe: NearDatePipe;

  beforeEach(() => {
    pipe = new NearDatePipe('en-US');
  });

  it('create an instance', () => {
    expect(pipe).toBeTruthy();
  });

  it('should add on only in inside format', () => {
    expect(pipe.transform(new Date('2012-01-01T00:00:00'), 'inside')).toMatch(/^on /);
    expect(pipe.transform(new Date('2012-01-01T00:00:00'), 'noPrep')).not.toMatch(/^on /);
    expect(pipe.transform(new Date('2012-01-01T00:00:00'), 'alone ')).not.toMatch(/^on /);
  });

  it('should start with an upper case only in alone mode', () => {
    expect(pipe.transform(new Date(), 'inside')).toMatch(/^today /);
    expect(pipe.transform(new Date(), 'noPrep')).toMatch(/^today /);
    expect(pipe.transform(new Date(), 'alone' )).toMatch(/^Today /);
  });

});
