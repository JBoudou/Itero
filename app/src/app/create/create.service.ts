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
import { Event, NavigationEnd, Router } from '@angular/router';

import { Observable, BehaviorSubject } from 'rxjs';
import { take, delayWhen, filter } from 'rxjs/operators';
import { cloneDeep, isEqual } from 'lodash';

import { CreateQuery } from '../api';
import { NavTreeNode, LinearNavTreeNode, FinalNavTreeNode } from './navtree/navtree.node';
import { NavStepStatus } from './navtree/navstep.status';
import { ServerError } from '../shared/server-error';


export const CREATE_TREE = new InjectionToken<NavTreeNode>('create.tree');

export const APP_CREATE_TREE: NavTreeNode =
  new LinearNavTreeNode('general', 'Generalities',
    new LinearNavTreeNode('simpleAlternatives', 'Alternatives',
      new LinearNavTreeNode('round', 'Rounds',
        new FinalNavTreeNode('access', 'Access')
      )
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
    // see _allowQuery$ below for explanations.
    return this._query$.pipe(delayWhen<Partial<CreateQuery>>(() => this._allowQuery$.pipe(filter((b: boolean) => b))));
  }

  // We use delayWhen along with _allowQuery$ to suspend and resume query$ events. See
  // suspendQuery() and resumeQuery().
  // This solution is not very efficient and rxjs does not guarantee that the blocked events will be
  // sent in the order they were emitted. A better solution would be to implement our own Operator.
  private _allowQuery$ = new BehaviorSubject<boolean>(true);

  /**
   * Prevent query$ to send next events until resumeQuery() is called.
   * Blocked events are sent when resumeQuery().
   */
  private suspendQuery(): void {
    this._allowQuery$.next(false);
  }

  /**
   * Stop blocking query$ to send next events.
   * Blocked events are sent right now.
   */
  private resumeQuery(): void {
    this._allowQuery$.next(true);
  }

  /**
   * Whether the service is in "sending state".
   * The service is in "sending state" in the meantime between the validation of the creation by the
   * user and the moment the result is displayed to the user.
   */
  private _sending: boolean = false;

  serverError: ServerError = new ServerError()

  constructor(
    @Inject(CREATE_TREE) private root: NavTreeNode,
    private router: Router,
    private http: HttpClient,
  ) {
    this.makeCurrent(this.root, { navigate: false });

    this.router.events
      .pipe(filter<Event, NavigationEnd>((evt: Event): evt is NavigationEnd => evt instanceof NavigationEnd))
      .subscribe({
        next: (evt: NavigationEnd) => {
          const lastSegment = evt.urlAfterRedirects.split('/').pop()
          const node = this.root.findSegment(lastSegment)
          if (node !== null) {
            this.makeCurrent(node, { navigate: false })
          }
        },
      })
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
   * To delete a value set it to undefined.
   */
  patchQuery(stepSegment: string, patch: Partial<CreateQuery>,
             options: { defaultValues: boolean } = { defaultValues: false}
  ): boolean {
    // When called from the wrong segment,
    // we still check whether the patch changes any value.
    // If it does, warning messages are displayed.
    const ret = stepSegment == this._current.segment;

    let modified = false;
    for (const prop in patch) {
      if (isEqual(patch[prop], this._current.query[prop])) { continue }

      // Warn in case of wrong segment.
      if (!ret) {
        console.warn(`Segment ${stepSegment} instead of ${this._current.segment} ` +
                      `to change ${prop} from ${this._current.query[prop]} to ${patch[prop]}`);
        continue;
      }
      
      // Update the query
      if (patch[prop] === undefined) {
        delete this._current.query[prop];
        this._current.handledFields.delete(prop);
      } else {
        this._current.query[prop] = cloneDeep(patch[prop]);
        this._current.handledFields.add(prop);
      }

      // Remove error
      if (!this.serverError.ok && this.serverError.status === 409 &&
          this.serverError.message.split(' ')[0] === prop) {
        this.serverError = new ServerError()
      }

      modified = true;
    }

    // Send the new query
    if (ret && modified) {
      if (!options.defaultValues) {
        this._queryModified = true;
      }
      this._query$.next(this._current.query);
    }
    return ret;
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
  getResult(): ServerError | Partial<CreateQuery> | undefined {
    if (!this._sending) {
      return undefined;
    }
    const ret = !this .serverError.ok ? this.serverError : this._current.query;
    this.reset();
    return ret;
  }


  // Private methods //

  /**
   * Set the current step.
   * Options:
   *  - `reset` set to true to reset the service while change the step.
   *  - `navigate` set to false to prevent changing the view.
   */
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
    if (!options.reset && !options.navigate && this._current === node) {
      return
    }

    if (options.reset) {
      this._sending = false;
      this.serverError = new ServerError()
      this.root.reset();
      this._queryModified = false;
    }

    // Sending queries is suspended while the component is changed.
    // That way, the previous component never receives the query of the next component,
    // and the next component never receives the query of the previous component.
    if (options.navigate) {
      this.suspendQuery();
    }
    this._current = node;
    this._query$.next(this._current.query);
    if (options.navigate) {
      this.navigateToCurrent().then(() => this.resumeQuery());
    }

    this._stepStatus$.next(this._current.makeStatus());
  }

  /**
   * Check whether the current route ends with the right segment, and navigate to it if it's not.
   * \returns Whether navigation occured.
   */
  private navigateToCurrent(): Promise<boolean> {
    const url = this.router.routerState.snapshot.url;
    if (url.endsWith('/' + this._current.segment)) {
      return Promise.resolve(false);
    }

    return this.router.navigate([this.currentUrl()]);
  }

  private sendRequest(): void {
    this._sending = true;
    console.log(JSON.stringify(this._current.query))
    this.http.post<string>('/a/create', this._current.query, { observe: 'body', responseType: 'json' })
      .pipe(take(1)).subscribe({

      next: (segment: string) => {
        this.serverError = new ServerError()
        this.router.navigateByUrl('/r/create/result/' + segment);
      },

      error: (err: HttpErrorResponse) => {
        this.serverError = new ServerError(err, 'creating a new poll')
        if (this.serverError.status == 409) {
          // 409 Conflict received. Try to find a page to display for the user to fix the error.
          const field = this.serverError.message.split(' ')[0]
          for (let cur = this._current; cur !== undefined; cur = cur.parent) {
            if (cur.handledFields.has(field)) {
              this._sending = false
              this.makeCurrent(cur)
              break
            }
          }
        }
        if (this._sending) {
          this.router.navigateByUrl('/r/create/result/error')
        }
      },

    });
  }
}
