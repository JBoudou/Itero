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

import {
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Component,
  OnDestroy,
  OnInit,
  TemplateRef,
  ViewChild,
} from '@angular/core';

import {
  AbstractControl,
  FormBuilder,
  FormControl,
  FormGroup,
  FormGroupDirective,
  NgForm,
  ValidationErrors,
  Validators,
} from '@angular/forms';

import { trigger, transition, style, animate } from '@angular/animations';
import { ActivatedRoute, UrlSegment } from '@angular/router';

import { Observable, Subscription } from 'rxjs';
import { filter, map, startWith, take } from 'rxjs/operators';

import { ErrorStateMatcher } from '@angular/material/core';

import { CreateService } from '../create.service';
import { SimpleAlternative, CreateQuery } from '../../api';

class ErrorStateNotRequired implements ErrorStateMatcher {
  isErrorState(control: FormControl|null, form: FormGroupDirective|NgForm|null): boolean {
    return !!(control && control.invalid && control.errors &&
              Object.keys(control.errors).some((prop: string) => prop !== 'required'));
  }
}

function noDuplicateNames(control: AbstractControl): ValidationErrors | null {
  let set = new Set<string>();
  for (let alt of control.value) {
    if (set.has(alt.Name)) {
      return { duplicateNames: alt.Name };
    }
    set.add(alt.Name);
  }
  return null;
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
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class SimpleAlternativesComponent implements OnInit, OnDestroy {

  /* Interface for parent components */

  @ViewChild('stepInfo') infoTemplate: TemplateRef<any>;

  get validable$(): Observable<boolean> {
    return this.altForm.statusChanges.pipe(map(val => val == 'VALID'), startWith(this.altForm.valid));
  }


  /* Interface for the template */

  altForm = this.formBuilder.array([], [
    Validators.required,
    Validators.minLength(2),
    noDuplicateNames,
  ]);
  newForm = this.formBuilder.control(
    '', [
      Validators.required,
      this.newValidator.bind(this),
    ]
  );

  justDeleted: number = -1;

  newErrorState = new ErrorStateNotRequired();

  hasErrors(): boolean {
    return !this.altForm.valid || this.newErrorState.isErrorState(this.newForm, null);
  }

  newIsDuplicate(): boolean {
    return !this.newForm.valid &&
            this.newForm.errors['existingNew'] !== undefined;
  }

  hasDuplicate(): boolean {
    return !this.altForm.valid && !!this.altForm.errors &&
            this.altForm.errors['duplicateNames'] !== undefined;
  }

  tooFewAlternatives(): boolean {
    return !this.altForm.valid && !!this.altForm.errors &&
           (this.altForm.errors['minlength'] !== undefined ||
            this.altForm.errors['required' ] !== undefined);
  }

  emptyAlternative(): boolean {
    const check = function (control: FormGroup): boolean {
      const name = control.get('Name');
      return !!(name && name.errors && name.errors['required']);
    }
    return !this.altForm.valid &&
            this.altForm.controls.some(check);
  }
                                       

  onAdd(): void {
    if (!this.newForm.valid) return;
    this.addAlternative(this.newForm.value);
    this.newForm.reset();
  }

  onDelete(pos: number): void {
    this.justDeleted = pos;
  }

  onDeleteDone(): void {
    const pos = this.justDeleted;
    if (pos < 0) {
      return
    }
    this.justDeleted = -1;
    this.altForm.removeAt(pos);
  }


  /* Initialisation / Destruction */

  private _subscriptions: Subscription[] = [];

  // For testing purpose
  get alternativesUpdates$(): Observable<SimpleAlternative[]> {
    return this.altForm.valueChanges.pipe(filter(this.filterEvent, this));
  }

  // What follows is needed because some FormArray methods do not have the option to disable event
  // sending. These has just been merged into Angular https://github.com/angular/angular/pull/31031.
  private _filteringEvents: boolean = false;
  private filterEvent(value: any, index: number): boolean {
    return !this._filteringEvents;
  }

  constructor(
    private service: CreateService,
    private route: ActivatedRoute,
    private formBuilder: FormBuilder,
  ){ }

  ngOnInit(): void {
    this.route.url.pipe(take(1)).subscribe({
      next: (segments: UrlSegment[]) => {
        const stepSegment = segments[segments.length - 1].toString();
        this._initModel(stepSegment);
      }
    });

    this._subscriptions.push(
      this.service.query$.subscribe({
        next: (query: Partial<CreateQuery>) => this.synchronizeFromService(query),
      }),
      this.altForm.valueChanges.subscribe({
        next: () => this.newForm.updateValueAndValidity()
      }),
    );
  }

  private _initModel(stepSegment: string): void {
    this._subscriptions.push(
      this.alternativesUpdates$.subscribe({
        next: (alternatives: SimpleAlternative[]) => {
          this.service.patchQuery(stepSegment, { Alternatives: alternatives });
        }
      }),
    );
  }

  private synchronizeFromService(query: Partial<CreateQuery>): void {
    if (query.Alternatives === undefined || !this.synchronizeAlternatives(query.Alternatives)) {
      return;
    }
  }

  ngOnDestroy(): void {
    this._subscriptions.forEach(sub => sub.unsubscribe());
  }


  /* Validations */

  private newValidator(control: AbstractControl): ValidationErrors | null {
    if (this.altForm === undefined) return null;
    const val = control.value;
    for (let alt of this.altForm.value) {
      if (alt.Name === val) {
        return { existingNew: true };
      }
    }
    return null;
  }


  /* Model handling */

  private addAlternative(name: string, cost?: number): void {
    cost = cost || 1;
    this.altForm.push(this.formBuilder.group({
      Name: [name, [ Validators.required ]],
      Cost: cost,
    }));
  }

  /**
   * Enforce the model of the alternatives to match the given list.
   * Return whether the model has been changed.
   */
  private synchronizeAlternatives(fromService: SimpleAlternative[]): boolean {
    let ret: boolean = false;
    const model = this.altForm;
    const serviceLen = fromService.length;

    this._filteringEvents = true;

    while (model.controls.length > serviceLen) {
      model.removeAt(serviceLen);
      ret = true;
    }
    const endCheck = model.controls.length;
    while (model.controls.length < serviceLen) {
      const alt = fromService[model.controls.length];
      this.addAlternative(alt.Name, alt.Cost);
      ret = true;
    }

    for (let i = 0; i < endCheck; i++) {
      const group = model.controls[i] as FormGroup;
      const alt = fromService[i];
      if (group.controls['Name'].value !== alt.Name ||
          group.controls['Cost'].value !== alt.Cost) {
        group.patchValue(alt, { onlySelf: false, emitEvent: false });
        ret = true;
      }
    }

    this._filteringEvents = false;
    if (ret) {
      model.updateValueAndValidity();
    }

    return ret;
  }

}
