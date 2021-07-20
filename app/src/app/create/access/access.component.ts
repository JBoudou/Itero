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

import { Component, OnInit, ChangeDetectionStrategy, OnDestroy, ViewChild, TemplateRef } from '@angular/core';
import { FormBuilder, Validators } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';

import { CreateService } from '../create.service';
import { CreateSubComponentBase } from '../create-sub-component-base';
import { SessionService } from 'src/app/session/session.service';

@Component({
  selector: 'app-access',
  templateUrl: './access.component.html',
  styleUrls: ['./access.component.sass'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AccessComponent extends CreateSubComponentBase implements OnInit, OnDestroy {

  @ViewChild('stepInfo') infoTemplate: TemplateRef<any>

  form = this.formBuilder.group({
    Electorate: [0],
    Hidden: [false],
  })

  constructor(
    public session: SessionService,
    protected service: CreateService,
    protected route: ActivatedRoute,
    private formBuilder: FormBuilder,
  ) {
    super();
  }

  ngOnInit(): void {
    this.initModel();
  }

  ngOnDestroy(): void {
    this.unsubscribeAll();
  }

}
