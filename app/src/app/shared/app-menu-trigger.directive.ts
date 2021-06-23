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

import { Directive, Input } from '@angular/core';
import { MatMenuPanel, MatMenuTrigger } from '@angular/material/menu';

/**
 * AppMenuTrigger is a subclass of MatMenuTrigger providing the possibility to not always open the
 * menu on click events.
 *
 * When a click event is raised a filter function is called with the event to decide whether the
 * menu must be opened. By default this function always returns false, meaning that the menu is not
 * opened. The input parameter `appMenuTriggerFilter` can be used to change the filter function.
 */
@Directive({
  selector: '[appMenuTriggerFor]'
})
export class AppMenuTrigger extends MatMenuTrigger {

  @Input('appMenuTriggerFor')
  get menu() { return super.menu }
  set menu(menu: MatMenuPanel) { super.menu = menu }

  @Input('appMenuTriggerFilter') _filter: (evt: Event) => boolean = (_: MouseEvent) => false

  _handleClick(evt: MouseEvent): void {
    if (this._filter(evt)) {
      super._handleClick(evt)
    }
  }

}
