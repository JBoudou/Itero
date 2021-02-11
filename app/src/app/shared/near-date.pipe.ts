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

import {formatDate} from '@angular/common';
import { Pipe, PipeTransform, Inject, LOCALE_ID } from '@angular/core';

/**
 * Transform a date into a simple expression.
 * Returned expressions have forms like 'tomorrow at 2:00 PM'.
 * 
 * ### Format
 * | Format | Description         | Example              |
 * |--------|---------------------|----------------------|
 * | inside | Inside a sentence   | on January 13, 2021  |
 * | noPrep | Without preposition | January 13, 2021     |
 * | alone  | Starting a sentence | Yesterday at 6:00 AM |
 */
@Pipe({
  name: 'nearDate'
})
export class NearDatePipe implements PipeTransform {
  constructor(@Inject(LOCALE_ID) private locale: string) {}

  transform(value: Date, format?: string, locale?: string): string {
    if (value === undefined) {
      return undefined;
    }

    const today = new Date(Date.now());
    today.setHours(0, 0, 0);

    const yesterday = new Date(today.valueOf() - (24 * 3600 * 1000));
    if (value < yesterday) {
      return this.raw(value, format, locale);
    }

    if (value < today) {
      return this.day('yesterday', value, format, locale);
    }

    const tomorrow = new Date(today.valueOf() + (24 * 3600 * 1000));
    if (value < tomorrow) {
      return this.day('today', value, format, locale);
    }

    const afterTomorrow = new Date(today.valueOf() + (2 * 24 * 3600 * 1000));
    if (value < afterTomorrow) {
      return this.day('tomorrow', value, format, locale);
    }

    const nextWeek = new Date(today.valueOf() + (8 * 24 * 3600 * 1000));
    if (value < nextWeek) {
      const day = 'next ' + formatDate(value, 'EEEE', locale || this.locale);
      return this.day(day, value, format, locale);
    }

    return this.raw(value, format, locale);
  }

  private raw(value: Date, format: string, locale: string): string {
    const base = formatDate(value, 'longDate', locale || this.locale);
    return format === 'inside' ? 'on ' + base : base;
  }

  private day(day: string, value: Date, format: string, locale: string): string {
    if (format === 'alone') {
      day = day[0].toUpperCase() + day.substr(1);
    }
    return day + ' at ' + formatDate(value, 'shortTime', locale || this.locale);
  }

}
