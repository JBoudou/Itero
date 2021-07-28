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

import { Component, ChangeDetectionStrategy, OnDestroy, ViewChild, TemplateRef, AfterViewInit } from '@angular/core';
import { FormBuilder, Validators } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';

import { MatCheckbox, MatCheckboxChange } from '@angular/material/checkbox';

import { CreateService } from '../create.service';
import { CreateSubComponentBase } from '../create-sub-component-base';
import { SessionService } from 'src/app/session/session.service';
import { CreateQuery } from 'src/app/api';
import { ServerError } from 'src/app/shared/server-error';

@Component({
  selector: 'app-access',
  templateUrl: './access.component.html',
  styleUrls: ['./access.component.sass'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AccessComponent extends CreateSubComponentBase implements AfterViewInit, OnDestroy {

  @ViewChild('stepInfo') infoTemplate: TemplateRef<any>
  @ViewChild('CBShortURL', { static: true }) shortURLElt: MatCheckbox

  form = this.formBuilder.group({
    Electorate: [0],
    Hidden: [false],
    ShortURL: [''],
  })

  get shortURLMinLength(): number {
    return 6
  }

  get shortURLIsTooShort(): boolean {
    const errors = this.form.get('ShortURL').errors
    return !!errors['minlength'] || !!errors['required']
  }

  get shortURLHasWrongFormat(): boolean {
    return !!this.form.get('ShortURL').errors['pattern']
  }

  get serverError(): ServerError {
    return this.service.serverError
  }

  constructor(
    public session: SessionService,
    protected service: CreateService,
    protected route: ActivatedRoute,
    private formBuilder: FormBuilder,
  ) {
    super();
  }

  // We use AfterViewInit instead of OnInit to ensure that shortURLElt is set when afterQueryFetch
  // is first called.
  ngAfterViewInit(): void {
    this.initModel();
  }

  ngOnDestroy(): void {
    this.unsubscribeAll();
  }

  protected afterQueryFetch(query: Partial<CreateQuery>): void {
    if (query.ShortURL) {
      this.shortURLElt.checked = true
      this.onCBShortURL({source: this.shortURLElt, checked: true})
    }
  }

  onCBShortURL(evt: MatCheckboxChange): void {
    const control = this.form.get('ShortURL')
    if (evt.checked) {
      control.setValidators([
        Validators.required,
        Validators.minLength(this.shortURLMinLength),
        Validators.pattern(/^[-_.~a-zA-Z0-9]*$/),
      ])
    } else {
      control.clearValidators()
    }
    control.updateValueAndValidity()
  }

  protected modifyQueryToSend(query: Partial<CreateQuery>): Partial<CreateQuery> {
    if (!this.shortURLElt?.checked) {
      query.ShortURL = undefined
    }
    return query
  }

}
