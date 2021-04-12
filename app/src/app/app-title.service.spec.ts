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

import { TestBed } from '@angular/core/testing';
import { Title } from '@angular/platform-browser';
import { Router } from '@angular/router';

import { RouterStub } from '../testing/router.stub';

import { AppTitleService } from './app-title.service';

describe('AppTitleService', () => {
  let service: AppTitleService;
  let routerStub: RouterStub;
  let titleSpy: jasmine.SpyObj<Title>;

  const baseTitle = 'base';
  let lastSetTitle: string|undefined;
  const setLastSetTitle = function(title: string) {
    lastSetTitle = title;
  };

  beforeEach(() => {
    routerStub = new RouterStub('');
    titleSpy = jasmine.createSpyObj('Title', ['getTitle','setTitle']);
    titleSpy.getTitle.and.returnValue(baseTitle);
    lastSetTitle = undefined;

    TestBed.configureTestingModule({
      providers: [
        { provide: Router, useValue: routerStub },
        { provide: Title, useValue: titleSpy },
      ],
    });
    service = TestBed.inject(AppTitleService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('forwards values from the real Title service', () => {
    titleSpy.getTitle.and.returnValue('some title');
    expect(service.getTitle()).toBe('some title');
    expect(titleSpy.getTitle).toHaveBeenCalled();
  });

  it('sets base title with undefined as argument', () => {
    titleSpy.setTitle.and.callFake(setLastSetTitle);
    service.setTitle(undefined);
    expect(titleSpy.setTitle).toHaveBeenCalled();
    expect(lastSetTitle).toBe(baseTitle);
  });

  it('sets title from a string', () => {
    titleSpy.setTitle.and.callFake(setLastSetTitle);
    service.setTitle('foo');
    expect(titleSpy.setTitle).toHaveBeenCalled();
    expect(lastSetTitle).toContain('foo');
    expect(lastSetTitle).toContain(baseTitle);
  });

  it('sets title from an array', () => {
    const array = ['foo', 'bar', 'baz'];
    titleSpy.setTitle.and.callFake(setLastSetTitle);
    service.setTitle(array);
    expect(titleSpy.setTitle).toHaveBeenCalled();
    for (const entry of array) {
      expect(lastSetTitle).toContain(entry);
    }
    expect(lastSetTitle).toContain(baseTitle);
  });

});
