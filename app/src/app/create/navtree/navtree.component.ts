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

import { Component, OnInit, ChangeDetectionStrategy, Input, Output, EventEmitter, OnDestroy } from '@angular/core';

import { NavStepStatus } from './navstep.status';


@Component({
  selector: 'app-navtree',
  templateUrl: './navtree.component.html',
  styleUrls: ['./navtree.component.sass'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class NavtreeComponent implements OnInit, OnDestroy {

  @Input() stepStatus: NavStepStatus;
  @Input() validable: boolean = false;

  /** Send the index to navigate to. */
  @Output() navigateTo = new EventEmitter<number>();

  constructor() { }

  ngOnInit(): void {
  }

  ngOnDestroy(): void {
    this.navigateTo.complete();
  }

  isNavigable(index: number): boolean {
    return this.stepStatus.isNavigable(index, this.validable);
  }

  isEnabled(index: number): boolean {
    return index === this.stepStatus.current || this.isNavigable(index);
  }

}
