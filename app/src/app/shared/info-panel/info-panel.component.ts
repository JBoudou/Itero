// Itero - Online iterative vote application
// Copyright (C) 2021 Wan JIN, Joseph Boudou
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


import { Component, HostBinding, OnInit } from '@angular/core';

@Component({
  selector: 'app-info-panel',
  templateUrl: './info-panel.component.html',
  styleUrls: ['./info-panel.component.sass']
})
export class InfoPanelComponent implements OnInit {

  @HostBinding('class') get stateClass() {
    return this.showInfo ? 'opened' : 'closed'
  }

  showInfo: boolean = false;

  constructor() { }

  ngOnInit(): void {
  }

}
