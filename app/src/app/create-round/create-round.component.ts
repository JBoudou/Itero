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

import { Component, OnInit } from '@angular/core';
import { FormBuilder, Validators } from '@angular/forms';

import { CreateService } from '../create/create.service';
import { CreateSubComponentBase } from '../create/create-sub-component-base';

@Component({
  selector: 'app-create-round',
  templateUrl: './create-round.component.html',
  styleUrls: ['./create-round.component.sass']
})
export class CreateRoundComponent extends CreateSubComponentBase implements OnInit {

  form = this.formBuilder.group({
    Deadline: [new Date(Date.now())],
    MinNbRounds: [2],
    MaxNbRounds: [10],
    MaxRoundDuration: [24*3600*1000],
  });

  constructor(
    protected service: CreateService,
    private formBuilder: FormBuilder,
  ) {
    super();
  }

  ngOnInit(): void {
    this.initModel();
  }

}
