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

import { ScheduleOne } from './schedule-one';

describe('ScheduleOne', () => {
  let object: ScheduleOne;

  beforeEach(() => {
    object = new ScheduleOne();
    jasmine.clock().install();
  })

  afterEach(() => {
    jasmine.clock().uninstall();
  });

  it('wait the given delay', () => {
    jasmine.clock().mockDate();
    const begin = new Date();
    let end: Date;
    
    object.schedule(() => end = new Date(), 200);
    jasmine.clock().tick(200);

    expect(end).toBeTruthy();
    expect(end.getTime() - begin.getTime()).toBeGreaterThanOrEqual(200);
  });

  it('cancel previous scheduled tasks', () => {
    let firstCalled  = false;
    let secondCalled = false;

    object.schedule(() => firstCalled  = true, 200);
    jasmine.clock().tick(50);
    object.schedule(() => secondCalled = true, 100);
    jasmine.clock().tick(200);

    expect(firstCalled ).toBeFalse();
    expect(secondCalled).toBeTrue();
  });

  it('wait the given delay for the canceling task', () => {
    jasmine.clock().mockDate();

    object.schedule(() => 0, 200);
    jasmine.clock().tick(50);

    const begin = new Date();
    let end: Date;
    object.schedule(() => end = new Date(), 200);
    jasmine.clock().tick(200);

    expect(end).toBeTruthy();
    expect(end.getTime() - begin.getTime()).toBeGreaterThanOrEqual(200);
  });

});
