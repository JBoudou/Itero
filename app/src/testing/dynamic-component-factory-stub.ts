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

import { ComponentRef, Type, ViewContainerRef } from '@angular/core';

function simpleComponentRef<T>(instance: T, type: Type<T>): ComponentRef<T> {
  return {
    instance: instance,
    componentType: type,
    destroy(): void {}
  } as ComponentRef<T>;
}

/** A stub for DynamicComponentFactoryService. */
export class DynamicComponentFactoryStub {

  private returnMap: Map<Type<any>, any[]>;
  private callsMap: Map<Type<any>, number>;

  constructor() {
    this.reset();
  }

  reset(): void {
    this.returnMap = new Map();
    this.callsMap = new Map();
  }

  /**
   * Provide the component to be returned by a call to createComponent with the given type.
   * Components are stored in a FIFO queue.
   */
  nextComponent(type: Type<any>, component: any): void {
    if (!this.returnMap.has(type)) {
      this.returnMap.set(type, new Array());
    }
    this.returnMap.get(type).push(component);
  }

  createComponent<T>(_: ViewContainerRef, type: Type<T>): ComponentRef<T> {
    if (this.callsMap.has(type)) {
      this.callsMap.set(type, this.callsMap.get(type) + 1);
    } else {
      this.callsMap.set(type, 1);
    }
    if (this.returnMap.has(type)) {
      return simpleComponentRef(this.returnMap.get(type).shift(), type);
    }
    return undefined;
  }

  /** The number of times createComponent has been called with the given type. */
  calls(type: Type<any>): number {
    if (this.callsMap.has(type)) {
      return this.callsMap.get(type);
    }
    return 0;
  }

}
