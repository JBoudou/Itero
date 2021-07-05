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

import { every, setEqual } from './collections'

describe('every', () => {
  const even = (a: number): boolean => a % 2 == 0

  it('returns true when all', () => {
    expect(every([2,8,4,0], even)).toBeTrue()
  })

  it('returns false when some', () => {
    expect(every([1,2,8,4,0], even)).toBeFalse()
    expect(every([2,8,27,4,0], even)).toBeFalse()
    expect(every([2,8,4,0,77], even)).toBeFalse()
  })
})

describe('setEqual', () => {
  it('returns true', () => {
    const a = new Set<number>([1,2,3,4])
    const b = new Set<number>([2,4,1,3])
    expect(setEqual(a,b)).toBeTrue()
    expect(setEqual(b,a)).toBeTrue()
  })
  it('finds one missing value', () => {
    const a = new Set<number>([1,2,3,4])
    const b = new Set<number>([2,1,3])
    expect(setEqual(a,b)).toBeFalse()
    expect(setEqual(b,a)).toBeFalse()
  })
})
