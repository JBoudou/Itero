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

import { Subject, Observable } from 'rxjs';

import { CdkDragDrop, moveItemInArray } from '@angular/cdk/drag-drop';

import { CreateService, CreateSubComponent } from '../create/create.service';
import { PollAlternative } from '../api';

@Component({
  selector: 'app-create-simple-alternatives',
  templateUrl: './create-simple-alternatives.component.html',
  styleUrls: ['./create-simple-alternatives.component.sass']
})
export class CreateSimpleAlternativesComponent implements OnInit, CreateSubComponent {

  form = this.formBuilder.group({
    New: ['', [
      Validators.required,
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

  onAdd(): void {
    this.alternatives.push({
      Id: this.alternatives.length,
      Name: this.form.value.New,
      Cost: 1,
    });
    
    if (this.alternatives.length == 2) {
      this._validable.next(true);
    }
  }

  onDrop(event: CdkDragDrop<string[]>) {
    moveItemInArray(this.alternatives, event.previousIndex, event.currentIndex);
    for (let i = Math.min(event.previousIndex, event.currentIndex),
          last = Math.max(event.previousIndex, event.currentIndex); i <= last; i++) {
      this.alternatives[i].Id = i;
    }
  }

}
