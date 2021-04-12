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

import { Event, Params, ParamMap, UrlTree, UrlSegmentGroup, UrlSegment } from '@angular/router';
import { Subject } from 'rxjs';

export class ParamMapStub implements ParamMap {
  keys: string[] = [];
  has(_: string): boolean {
    return false;
  }
  get(_: string): string|null {
    return null;
  }
  getAll(_: string): string[] {
    return [];
  }
}

export class UrlTreeStub implements UrlTree {

  root: UrlSegmentGroup;

  constructor(
    private segments: string,
  ) {
    const urlSegments = segments.split('/').map((s: string) => new UrlSegment(s, {}));
    this.root = new UrlSegmentGroup(urlSegments, {});
  }

  queryParams: Params = {};
  fragment: string|null = null;
  
  get queryParamMap(): ParamMap {
    return new ParamMapStub();
  }

  toString(): string {
    return this.segments;
  }

}

export class RouterStub {
  navigate = jasmine.createSpy('navigate');
  navigateByUrl = jasmine.createSpy('navigateByUrl');
  parseUrl = jasmine.createSpy('parseUrl');
  events = new Subject<Event>();

  get routerState() {
    return { snapshot: { url: this.url } };
  }

  constructor(
    public url: string = '',
  ) {

    this.navigate.and.callFake((url: string[]) => {
      this.url = url.join('/');
      // Fast promise because we don't want to wait one cycle.
      return { then: (resolve: (v: boolean) => void) => resolve(true) };
    });
    
    this.navigateByUrl.and.callFake((url: string) => {
      this.url = url;
      // Fast promise because we don't want to wait one cycle.
      return { then: (resolve: (v: boolean) => void) => resolve(true) };
    });

    this.parseUrl.and.callFake((url: string) => new UrlTreeStub(url));
  }

}
