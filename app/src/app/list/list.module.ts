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

import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

import { MatDialogModule } from '@angular/material/dialog';
import { MatIconModule }    from '@angular/material/icon'; 

import { SharedModule } from '../shared/shared.module';

import { ListRoutingModule } from './list-routing.module';

import { ListComponent } from './list.component';
import { PollsTableComponent } from './polls-table/polls-table.component';
import { DeletePollDialog } from './list.service';


@NgModule({
  declarations: [
    DeletePollDialog,
    ListComponent,
    PollsTableComponent,
  ],
  imports: [
    CommonModule,
    ListRoutingModule,
    MatDialogModule,
    MatIconModule,
    SharedModule,
  ],
  exports: [
    ListComponent,
  ],
})
export class ListModule { }
