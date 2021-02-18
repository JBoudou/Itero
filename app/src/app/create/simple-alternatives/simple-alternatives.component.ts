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
  FormArray,
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

// TODO delete
class ErrorStateFromObservable implements ErrorStateMatcher {

  private lastState: boolean = false;
  private subscription: Subscription;

  constructor(source: Observable<boolean>) {
    this.subscription = source.subscribe({
      next: (state: boolean) => this.lastState = state,
    });
  }

  destroy(): void {
    this.subscription.unsubscribe();
  }

  isErrorState(control: FormControl|null, form: FormGroupDirective|NgForm|null): boolean {
    return this.lastState;
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
    return this.Alternatives.statusChanges.pipe(map(val => val == 'VALID'), startWith(this.Alternatives.valid));
  }


  /* Interface for the template */

  form = this.formBuilder.group({
    Alternatives: this.formBuilder.array([], [
      Validators.required,
      Validators.minLength(2),
      noDuplicateNames,
    ]),
    New: ['', [
      Validators.required,
      this.newValidator.bind(this),
    ]],
  });

  get Alternatives(): FormArray {
    return this.form?.get('Alternatives') as FormArray;
  }
  
  errorStateMatcher: ErrorStateFromObservable;
  
  justDeleted: number = -1;

  newIsDuplicate(): boolean {
    return !this.form.controls['New'].valid &&
            this.form.controls['New'].errors['existingNew'] !== undefined;
  }

  hasDuplicate(): boolean {
    return !this.Alternatives.valid &&
            this.Alternatives.errors['duplicateNames'] !== undefined;
  }

  tooFewAlternatives(): boolean {
    return !this.Alternatives.valid &&
           (this.Alternatives.errors['minlength'] !== undefined ||
            this.Alternatives.errors['required' ] !== undefined);
  }

  onAdd(): void {
    console.log('on add ' + this.form.value.New);
    this.addAlternative(this.form.value.New);
    this.form.patchValue({New: ''});
  }

  onDelete(pos: number): void {
    console.log('on delete ' + pos);
    this.justDeleted = pos;
  }

  onDeleteDone(): void {
    const pos = this.justDeleted;
    if (pos < 0) {
      return
    }
    console.log('delete ' + pos);
    this.justDeleted = -1;
    this.Alternatives.removeAt(pos);
  }


  /* Initialisation / Destruction */

  private _subscriptions: Subscription[] = [];

  get alternativesUpdates$(): Observable<SimpleAlternative[]> {
    return this.Alternatives.valueChanges.pipe(filter(this.filterEvent, this));
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
    private changeDetector: ChangeDetectorRef,
  ){
    this.errorStateMatcher = new ErrorStateFromObservable(this.validable$.pipe(map((state: boolean) => !state)));
  }

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
    );
  }

  private _initModel(stepSegment: string): void {
    this._subscriptions.push(
      this.alternativesUpdates$.subscribe({
        next: (alternatives: SimpleAlternative[]) => {
          // TODO remove
          console.log('Send ' + JSON.stringify(alternatives));
          this.service.patchQuery(stepSegment, { Alternatives: alternatives });
        }
      }),
    );
  }

  private synchronizeFromService(query: Partial<CreateQuery>): void {
    // TODO remove debug
    console.log('Receive ' + JSON.stringify(query.Alternatives));
    if (query.Alternatives === undefined || !this.synchronizeAlternatives(query.Alternatives)) {
      return;
    }
  }

  ngOnDestroy(): void {
    this._subscriptions.forEach(sub => sub.unsubscribe());
    this.errorStateMatcher.destroy();
  }


  /* Validations */

  private newValidator(control: AbstractControl): ValidationErrors | null {
    if (this.Alternatives === undefined) return null;
    const val = control.value;
    for (let alt of this.Alternatives.value) {
      if (alt.Name === val) {
        return { existingNew: true };
      }
    }
    return null;
  }


  /* Model handling */

  private addAlternative(name: string, cost?: number): void {
    cost = cost || 1;
    this.Alternatives.push(this.formBuilder.group({
      Name: name,
      Cost: cost,
    }));
  }

  /**
   * Enforce the model of the alternatives to match the given list.
   * Return whether the model has been changed.
   */
  private synchronizeAlternatives(fromService: SimpleAlternative[]): boolean {
    let ret: boolean = false;
    const model = this.Alternatives;
    const serviceLen = fromService.length;

    this._filteringEvents = true;

    while (model.controls.length > serviceLen) {
      console.log('sync remove');
      model.removeAt(serviceLen);
      ret = true;
    }
    const endCheck = model.controls.length;
    while (model.controls.length < serviceLen) {
      console.log('sync add');
      const alt = fromService[model.controls.length];
      this.addAlternative(alt.Name, alt.Cost);
      ret = true;
    }

    for (let i = 0; i < endCheck; i++) {
      const group = model.controls[i] as FormGroup;
      const alt = fromService[i];
      if (group.controls['Name'].value !== alt.Name ||
          group.controls['Cost'].value !== alt.Cost) {
        console.log('sync patch');
        group.patchValue(alt, { onlySelf: false, emitEvent: false });
        ret = true;
      }
    }

    this._filteringEvents = false;
    if (ret) {
      console.log('sync updt');
      model.updateValueAndValidity();
    }

    return ret;
  }

}
