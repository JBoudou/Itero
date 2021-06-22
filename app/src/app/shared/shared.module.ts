// Itero - Online iterative vote application
// Copyright (C) 2021 Joseph Boudou, Wan JIN
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
import { OrdinalPipe } from './ordinal.pipe';
import { NearDatePipe } from './near-date.pipe';
import { ServerErrorComponent } from './server-error/server-error.component';
import { InfoPanelComponent } from './info-panel/info-panel.component';

import { MatButtonModule } from '@angular/material/button';
import { MatIconModule }    from '@angular/material/icon';

@NgModule({
  declarations: [OrdinalPipe, NearDatePipe, ServerErrorComponent, InfoPanelComponent],
  imports: [
    CommonModule,
    MatButtonModule,
    MatIconModule,
  ],
  exports: [
    InfoPanelComponent,
    NearDatePipe,
    OrdinalPipe,
    ServerErrorComponent,
  ]
})
export class SharedModule { }
