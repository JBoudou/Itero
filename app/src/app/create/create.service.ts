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

import { Inject, Injectable, InjectionToken } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Router } from '@angular/router';

import { Observable, ReplaySubject, Subscription } from 'rxjs';
import { cloneDeep } from 'lodash';

import { CreateQuery } from '../api';
import { NavTreeNode, LinearNavTreeNode, FinalNavTreeNode } from './navtree/navtree.node';
import { NavStepStatus } from './navtree/navstep.status';


export interface CreateSubComponent {
  /** Name of query's properties that may be modified by the component. */
  readonly handledFields: Set<string>;
  readonly validable$: Observable<boolean>;
  isStarted(): boolean;
}

/** Event about the next() action of CreateService. */
export class CreateNextStatus {
  constructor(
    /** Whether calling next() is possible. */
    public validable: boolean,
    /** Whether calling next() will sends the request to the middleware. */
    public final: boolean
  ) {};
}

export const CREATE_TREE = new InjectionToken<NavTreeNode>('create.tree');

export const APP_CREATE_TREE: NavTreeNode =
  new LinearNavTreeNode('general', 'Generalities',
    new LinearNavTreeNode('simpleAlternatives', 'Alternatives',
      new FinalNavTreeNode('round', 'Rounds')
    )
  );


/**
 * CreateService is the central service for the main creation component and all its subcomponents.
 * It handles the navigation across the subcomponents, the models for them and sends the final
 * request to the middleware.
 *
 * It informs the main component about the current state of the creation (see createStepStatus$ and
 * createNextStatus$) and provides the back() and next() action for it.
 *
 * Subcomponents must call register() when they're displayed.
 */
@Injectable()
export class CreateService {

  private _createNext = new ReplaySubject<CreateNextStatus>(1);
  private _createStep = new ReplaySubject<NavStepStatus>(1);

  get createNextStatus$(): Observable<CreateNextStatus> {
    return this._createNext;
  }

  get createStepStatus$(): Observable<NavStepStatus> {
    return this._createStep;
  }

  private _httpError: HttpErrorResponse;

  private current: NavTreeNode;
  private subComponent: CreateSubComponent|undefined;
  private subComponentSubscription: Subscription|undefined;

  /**
   * Whether the service is in "sending state".
   * The service is in "sending state" in the meantime between the validation of the creation by the
   * user and the moment the result is displayed to the user.
   */
  private _sending: boolean = false;

  constructor(
    @Inject(CREATE_TREE) private root: NavTreeNode,
    private router: Router,
    private http: HttpClient,
  ) {
    this.current = this.root;
  }

  /**
   * Asks the router to go back in the creation tree.
   *
   * Must not be called when the current node is the root node, which is notified by a
   * NavStepStatus with current at zero.
   */
  back(steps?: number): void {
    if (steps === undefined) {
      steps = 1;
    }

    let current = this.current;
    for (let i = 0; i < steps; i++) {
      const parent = current.parent;
      if (parent === undefined) {
        console.warn('CreateService back on the root node !')
        break;
      }
      parent.handledFields = new Set<string>();
      current = parent
    }

    this.makeCurrent(current);
  }

  /**
   * Terminates the current step.
   *
   * Must not be called when the current is not validable, wich is notified by a CreateNextStatus
   * with validable at false.
   *
   * If the current step is a final one, which is notified by a CreateNextStatus with final at true,
   * the request is sent to the middleware and the creation tree is reseted.
   * Otherwise the router is asked to display the next step.
   */
  next(): void {
    if (this.subComponent === undefined) {
      console.warn('CreateService next without component !')
    }

    if (this.current.isFinal) {
      this.sendRequest();
      return;
    }

    const next = this.current.next();
    next.handledFields = new Set<string>();
    for (const prop of this.current.handledFields) {
      next.handledFields.add(prop);
      next.query[prop] = cloneDeep(this.current.query[prop]);
    }
    this.makeCurrent(next);
  }

  /**
   * Notify the CreateService that a subcomponent is now displayed.
   * This method must be called by subcomponent each time they're displayed.
   * The returned object can (and should) be modified by the subcomponent.
   */
  register(subComp: CreateSubComponent): Partial<CreateQuery> {
    if (this._sending) {
      console.warn("Registering while sending!");
      this.reset();
    }
    if (this.navigateToCurrent()) {
      return {};
    }

    // Since the subcomponent registers in some initialisation hook,
    // and the notification may be directed to a parent of the subcomponent,
    // there are high risks of ExpressionChangedAfterItHasBeenCheckedError.
    // Hence, we deliver notifications asynchronously.
    // Use of Angular's EventEmitter(true) may be more appropriate than setTimeout() (but heavier).

    setTimeout(() => {
      this._createNext.next(new CreateNextStatus(false, this.current.isFinal)),
      this._createStep.next(this.current.makeStatus());
    });

    this.subComponent = subComp;
    this.subComponent.validable$.subscribe({
      next: (validable: boolean) => {
        setTimeout(() =>
          this._createNext.next(new CreateNextStatus(validable, this.current.isFinal))
        );
      },
    });

    for (const prop of this.subComponent.handledFields) {
      this.current.handledFields.add(prop);
    }
    return this.current.query;
  }

  /** Whether the user already set some values. */
  canLeave(): boolean {
    return this._sending ||
      (this.current === this.root && (!this.subComponent || !this.subComponent.isStarted()));
  }

  /**
   * Get the result of sending the create request.
   * Call to this method reinitialise the service.
   */
  getResult(): HttpErrorResponse | Partial<CreateQuery> | undefined {
    if (!this._sending) {
      return undefined;
    }
    const ret = !!this ._httpError ? this._httpError : this.current.query;
    this.reset();
    return ret;
  }

  reset() {
    this.root.reset();
    this._sending = false;
    this._httpError = undefined;
    this.current = this.root;
  }

  private makeCurrent(node: NavTreeNode|undefined): void {
    if (node === undefined) {
      console.warn('CreateService undefined current');
      return;
    }
    this.current = node;
    this.navigateToCurrent();
  }
  
  /**
   * Check whether the current route ends with the right segment, and navigate to it if it's not.
   * \returns Whether navigation occured.
   */
  private navigateToCurrent(): boolean {
    const url = this.router.routerState.snapshot.url;
    if (url.endsWith('/' + this.current.segment)) {
      return false;
    }

    if (this.subComponentSubscription !== undefined) {
      this.subComponentSubscription.unsubscribe();
      this.subComponentSubscription = undefined;
    }
    this.subComponent = undefined;

    this.router.navigate(['/r/create/' + this.current.segment]);
    return true;
  }

  private sendRequest(): void {
    this._sending = true;
    this.http.post<string>('/a/create', this.current.query, { observe: 'body', responseType: 'json' })
      .subscribe({
      next: (segment: string) => {
        this._httpError = undefined;
        this.router.navigateByUrl('/r/create/result/' + segment);
      },
      error: (err: HttpErrorResponse) => {
        this._httpError = err;
        this.router.navigateByUrl('/r/create/result/error');
      },
    });
  }
}
