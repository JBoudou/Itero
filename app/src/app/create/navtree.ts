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

import { CreateQuery } from '../api';

/** Event about the current step in the creation tree. */
export class NavStepStatus {
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
export abstract class NavTreeNode {
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

  private _parent: NavTreeNode|undefined;

  get parent(): NavTreeNode|undefined {
    return this._parent;
  }

  /** Utility function for subclasses. */
  protected setAsChild(child: NavTreeNode): void {
    child._parent = this;
  }

  /** All the possible successors of the node. */
  abstract readonly children: NavTreeNode[];

  /**
   * Determine the successor of the node depending on the current state of the query.
   * Should always be called without parameters.
   */
  abstract next(query?: Partial<CreateQuery>): NavTreeNode|undefined;

  /** Construct the NavStepStatus event to sent when the node is the current one. */
  makeStatus(): NavStepStatus {
    const steps = [this.asString]
    
    let pos = 0;
    let parent = this._parent;
    while (parent !== undefined) {
      steps.unshift(parent.asString);
      pos += 1;
      parent = parent._parent;
    }
    
    let child = this as NavTreeNode;
    while (true) {
      let next = child.next(this.query);
      if (next === undefined) {
        break;
      }
      child = next;
      steps.push(child.asString);
    }

    return new NavStepStatus(pos, steps, !child.isFinal);
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
export class LinearNavTreeNode extends NavTreeNode {
  constructor(
    segment: string, asString: string,
    private _next: NavTreeNode
  ) {
    super(segment, asString);
    this.setAsChild(this._next);
  }

  get children(): NavTreeNode[] {
    return [this._next];
  }

  next(_?: Partial<CreateQuery>): NavTreeNode|undefined {
    return this._next;
  }
}

/** A leaf node. */
export class FinalNavTreeNode extends NavTreeNode {
  constructor(
    segment: string, asString: string,
  ) {
    super(segment, asString);
  }

  get isFinal(): boolean {
    return true;
  }

  get children(): NavTreeNode[] {
    return [];
  }

  next(_?: Partial<CreateQuery>): NavTreeNode|undefined {
    return undefined;
  }
}
