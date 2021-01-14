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
import { FormBuilder, Validators, AbstractControl, ValidatorFn, ValidationErrors  } from '@angular/forms';

import { Subject, Observable } from 'rxjs';

import { CdkDragDrop, moveItemInArray } from '@angular/cdk/drag-drop';

import { CreateService, CreateSubComponent } from '../create/create.service';
import { PollAlternative } from '../api';


function duplicateValidator(component: CreateSimpleAlternativesComponent): ValidatorFn {
  return (control: AbstractControl): ValidationErrors | null => {
    if (component.alternatives === undefined) {
      return { undefined: true }
    }
    for (let alt of component.alternatives) {
      if (alt.Name == control.value) {
        return { duplicatedAlternative: alt.Id }
      }
    }
    return null
  }
}

@Component({
  selector: 'app-create-simple-alternatives',
  templateUrl: './create-simple-alternatives.component.html',
  styleUrls: ['./create-simple-alternatives.component.sass']
})
export class CreateSimpleAlternativesComponent implements OnInit, CreateSubComponent {

  form = this.formBuilder.group({
    New: ['', [
      Validators.required,
      duplicateValidator(this),
    ]],
  });

  alternatives: PollAlternative[];

  readonly handledFields = new Set<string>(['Alternatives']);

  private _validable = new Subject<boolean>();
  get validable$(): Observable<boolean> { return this._validable; }

  constructor(
    private service: CreateService,
    private formBuilder: FormBuilder,
  ){ }

  ngOnInit(): void {
    const query = this.service.register(this);
    if (query.Alternatives === undefined) {
      query.Alternatives = [];
    }
    this.alternatives = query.Alternatives;
    this._validable.next(this.alternatives.length >= 2);
  }

  hasDuplicate(): boolean {
    return !this.form.controls['New'].valid &&
            this.form.controls['New'].errors['duplicatedAlternative'] !== undefined;
  }

  onAdd(): void {
    this.alternatives.push({
      Id: this.alternatives.length,
      Name: this.form.value.New,
      Cost: 1,
    });
    this.form.patchValue({New: ''});
    
    if (this.alternatives.length == 2) {
      this._validable.next(true);
    }
  }

  onDrop(event: CdkDragDrop<string[]>): void {
    moveItemInArray(this.alternatives, event.previousIndex, event.currentIndex);
    for (let i = Math.min(event.previousIndex, event.currentIndex),
          last = Math.max(event.previousIndex, event.currentIndex); i <= last; i++) {
      this.alternatives[i].Id = i;
    }
  }

  onDelete(pos: number): void {
    this.alternatives.splice(pos, 1);
    for (let i = pos, end = this.alternatives.length; i < end; i++) {
      this.alternatives[i].Id -= 1;
    }
  }

}
