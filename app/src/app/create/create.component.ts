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

import { Component, OnInit, ViewEncapsulation } from '@angular/core';

import { CreateService, CreateNextStatus } from './create.service';
import { NavStepStatus } from './navtree/navstep.status';

@Component({
  selector: 'app-create',
  templateUrl: './create.component.html',
  styleUrls: ['./create.component.sass'],
  encapsulation: ViewEncapsulation.None,
})
export class CreateComponent implements OnInit {

  // TODO: Use await.
  canBack: boolean = false;
  canNext: boolean = false;
  isValidate: boolean = false;
  createStepStatus: NavStepStatus = new NavStepStatus(0, []);

  constructor(
    private service: CreateService,
  ) {
    this.service.createStepStatus$.subscribe({
      next: (status: NavStepStatus) => this.onCreateStepStatus(status),
    });
    this.service.createNextStatus$.subscribe({
      next: (status: CreateNextStatus) => this.onCreateNextStatus(status),
    });
  }

  isJumpable(pos: number): boolean {
    return pos >= 0 && ( (pos < this.createStepStatus.current) ||
                         (this.canNext && pos == this.createStepStatus.current + 1) );
  }

  ngOnInit(): void {
  }

  onJump(pos: number): void {
    if (!this.isJumpable(pos)) {
      return;
    }
    if (pos < this.createStepStatus.current) {
      this.service.back(this.createStepStatus.current - pos);
    } else {
      this.service.next();
    }
  }

  onBack(): void {
    this.service.back();
  }

  onNext(): void {
    this.service.next();
  }

  private onCreateStepStatus(status: NavStepStatus): void {
    this.createStepStatus = status;
    this.canBack = status.current > 0;
  }

  private onCreateNextStatus(status: CreateNextStatus): void {
    this.canNext = status.validable;
    this.isValidate = status.final;
  }

}
