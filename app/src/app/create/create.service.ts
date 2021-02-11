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

import { Observable, BehaviorSubject } from 'rxjs';
import { cloneDeep, isEqual } from 'lodash';

import { CreateQuery } from '../api';
import { NavTreeNode, LinearNavTreeNode, FinalNavTreeNode } from './navtree/navtree.node';
import { NavStepStatus } from './navtree/navstep.status';


export const CREATE_TREE = new InjectionToken<NavTreeNode>('create.tree');

export const APP_CREATE_TREE: NavTreeNode =
  new LinearNavTreeNode('general', 'Generalities',
    new LinearNavTreeNode('simpleAlternatives', 'Alternatives',
      new FinalNavTreeNode('round', 'Rounds')
    )
  );


/**
 * CreateService is the central service for the creation of polls.
 * It handles the navigation across the subcomponents, the models for them and sends the final
 * request to the middleware.
 *
 * It informs the main component about the current state of the creation (see stepStatus$) and
 * provides the back() and next() action for it.
 */
@Injectable()
export class CreateService {

  private _current: NavTreeNode;

  private _stepStatus$ = new BehaviorSubject<NavStepStatus>(undefined);

  get stepStatus$(): Observable<NavStepStatus> {
    return this._stepStatus$;
  }

  private _query$ = new BehaviorSubject<Partial<CreateQuery>>({});
  private _queryModified: boolean;

  /**
   * The current partial query.
   * The sent objects must never be modified directly. Use patchQuery() instead.
   */
  get query$(): Observable<Partial<CreateQuery>> {
    return this._query$;
  }

  /**
   * Whether the service is in "sending state".
   * The service is in "sending state" in the meantime between the validation of the creation by the
   * user and the moment the result is displayed to the user.
   */
  private _sending: boolean = false;

  private _httpError: HttpErrorResponse;

  constructor(
    @Inject(CREATE_TREE) private root: NavTreeNode,
    private router: Router,
    private http: HttpClient,
  ) {
    this.makeCurrent(this.root, { navigate: false });
  }


  // Interface for the layout component //

  /**
   * Asks the router to go back in the creation tree.
   *
   * Must not be called when the current node is the root node, which is notified by a
   * StepStatus with current at zero.
   */
  back(steps: number = 1): void {
    let current = this._current;
    for (let i = 0; i < steps; i++) {
      const parent = current.parent;
      if (parent === undefined) {
        console.warn('CreateService back on the root node !')
        break;
      }
      for (const prop in current.query) {
        if (!current.handledFields.has(prop)) {
          current.query[prop] = undefined;
        }
      }
      current = parent
    }

    this.makeCurrent(current);
  }

  /**
   * Terminates the current step.
   *
   * If the current step is a final one, the request is sent to the middleware and the creation tree
   * is reseted.
   * Otherwise the router is asked to display the next step.
   */
  next(): void {
    if (this._current.isFinal) {
      this.sendRequest();
      return;
    }

    const next = this._current.next();
    for (const prop in this._current.query) {
      next.handledFields.delete(prop);
      next.query[prop] = cloneDeep(this._current.query[prop]);
    }
    
    this.makeCurrent(next);
  }


  // Interface for editing components //

  /**
   * Modifies some fields of the current query.
   * 
   * This method is to be called by components editing the parameters of the poll to be created.
   * Beware that, since this method modifies the query, it usually sends an event on query$.
   * The query is marked as modified, unless defaultValues option is set to true.
   */
  patchQuery(stepSegment: string, patch: Partial<CreateQuery>,
             options: { defaultValues: boolean } = { defaultValues: false}
  ): boolean {
    if (stepSegment != this._current.segment) {
      console.warn('CreateService patch from wrong step');
      return false;
    }

    let modified = false;
    for (const prop in patch) {
      if (!isEqual(patch[prop], this._current.query[prop])) {
        this._current.query[prop] = cloneDeep(patch[prop]);
        this._current.handledFields.add(prop);
        modified = true;
      }
    }
    if (modified) {
      if (!options.defaultValues) {
        this._queryModified = true;
      }
      this._query$.next(this._current.query);
    }
    return true;
  }


  // Interface for guards //

  currentUrl(): string {
    return '/r/create/' + this._current.segment;
  }

  /** Whether the user can leave the create section. */
  canLeave(): boolean {
    return this._sending || !this._queryModified;
  }

  reset() {
    this.makeCurrent(this.root, { reset: true, navigate: false });
  }


  // Interface for result page //

  /**
   * Get the result of sending the create request.
   * Call to this method reinitialise the service.
   */
  getResult(): HttpErrorResponse | Partial<CreateQuery> | undefined {
    if (!this._sending) {
      return undefined;
    }
    const ret = !!this ._httpError ? this._httpError : this._current.query;
    this.reset();
    return ret;
  }


  // Private methods //

  private makeCurrent(
    node: NavTreeNode,
    options: { reset?: boolean, navigate?: boolean } = { navigate: true }
  ): void {
    if (node === undefined) {
      console.warn('CreateService undefined current');
      return;
    }
    if (!options.reset && this._sending) {
      console.warn('CreateService change root while current');
      options.reset = true;
    }

    if (options.reset) {
      this._sending = false;
      this._httpError = undefined;
      this.root.reset();
      this._queryModified = false;
    }

    this._current = node;
    if (options.navigate) {
      this.navigateToCurrent();
    }

    this._stepStatus$.next(this._current.makeStatus());
    // The query must be sent later, such that the correct component receives it.
    setTimeout(() => this._query$.next(this._current.query), 0);
  }
  
  /**
   * Check whether the current route ends with the right segment, and navigate to it if it's not.
   * \returns Whether navigation occured.
   */
  private navigateToCurrent(): boolean {
    const url = this.router.routerState.snapshot.url;
    if (url.endsWith('/' + this._current.segment)) {
      return false;
    }

    this.router.navigate([this.currentUrl()]);
    return true;
  }

  private sendRequest(): void {
    this._sending = true;
    this.http.post<string>('/a/create', this._current.query, { observe: 'body', responseType: 'json' })
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
