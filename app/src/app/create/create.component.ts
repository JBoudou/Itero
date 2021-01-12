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

import { CreateService, CREATE_TREE, APP_CREATE_TREE, CreateStepStatus, CreateNextStatus } from './create.service';

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
  createStepStatus: CreateStepStatus = new CreateStepStatus(0, []);

  constructor(
    private service: CreateService,
  ) {
    this.service.createStepStatus$.subscribe({
      next: (status: CreateStepStatus) => this.onCreateStepStatus(status),
    });
    this.service.createNextStatus$.subscribe({
      next: (status: CreateNextStatus) => this.onCreateNextStatus(status),
    });
  }

  ngOnInit(): void {
  }

  onBack(): void {
    this.service.back();
  }

  onNext(): void {
    this.service.next();
  }

  private onCreateStepStatus(status: CreateStepStatus): void {
    this.createStepStatus = status;
    this.canBack = status.current > 0;
  }

  private onCreateNextStatus(status: CreateNextStatus): void {
    this.canNext = status.validable;
    this.isValidate = status.final;
  }

}
