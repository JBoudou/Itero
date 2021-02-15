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

/**
 * Display numbers as localized ordinal.
 *
 * Do NOT pipe to TitleCase because it will render 4 as 4Th;
 * use the optional titleCase argument instead.
 */
@Pipe({
  name: 'ordinal'
})
export class OrdinalPipe implements PipeTransform {

  transform(value: number, titleCase?: boolean): string {
    switch (value) {
    case undefined:
      return '';
    case 1:
      return titleCase ? 'First' : 'first';
    case 2:
      return titleCase ? 'Second' : 'second';
    case 3:
      return titleCase ? 'Third' : 'third';
    default:
      return value.toString() + 'th';
    }
  }

}
