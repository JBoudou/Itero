import { Component, OnInit } from '@angular/core';
import { FormBuilder, Validators } from '@angular/forms';

import { map } from 'rxjs/operators';

import { CreateService, CreateSubComponent } from '../create/create.service';

@Component({
  selector: 'app-create-general',
  templateUrl: './create-general.component.html',
  styleUrls: ['./create-general.component.sass']
})
export class CreateGeneralComponent implements OnInit, CreateSubComponent {

  form = this.formBuilder.group({
    Title: ['', [
      Validators.required,
    ]],
    Description: [''],
  });

  get handledFields(): Set<string> {
    return new Set<string>(Object.keys(this.form.controls));
  }

  readonly validable$ = this.form.statusChanges.pipe(map(val => val == 'VALID'));

  constructor(
    private service: CreateService,
    private formBuilder: FormBuilder,
  ) { }

  ngOnInit(): void {
     const query = this.service.register(this);
     for (const prop in this.form.controls) {
       this.form.controls[prop].setValue(query[prop]);
       this.form.controls[prop].valueChanges.subscribe({
         next: value => query[prop] = value,
       });
     }
  }

}
