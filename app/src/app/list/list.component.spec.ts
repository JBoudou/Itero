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

import { ListComponent } from './list.component';
import { ListService } from './list.service';

describe('ListComponent', () => {
  let component: ListComponent;
  let fixture: ComponentFixture<ListComponent>;
  let serviceSpy: jasmine.SpyObj<ListService>;

  beforeEach(async () => {
    serviceSpy = jasmine.createSpyObj('ListService', ['activate', 'desactivate']);

    await TestBed.configureTestingModule({
      declarations: [ ListComponent ],
      providers: [
        {provide: ListService, useValue: serviceSpy},
      ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('calls activate/desactivate on init/destroy', () => {
    expect(serviceSpy.activate).toHaveBeenCalled();
    fixture.destroy();
    expect(serviceSpy.desactivate).toHaveBeenCalled();
  });
});
