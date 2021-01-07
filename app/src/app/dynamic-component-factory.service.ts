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

import { Injectable, ComponentFactoryResolver, ViewContainerRef, Type } from '@angular/core';

/**
 * A thin wrapper around ComponentFactoryResolver.
 *
 * The only goal of this service is to ease tests of component with dynamic nested components.
 * If you know about better alternative, please let us know.
 * Testing this service would be particularly difficult. Hence please keep it stupid simple.
 */
@Injectable({
  providedIn: 'root'
})
export class DynamicComponentFactoryService {

  constructor(
    private componentFactoryResolver: ComponentFactoryResolver,
  ) { }

  createComponent<T>(viewContainerRef: ViewContainerRef, type: Type<T>): T {
    const componentFactory = this.componentFactoryResolver.resolveComponentFactory(type);
    const componentRef = viewContainerRef.createComponent<T>(componentFactory);
    return componentRef.instance;
  }
    
}
