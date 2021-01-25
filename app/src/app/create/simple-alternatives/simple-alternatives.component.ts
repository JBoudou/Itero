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
import { trigger, transition, style, animate } from '@angular/animations';

import { Subject, Observable } from 'rxjs';

import { CdkDragDrop, moveItemInArray } from '@angular/cdk/drag-drop';

import { CreateService, CreateSubComponent } from '../create.service';
import { PollAlternative } from '../../api';


function duplicateValidator(component: SimpleAlternativesComponent): ValidatorFn {
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
  templateUrl: './simple-alternatives.component.html',
  styleUrls: [ './simple-alternatives.component.sass'],
  animations: [
    trigger('deleteTrigger', [
      transition('* => justDeleted', [
        style({
          opacity: 0,
          padding: 0,
          border: 'none',
          margin: 0,
        }),
        animate("250ms cubic-bezier(0, 0, 0.2, 1)", style({ height: '0px' }))
      ])
    ])
  ]
})
export class SimpleAlternativesComponent implements OnInit, CreateSubComponent {

  form = this.formBuilder.group({
    New: ['', [
      Validators.required,
      duplicateValidator(this),
    ]],
  });

  alternatives: PollAlternative[];
  justDeleted: number|undefined;

  readonly handledFields = new Set<string>(['Alternatives']);

  private _validable = new Subject<boolean>();
  get validable$(): Observable<boolean> { return this._validable; }

  isStarted(): boolean {
    return this.alternatives.length > 0;
  }

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
    this.justDeleted = pos;
  }

  onDeleteDone(): void {
    const pos = this.justDeleted;
    if (pos === undefined) {
      return
    }
    this.alternatives.splice(pos, 1);
    for (let i = pos, end = this.alternatives.length; i < end; i++) {
      this.alternatives[i].Id -= 1;
    }
    this.justDeleted = undefined;
    
    if (this.alternatives.length == 1) {
      this._validable.next(false);
    }
  }

}
