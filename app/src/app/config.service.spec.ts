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
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';

import { ConfigService } from './config.service';
import {Recorder} from 'src/testing/recorder';

describe('ConfigService', () => {
  let service: ConfigService;
  let httpControler: HttpTestingController;

  beforeEach(() => {
    jasmine.clock().install();

    TestBed.configureTestingModule({
      imports: [
        HttpClientTestingModule,
      ],
    });
    service = TestBed.inject(ConfigService);
    httpControler = TestBed.inject(HttpTestingController);
  });

  afterEach(function() {
    jasmine.clock().uninstall();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('forwards DemoPollSegment', () => {
    const recorder = new Recorder<string>();
    recorder.listen(service.demoPollSegment$);

    const req = httpControler.expectOne('/a/config');
    req.flush({DemoPollSegment: 'abc'});
    jasmine.clock().tick(1);

    expect(recorder.record.length).toBeGreaterThan(0);
    expect(recorder.record[recorder.record.length - 1]).toBe('abc');
  });

  it('sends empty string as DemoPollSegment if missing', () => {
    const recorder = new Recorder<string>();
    recorder.listen(service.demoPollSegment$);

    const req = httpControler.expectOne('/a/config');
    req.flush({Foo: 'abc'});
    jasmine.clock().tick(1);

    expect(recorder.record.length).toBeGreaterThan(0);
    expect(recorder.record[0]).toBe('');
  });

  it('sends empty string as DemoPollSegment on error', () => {
    const recorder = new Recorder<string>();
    recorder.listen(service.demoPollSegment$);

    const req = httpControler.expectOne('/a/config');
    req.flush({DemoPollSegment: 'abc'}, {status: 403, statusText: 'Forbidden'});
    jasmine.clock().tick(1);

    expect(recorder.record.length).toBeGreaterThan(0);
    expect(recorder.record[0]).toBe('');
  });

});
