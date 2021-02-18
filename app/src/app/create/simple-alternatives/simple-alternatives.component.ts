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
  ValidatorFn,
  Validators,
} from '@angular/forms';

import { trigger, transition, style, animate } from '@angular/animations';
import { ActivatedRoute, UrlSegment } from '@angular/router';

import { BehaviorSubject, Observable, Subscription } from 'rxjs';
import { filter, map, take } from 'rxjs/operators';

import { ErrorStateMatcher } from '@angular/material/core';

import { CreateService } from '../create.service';
import { PollAlternative, SimpleAlternative, CreateQuery } from '../../api';

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

function duplicateValidator(component: SimpleAlternativesComponent): ValidatorFn {
  return (control: AbstractControl): ValidationErrors | null => {
    if (component.alternatives === undefined) {
      return { undefined: true }
    }
    for (let alt of component.alternatives) {
      if (alt.Name == control.value) {
        return { duplicatedAlternative: true }
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
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class SimpleAlternativesComponent implements OnInit, OnDestroy {

  @ViewChild('stepInfo') infoTemplate: TemplateRef<any>;

  form = this.formBuilder.group({
    Alternatives: this.formBuilder.array([]),
    New: ['', [
      Validators.required,
      duplicateValidator(this),
    ]],
  });

  // TODO: Remove alternatives and rename Alternatives
  alternatives: PollAlternative[] = [];

  get Alternatives(): FormArray {
    return this.form.get('Alternatives') as FormArray;
  }

  private _stepSegment: string;
  private _subscriptions: Subscription[] = [];
  
  justDeleted: number = -1;

  private _validable$ = new BehaviorSubject<boolean>(false);
  get validable$(): Observable<boolean> { return this._validable$; }

  get alternativesUpdates$(): Observable<SimpleAlternative[]> {
    return this.Alternatives.valueChanges.pipe(filter(this.filterEvent, this));
  }

  
  errorStateMatcher: ErrorStateFromObservable;

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

  ngOnDestroy(): void {
    this._subscriptions.forEach(sub => sub.unsubscribe());
    this.errorStateMatcher.destroy();
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
    // TODO remove everything.
    this._validable$.next(this.alternatives.length >= 2);
  }

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

  // What follows is needed because some FormArray methods do not have the option to disable event
  // sending. These has just been merged into Angular https://github.com/angular/angular/pull/31031.
  private _filteringEvents: boolean = false;
  private filterEvent(value: any, index: number): boolean {
    return !this._filteringEvents;
  }

  hasDuplicate(): boolean {
    return !this.form.controls['New'].valid &&
            this.form.controls['New'].errors['duplicatedAlternative'] !== undefined;
  }

  onAdd(): void {
    this.addAlternative(this.form.value.New);
    this.form.patchValue({New: ''});
    
    if (this.Alternatives.length == 2) {
      this._validable$.next(true);
    }
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
    if (this.Alternatives.length == 1) {
      this._validable$.next(false);
    }
  }

}
