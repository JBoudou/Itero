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
import { ReactiveFormsModule, FormsModule } from '@angular/forms';

import { ClipboardModule }  from '@angular/cdk/clipboard';
import { MatButtonModule }  from '@angular/material/button';
import { MatDialogModule }  from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule }    from '@angular/material/icon'; 
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule }  from '@angular/material/select'; 
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatRadioModule } from '@angular/material/radio';

import { FormWidgetModule } from '../form-widget/form-widget.module';
import { PollModule }       from '../poll/poll.module';
import { SharedModule }     from '../shared/shared.module';

import { CreateRoutingModule } from './create-routing.module';
import { CreateService, CREATE_TREE, APP_CREATE_TREE } from './create.service';
import { CreateGuard, LeaveCreateDialog } from './create.guard';

import { CreateComponent }              from './create.component';
import { GeneralComponent }             from './general/general.component';
import { ResultComponent }              from './result/result.component';
import { RoundComponent }               from './round/round.component';
import { SimpleAlternativesComponent }  from './simple-alternatives/simple-alternatives.component';
import { NavtreeComponent } from './navtree/navtree.component';
import { NavbuttonsComponent } from './navtree/navbuttons.component';

@NgModule({
  declarations: [
    CreateComponent,
    GeneralComponent,
    LeaveCreateDialog,
    ResultComponent,
    RoundComponent,
    SimpleAlternativesComponent,
    NavtreeComponent,
    NavbuttonsComponent,
  ],
  imports: [
    ClipboardModule,
    CommonModule,
    CreateRoutingModule,
    FormsModule,
    FormWidgetModule,
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatRadioModule,
    MatSelectModule,
    MatTooltipModule,
    PollModule,
    ReactiveFormsModule,
    SharedModule,
  ],
  providers: [
    CreateService,
    CreateGuard,
    { provide: CREATE_TREE, useValue: APP_CREATE_TREE },
  ],
})
export class CreateModule { }
