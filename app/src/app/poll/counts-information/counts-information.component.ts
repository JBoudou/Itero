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

import { Component, Input, Output, EventEmitter, OnInit, OnDestroy } from '@angular/core';

import { PollSubComponent, ServerError } from '../common';
import { CountInfoAnswer } from '../../api';
import { CountsInformationService } from './counts-information.service';

/** Extract the first element of a comma separated list, removing one level of enclosing " or '. */
function extractFontFamily(css: string): string {
  var first = css.split(",", 1)[0].trim();
  if (first[0] == '"' || first[0] == "'") {
    first = first.split(first[0])[1];
  }
  return first;
}

/**
 * Extract a font size in px.
 * An approximation is done if the size is in pt.
 * A default value of 12 is returned is the size cannot be extracted.
 */
function extractFontSize(css: string): number {
  if (css.endsWith("px")) {
    return Number(css.slice(0, -2));
  }
  if (css.endsWith("pt")) {
    return Number(css.slice(0, -2)) * 1.2;
  }
  return 12;
}


@Component({
  selector: 'app-counts-information',
  templateUrl: './counts-information.component.html',
  styleUrls: ['./counts-information.component.sass']
})
export class CountsInformationComponent implements OnInit, OnDestroy, PollSubComponent {

  @Input() pollSegment: string;
  @Input() round: number|undefined;
  @Output() winner = new EventEmitter<string>();
  @Output() errors = new EventEmitter<ServerError>();

  data: any[][];

  options = {
    bars: 'horizontal',
    legend: { position: 'none' },
    height: undefined,
    fontName: undefined,
    fontSize: undefined,
    chartArea: { left: '35%', top: 0 },
    animation: {
      duration: 1200,
      easing: 'inAndOut',
      startup: true
    },
    hAxis: {
      minValue: 0,
      maxValue: undefined
    },
    bar: { groupWidth: '80%' },
    annotations: {
      textStyle: { fontSize: undefined }
    }
  };

  columns = [
    {label: 'Alternative'},
    {label: 'Votes'},
    {role: 'tooltip'},
    {role: 'annotation'},
    {role: 'style'}
  ];

  constructor(
    private service: CountsInformationService,
  ) { }

  private static palette = [ '#602c57', '#f4723c', '#9c365f', '#ffa600', '#d14b55' ];

  ngOnInit(): void {
    const style = window.getComputedStyle(document.body);
    const fontSize = extractFontSize(style.fontSize);
    this.options.fontName = extractFontFamily(style.fontFamily);
    this.options.fontSize = fontSize * 0.9;
    this.options.annotations.textStyle.fontSize = fontSize * 0.75;

    this.service.information(this.pollSegment, this.round).then(
      (answer: CountInfoAnswer) => {
        // First pass
        var maxCount = 0;
        var sumCount = 0;
        var maxLen = 0;
        for (let entry of answer.Result) {
          sumCount += entry.Count;
          if (entry.Count > maxCount) {
            maxCount = entry.Count;
            this.winner.emit(entry.Alternative.Name);
          }
          if (entry.Alternative.Name.length > maxLen) {
            maxLen = entry.Alternative.Name.length;
          }
        };

        // Set global options
        this.options.height = 32 * (answer.Result.length + 1);
        this.options.hAxis.maxValue = Math.min(5 * ((maxCount / 5) + 1), sumCount, maxCount * 2);
        if (maxLen > 10) {
          this.options.chartArea.left = '35%';
        } else {
          this.options.chartArea.left = '20%';
        }

        // Second pass
        this.data = [];
        for (let entry of answer.Result) {
          var shortName = entry.Alternative.Name;
          if (shortName.length > 21) {
            shortName = shortName.slice(0, 20) + '...';
          }
          const tooltip = entry.Alternative.Name;
          const annotation = String(Math.round(entry.Count * 1000 / sumCount) / 10) + '%';
          const palletteId = this.data.length % CountsInformationComponent.palette.length;
          const style = CountsInformationComponent.palette[palletteId];
          this.data.push([shortName, entry.Count, tooltip, annotation, style]);
        }
      },
      (err: any) =>
       this.errors.emit(err as ServerError)
    );
  }

  ngOnDestroy(): void {
    this.winner.complete();
    this.errors.complete();
  }

}
