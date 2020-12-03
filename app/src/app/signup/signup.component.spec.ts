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
import { FormBuilder } from '@angular/forms';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { map } from 'rxjs/operators';

import { SignupComponent } from './signup.component';
import { SessionService } from '../session/session.service';

describe('SignupComponent', () => {
  let component: SignupComponent;
  let fixture: ComponentFixture<SignupComponent>;
  let httpControler: HttpTestingController;
  let sessionSpy: jasmine.SpyObj<SessionService>;

  beforeEach(async () => {
    const sessionSpy = jasmine.createSpyObj('SessionService', ['httpOperator']);

    await TestBed.configureTestingModule({
      declarations: [ SignupComponent ],
      imports: [ HttpClientTestingModule ],
      providers: [
        FormBuilder,
        {provide: SessionService, useValue: sessionSpy},
      ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(SignupComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
    httpControler = TestBed.inject(HttpTestingController);
    sessionSpy = TestBed.inject(SessionService) as jasmine.SpyObj<SessionService>;
  });

  afterEach(() => {
    httpControler.verify();
  })

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('pipes signup to SessionService.httpOperator', () => {
    const sessionId = 'ABCD';

    const spyFct = jasmine.createSpy('filter');
    spyFct.withArgs(sessionId).and.returnValue(true);
    sessionSpy.httpOperator.and.returnValue(map(spyFct));

    component.onSignup();

    const req = httpControler.expectOne('/a/signup');
    req.flush(sessionId);
    httpControler.verify();

    expect(spyFct.calls.count()).toBe(1);
  });
});
