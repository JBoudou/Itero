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
import { Router } from '@angular/router';

import { Observable, ReplaySubject, Subscription } from 'rxjs';
import { cloneDeep } from 'lodash';

import { CreateQuery } from '../api';


export interface CreateSubComponent {
  /** Name of query's properties that may be modified by the component. */
  readonly handledFields: Set<string>;
  readonly validable$: Observable<boolean>;
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

/** Event about the current step in the creation tree. */
export class CreateStepStatus {
  constructor(
    /** Depth of the current node in the tree (root has depth zero). This is also the index in steps. */
    public current: number,
    /** Branch of the the tree containing the current node. */
    public steps: string[],
    /** Whether steps is a partial branch, up to the last undecided conditional node. */
    public mayHaveMore = false
  ) {};
}

/**
 * Base class for nodes of the creation tree.
 *
 * The creation tree describes the steps of the creation procedure. Each step corresponds to a route
 * wich is obviously associated to a component. Each node contains the current state of the query,
 * allowing fine grain undo/redo behaviour. Each node decides wich node is its current successor,
 * depending on the current parameters of the query.
 */
export abstract class CreateTreeNode {
  public query: Partial<CreateQuery> = {};
  public handledFields: Set<string> = new Set<string>();

  constructor(
    /** Last segment of the route corresponding to the step. */
    public readonly segment: string,
    /** Description of the step, for display to the user. */
    public readonly asString: string,
  ) {}
  
  /** Whether the node is a leaf. */
  get isFinal(): boolean {
    return false;
  }

  private _parent: CreateTreeNode|undefined;

  get parent(): CreateTreeNode|undefined {
    return this._parent;
  }

  /** Utility function for subclasses. */
  protected setAsChild(child: CreateTreeNode): void {
    child._parent = this;
  }

  /** All the possible successors of the node. */
  abstract readonly children: CreateTreeNode[];

  /**
   * Determine the successor of the node depending on the current state of the query.
   * Should always be called without parameters.
   */
  abstract next(query?: Partial<CreateQuery>): CreateTreeNode|undefined;

  /** Construct the CreateStepStatus event to sent when the node is the current one. */
  makeStatus(): CreateStepStatus {
    const steps = [this.asString]
    
    let pos = 0;
    let parent = this._parent;
    while (parent !== undefined) {
      steps.unshift(parent.asString);
      pos += 1;
      parent = parent._parent;
    }
    
    let child = this as CreateTreeNode;
    while (true) {
      let next = child.next(this.query);
      if (next === undefined) {
        break;
      }
      child = next;
      steps.push(child.asString);
    }

    return new CreateStepStatus(pos, steps, !child.isFinal);
  }

  /** 
   * Reset the state of the query for this node and transitively all its descendants.
   * Should be called only on the root of the tree.
   */
  reset(): void {
    this.query = {};
    this.handledFields = new Set<string>();
    for (let child of this.children) {
      child.reset();
    }
  }
}

/** A node with always exactely one child. */
export class LinearCreateTreeNode extends CreateTreeNode {
  constructor(
    segment: string, asString: string,
    private _next: CreateTreeNode
  ) {
    super(segment, asString);
    this.setAsChild(this._next);
  }

  get children(): CreateTreeNode[] {
    return [this._next];
  }

  next(_?: Partial<CreateQuery>): CreateTreeNode|undefined {
    return this._next;
  }
}

/** A leaf node. */
export class FinalCreateTreeNode extends CreateTreeNode {
  constructor(
    segment: string, asString: string,
  ) {
    super(segment, asString);
  }

  get isFinal(): boolean {
    return true;
  }

  get children(): CreateTreeNode[] {
    return [];
  }

  next(_?: Partial<CreateQuery>): CreateTreeNode|undefined {
    return undefined;
  }
}

export const CREATE_TREE = new InjectionToken<CreateTreeNode>('create.tree');

export const APP_CREATE_TREE: CreateTreeNode =
  new LinearCreateTreeNode('general', 'Generalities',
    new LinearCreateTreeNode('simpleAlternatives', 'Alternatives',
      new FinalCreateTreeNode('round', 'Rounds')
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
@Injectable({
  providedIn: 'root'
})
export class CreateService {

  private _createNext = new ReplaySubject<CreateNextStatus>(1);
  private _createStep = new ReplaySubject<CreateStepStatus>(1);

  get createNextStatus$(): Observable<CreateNextStatus> {
    return this._createNext;
  }

  get createStepStatus$(): Observable<CreateStepStatus> {
    return this._createStep;
  }

  private current: CreateTreeNode;
  private subComponent: CreateSubComponent|undefined;
  private subComponentSubscription: Subscription|undefined;

  constructor(
    @Inject(CREATE_TREE) private root: CreateTreeNode,
    private router: Router,
  ) {
    this.current = this.root;
  }

  /**
   * Asks the router to go back in the creation tree.
   *
   * Must not be called when the current node is the root node, which is notified by a
   * CreateStepStatus with current at zero.
   */
  back(): void {
    const parent = this.current.parent;
    if (parent === undefined) {
      console.warn('CreateService back on the root node !')
      return;
    }

    if (this.subComponent === undefined) {
      console.warn('CreateService back without component !')
    }

    parent.handledFields = new Set<string>();
    this.makeCurrent(parent);
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
      this.root.reset();
      this.current = this.root;
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

  private makeCurrent(node: CreateTreeNode|undefined): void {
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
    console.log(JSON.stringify(this.current.query));
    this.router.navigateByUrl('/r/list');
    // TODO
  }
}
