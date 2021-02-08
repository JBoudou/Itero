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
import { Router } from '@angular/router';

import { MatDialogModule }  from '@angular/material/dialog';

import { CreateGuard } from './create.guard';

import { CreateService } from './create.service';

import { RouterStub } from '../../testing/router.stub';

describe('CreateGuard', () => {
  let guard: CreateGuard;
  let createSpy: jasmine.SpyObj<CreateService>;
  let routerSpy: RouterStub;

  beforeEach(() => {
    createSpy = jasmine.createSpyObj('CreateService', ['isStarted', 'reset']);

    TestBed.configureTestingModule({
      imports: [ MatDialogModule ],
      providers: [
        CreateGuard,
        { provide: CreateService, useValue: createSpy },
        { provide: Router, useValue: routerSpy },
      ],
    });
    guard = TestBed.inject(CreateGuard);
  });

  it('should be created', () => {
    expect(guard).toBeTruthy();
  });
});
