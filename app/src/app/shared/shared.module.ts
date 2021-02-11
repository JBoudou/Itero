import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { OrdinalPipe } from './ordinal.pipe';
import { NearDatePipe } from './near-date.pipe';



@NgModule({
  declarations: [OrdinalPipe, NearDatePipe],
  imports: [
    CommonModule
  ],
  exports: [OrdinalPipe, NearDatePipe]
})
export class SharedModule { }
