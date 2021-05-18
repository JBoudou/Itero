// Itero - Online iterative vote application
// Copyright (C) 2020 Joseph Boudou
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

import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CountsInformationComponent, ticks1235 } from './counts-information.component';
import { CountsInformationService } from './counts-information.service';

describe('CountsInformationComponent', () => {
  let component: CountsInformationComponent;
  let fixture: ComponentFixture<CountsInformationComponent>;
  let serviceSpy: jasmine.SpyObj<CountsInformationService>;

  beforeEach(async () => {
    serviceSpy = jasmine.createSpyObj('CountsInformationService', {information: Promise.resolve({})});

    await TestBed.configureTestingModule({
      declarations: [ CountsInformationComponent ],
      providers: [
        { provide: CountsInformationService, useValue: serviceSpy },
      ],
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(CountsInformationComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});

describe('ticks1235', () => {

  const tests = [
    { input:     2, expect: [0, 1, 2] },
    { input:     6, expect: [0, 1, 2, 3, 4, 5, 6] },
    { input:    11, expect: [0, 2, 4, 6, 8, 10] },
    { input: 13579, expect: [0, 3000, 6000, 9000, 12000] },
  ]
  for (let t of tests) {
    it('compute the ticks for ' + t.input, () => {
      expect(ticks1235({ domain(): number[] { return [0, t.input] }})).toEqual(t.expect)
    })
  }
})
