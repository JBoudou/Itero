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

import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ReactiveFormsModule, FormBuilder, FormsModule } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatRadioModule } from '@angular/material/radio';

import { TestbedHarnessEnvironment } from '@angular/cdk/testing/testbed';
import { HarnessLoader } from '@angular/cdk/testing';
import { MatCheckboxHarness } from '@angular/material/checkbox/testing';

import { BehaviorSubject, of } from 'rxjs';

import { AccessComponent } from './access.component';

import { CreateService } from '../create.service';
import { SessionService } from 'src/app/session/session.service';
import { CreateQuery } from 'src/app/api';

import { ActivatedRouteStub } from 'src/testing/activated-route-stub'
import {ServerError} from 'src/app/shared/server-error';

describe('AccessComponent', () => {
  let component: AccessComponent;
  let fixture: ComponentFixture<AccessComponent>;
  let loader: HarnessLoader
  let activatedRouteStub: ActivatedRouteStub;
  let query$: BehaviorSubject<Partial<CreateQuery>>;
  let sessionSpy : jasmine.SpyObj<SessionService>;
  let serviceSpy: jasmine.SpyObj<CreateService>;

  beforeEach(async () => {
    activatedRouteStub = new ActivatedRouteStub();
    query$ = new BehaviorSubject<Partial<CreateQuery>>({Title: 'Foo'});
    serviceSpy = jasmine.createSpyObj('CreateService',
                                      { patchQuery: true },
                                      { query$: query$, serverError: new ServerError() });
    sessionSpy = jasmine.createSpyObj('SessionService', ['checkSession', 'login'], {
      state$: of({logged: false}),
    });
    
    await TestBed.configureTestingModule({
      declarations: [ AccessComponent ],
      imports: [
        FormsModule,
        MatCheckboxModule,
        MatFormFieldModule,
        MatInputModule,
        MatRadioModule,
        NoopAnimationsModule,
        ReactiveFormsModule,
      ],
      providers: [
        FormBuilder,
        { provide: CreateService, useValue: serviceSpy },
        { provide: ActivatedRoute, useValue: activatedRouteStub },
        { provide: SessionService, useValue: sessionSpy },
      ],
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(AccessComponent);
    loader = TestbedHarnessEnvironment.loader(fixture)
    component = fixture.componentInstance;
    fixture.detectChanges();
    jasmine.clock().install();
    activatedRouteStub.nextUrlFromString('test');
  });

  afterEach(() => {
    jasmine.clock().uninstall();
  });

  it('should create', () => {
    expect(component).toBeTruthy()
    expect(component.shortURLElt.checked).toBeFalse()
  });

  it('check the checkbox when receiving the query', async () => {
    const query = {
      Electorate: 1,
      Hidden: true,
      ShortURL: 'a b',
    }
    query$.next(query)

    jasmine.clock().tick(1)
    await fixture.whenStable()
    
    expect(component.form.value).toEqual(query)
    expect(component.shortURLElt.checked).toBeTrue()
  })

  it('remove ShortURL from the query when unchecking the checkbox', async () =>  {
    query$.next({ ShortURL: 'something'})
    jasmine.clock().tick(1)
    await fixture.whenStable()
    serviceSpy.patchQuery.calls.reset()

    const cb = await loader.getHarness(MatCheckboxHarness)
    await cb.uncheck()

    expect(serviceSpy.patchQuery).toHaveBeenCalledWith('test', jasmine.objectContaining({ShortURL: undefined}))
  })

});
