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

import { Component, ChangeDetectionStrategy, OnInit, OnDestroy, ViewChild, TemplateRef } from '@angular/core';
import { FormBuilder, Validators, AbstractControl, FormGroup, ValidationErrors } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';

import { CreateService } from '../create.service';
import { CreateSubComponentBase } from '../create-sub-component-base';

function customValidator(grp: FormGroup): ValidationErrors | null {
  let ret: ValidationErrors | null = null;

  // deadline is too early
  const diff = grp.value.Deadline.getTime() - Date.now();
  if (diff < 3600 * 1000) {
    ret = { earlyDeadline: true };
  }

  // min greater than max
  if (grp.value.MinNbRounds > grp.value.MaxNbRounds) {
    ret = ret === null ? {} : ret;
    ret.minMaxOrder = true;
  }

  return ret;
}

function integerValidator(control: AbstractControl): ValidationErrors | null {
  return Number.isInteger(control.value) ? null : { notInteger: true };
}

@Component({
  selector: 'app-create-round',
  templateUrl: './round.component.html',
  styleUrls: [ './round.component.sass'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class RoundComponent extends CreateSubComponentBase implements OnInit, OnDestroy {

  @ViewChild('stepInfo') infoTemplate: TemplateRef<any>;

  form = this.formBuilder.group({
    Deadline: [new Date(Date.now() + (7 * 24 * 3600 * 1000))],
    MinNbRounds: [2, [
      integerValidator,
      Validators.min(2),
      Validators.max(255),
    ]],
    MaxNbRounds: [10, [
      integerValidator,
      Validators.min(2),
      Validators.max(255),
    ]],
    MaxRoundDuration: [24*3600*1000, [
      integerValidator,
      Validators.min(60 * 1000),
    ]],
    ReportVote: [true],
  }, {
    validators: [ customValidator ],
  });

  constructor(
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

  minMax(): number {
    return this.form.value.MaxNbRounds;
  }

  maxMin(): number {
    return this.form.value.MinNbRounds;
  }

  tooEarlyDeadline(): boolean {
    return this.form.controls['Deadline'].dirty &&
          !this.form.valid &&
         !!this.form.errors?.['earlyDeadline'];
  }

  wrongInterval(): boolean {
    return (this.form.controls['MinNbRounds'].dirty ||  this.form.controls['MaxNbRounds'].dirty) && !this.form.valid &&
          (!this.form.controls['MinNbRounds'].valid || !this.form.controls['MaxNbRounds'].valid  ||
            ( !!this.form.errors && !!this.form.errors['minMaxOrder'] )
          );
  }

  tooFewRounds(): boolean {
    return this.form.controls['MinNbRounds'].dirty &&
          !this.form.controls['MinNbRounds'].valid &&
         !!this.form.controls['MinNbRounds'].errors['min'];
  }

  wrongDuration(): boolean {
    return this.form.controls['MaxRoundDuration'].dirty &&
          !this.form.controls['MaxRoundDuration'].valid &&
         !!this.form.controls['MaxRoundDuration'].errors['notInteger'];
  }

  tooShortDuration(): boolean {
    return this.form.controls['MaxRoundDuration'].dirty &&
          !this.form.controls['MaxRoundDuration'].valid &&
         !!this.form.controls['MaxRoundDuration'].errors['min'];
  }

}
