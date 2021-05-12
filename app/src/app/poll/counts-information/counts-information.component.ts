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

import { Component, Input, Output, EventEmitter, OnInit, OnDestroy, ElementRef } from '@angular/core';

import * as d3 from 'd3';

import { PollSubComponent, ServerError } from '../common';
import { CountInfoAnswer, CountInfoEntry } from '../../api';
import { CountsInformationService } from './counts-information.service';

const MaxAlternativeNameLength = 20;

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

function randomString(length: number): string {
  const alphabet = '0123456789abcdefghijklmnopqrstuvwxyz'
  const len = alphabet.length
  let ret = ''
  for (let i = 0; i < length; i++) {
    ret += alphabet.charAt(Math.floor(Math.random() * len))
  }
  return ret
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

  private hostElement: Element;

  constructor(
    eltRef: ElementRef<Element>,
    private service: CountsInformationService,
  ) {
    this.hostElement = eltRef.nativeElement;
  }


  ngOnInit(): void {
    this.service.information(this.pollSegment, this.round).then(
      (answer: CountInfoAnswer) => {
        // Check answer, mostly for tests
        if (answer === null || answer.Result === undefined || typeof answer.Result[Symbol.iterator] !== 'function') {
          // TODO send error
          return
        }

        this.createGraph(answer.Result);
      },
      (err: any) =>
       this.errors.emit(err as ServerError)
    );
  }

  ngOnDestroy(): void {
    this.winner.complete();
    this.errors.complete();
  }

  private createGraph(data: Array<CountInfoEntry>): void {
    const bar = { height: 24, sep: 2 }
    const size = { width: 400, height: (data.length + 1) * (bar.height + bar.sep) }
    const padding = { left: 10, right: 10, top: 0, bottom: bar.height + (2 * bar.sep) }
    const labelPadding = { left: 4, bottom: 6 }
    const anim = { duration: 1200, ease: d3.easeCubicInOut }

    const palette = (i: number) => {
      const values = [ '#602c57', '#f4723c', '#9c365f', '#ffa600', '#d14b55' ]
      return values[i % values.length]
    }

    const uniqId = randomString(7)

    const x = d3.scaleLinear()
      .domain([0, d3.max(data, (d: CountInfoEntry) => d.Count)])
      .range([padding.left, size.width - padding.right])

    const y = d3.scaleBand()
      .domain(data.map((d: CountInfoEntry) => d.Alternative.Name))
      .range([padding.top, size.height - padding.bottom])
      .paddingInner(bar.sep / (bar.sep + bar.height))

    const svg = d3.select(this.hostElement).select('svg')
      .attr('viewBox', '0 0 ' + size.width + ' ' + size.height)

    const labelBack = svg.append('g')
      .selectAll('text')
      .data(data)
      .join('text')
        .attr('x', x(0) + labelPadding.left)
        .attr('y', (d: CountInfoEntry) => y(d.Alternative.Name))
        .attr('dy', y.bandwidth() - labelPadding.bottom)
        .attr('fill', '#111')
        .text((d: CountInfoEntry) => d.Alternative.Name)

    const bars = svg.append('g')
      .selectAll('rect')
      .data(data)
      .join('rect')
        .attr('fill', (_: any, i: number) => palette(i))
        .attr('x', x(0))
        .attr('y', (d: CountInfoEntry) => y(d.Alternative.Name))
        .attr('height', y.bandwidth())
        .attr('width', 0)
        .transition()
          .duration(anim.duration)
          .ease(anim.ease)
          .attr('width', (d: CountInfoEntry) => x(d.Count) - x(0))

    const labelFront = svg.append('g')
    labelFront
      .selectAll('clipPath')
      .data(data)
      .join('clipPath')
        .attr('id', (d: any, i: number) => uniqId + '-lc' + i)
        .append('rect')
          .attr('x', x(0))
          .attr('y', (d: CountInfoEntry) => y(d.Alternative.Name))
          .attr('height', y.bandwidth())
          .attr('width', 0)
          .transition()
            .duration(anim.duration)
            .ease(anim.ease)
            .attr('width', (d: CountInfoEntry) => x(d.Count) - x(0))

    labelFront
      .selectAll('text')
      .data(data)
      .join('text')
        .attr('x', x(0) + labelPadding.left)
        .attr('y', (d: CountInfoEntry) => y(d.Alternative.Name))
        .attr('dy', y.bandwidth() - labelPadding.bottom)
        .attr('fill', (_: any, i: number) => d3.lab(palette(i)).l < 60 ? 'white' : 'black')
        .attr('clip-path', (_:any, i: number) => 'url(#' + uniqId + '-lc' + i + ')')
        .text((d: CountInfoEntry) => d.Alternative.Name)

    const xAxis = svg.append('g')
      .attr('transform', 'translate(0,' + (size.height - padding.bottom) + ')')
      .call(
        d3.axisBottom(x)
          .tickValues(x.ticks().filter(Number.isInteger))
          .tickFormat(d3.format('d'))
      )
  }

}
