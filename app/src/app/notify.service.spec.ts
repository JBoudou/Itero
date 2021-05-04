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

import { NotifyService } from './notify.service';

import { RouterStub } from '../testing/router.stub';
import { PollNotifService } from './poll-notif.service';

describe('NotifyService', () => {
  let service: NotifyService;
  let pollNotifSpy: jasmine.SpyObj<PollNotifService>;
  let routerStub: RouterStub;

  beforeEach(() => {
    pollNotifSpy = jasmine.createSpyObj('PollNotifService', [], ['event$']);
    routerStub = new RouterStub('');

    TestBed.configureTestingModule({
      providers: [
        { provide: PollNotifService, useValue: pollNotifSpy },
        { provide: Router, useValue: routerStub },
      ],
    });
    service = TestBed.inject(NotifyService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
