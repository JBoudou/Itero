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

import { MatButtonModule }  from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';

import { HarnessLoader } from '@angular/cdk/testing';
import { TestbedHarnessEnvironment } from '@angular/cdk/testing/testbed';
import { MatButtonHarness }  from '@angular/material/button/testing';
import { MatInputHarness } from '@angular/material/input/testing';

import { Subject } from 'rxjs';
import { filter } from 'rxjs/operators';

import { SimpleAlternativesComponent } from './simple-alternatives.component';

import { CreateService } from '../create.service';
import { CreateQuery, PollAlternative } from 'src/app/api';

import { ActivatedRouteStub } from '../../../testing/activated-route-stub'
import { Recorder } from 'src/testing/recorder';

function temporary(full: PollAlternative[]): { Name: string, Cost: number}[] {
  return full.map((alt: PollAlternative) => {
    return { Name: alt.Name, Cost: alt.Cost };
  });
}

describe('SimpleAlternativesComponent', () => {
  let component: SimpleAlternativesComponent;
  let fixture: ComponentFixture<SimpleAlternativesComponent>;
  let loader: HarnessLoader;
  let serviceSpy: jasmine.SpyObj<CreateService>;
  let activatedRouteStub: ActivatedRouteStub;
  let query$: Subject<Partial<CreateQuery>>;

  beforeEach(async () => {
    query$ = new Subject<Partial<CreateQuery>>();
    serviceSpy = jasmine.createSpyObj('CreateService', {register: {}}, { query$: query$ });
    activatedRouteStub = new ActivatedRouteStub();
    
    await TestBed.configureTestingModule({
      declarations: [ SimpleAlternativesComponent ],
      imports: [
        FormsModule,
        MatButtonModule,
        MatFormFieldModule,
        MatIconModule,
        MatInputModule,
        NoopAnimationsModule,
        ReactiveFormsModule,
      ],
      providers: [
        FormBuilder,
        { provide: CreateService, useValue: serviceSpy },
        { provide: ActivatedRoute, useValue: activatedRouteStub },
      ],
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(SimpleAlternativesComponent);
    loader = TestbedHarnessEnvironment.loader(fixture);
    component = fixture.componentInstance;
    fixture.detectChanges();
    jasmine.clock().install();
  });

  afterEach(() => {
    jasmine.clock().uninstall();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('synchronize alternatives correctly', async () => {
    const recorder = new Recorder<{ Name: string, Cost: number }[]>();
    recorder.listen(component.Alternatives.valueChanges.pipe(filter(component.filterEvent, component)));

    const seq = [
      [{Id: 0, Name: 'Un', Cost: 1}],
      [{Id: 0, Name: 'Un', Cost: 1}, {Id: 1, Name: 'Deux', Cost: 2}],
      [{Id: 0, Name: 'Un', Cost: 1}],
      [{Id: 1, Name: 'Deux', Cost: 2}],
      [{Id: 0, Name: 'Un', Cost: 1}, {Id: 1, Name: 'Deux', Cost: 2}, {Id: 2, Name: 'Trois', Cost: 3}],
      [{Id: 2, Name: 'Trois', Cost: 3}],
      [],
    ];
    for (const alts of seq) {
      query$.next({Alternatives: alts});
    }

    jasmine.clock().tick(1);
    fixture.detectChanges();
    await fixture.whenStable();

    recorder.unsubscribe();
    const len = seq.length;
    for (let i = 0; i < len; i++) {
      expect(recorder.record[i]).toEqual(temporary(seq[i]));
    }
  });

  it('add an alternative using the form', async () => {
    // TODO Change the recording. Use serviceSpy instead.
    const recorder = new Recorder<{ Name: string, Cost: number }[]>();
    recorder.listen(component.Alternatives.valueChanges.pipe(filter(component.filterEvent, component)));

    const newAlt = await loader.getChildLoader('.new-alternative');
    const input  = (await newAlt.getHarness(MatInputHarness )) as MatInputHarness;
    const button = (await newAlt.getHarness(MatButtonHarness)) as MatButtonHarness;
    await input.setValue('Un');
    await button.click(); 

    jasmine.clock().tick(1);
    fixture.detectChanges();
    await fixture.whenStable();
    recorder.unsubscribe();

    const last = recorder.record.length - 1;
    expect(recorder.record[last]).toEqual([{Name: 'Un', Cost: 1}]);
  });

  it('add three alternatives using the form', async () => {
    // TODO Change the recording. Use serviceSpy instead.
    const recorder = new Recorder<{ Name: string, Cost: number }[]>();
    recorder.listen(component.Alternatives.valueChanges.pipe(filter(component.filterEvent, component)));

    const newAlt = await loader.getChildLoader('.new-alternative');
    const input  = (await newAlt.getHarness(MatInputHarness )) as MatInputHarness;
    const button = (await newAlt.getHarness(MatButtonHarness)) as MatButtonHarness;

    const values = [
      { Name: 'Un'   , Cost: 1 },
      { Name: 'Deux' , Cost: 1 },
      { Name: 'Trois', Cost: 1 },
    ];
    for (const alt of values) {
      await input.setValue(alt.Name);
      await button.click();
    }

    jasmine.clock().tick(1);
    fixture.detectChanges();
    await fixture.whenStable();
    recorder.unsubscribe();

    const last = recorder.record.length - 1;
    expect(recorder.record[last]).toEqual(values);
  });


});
