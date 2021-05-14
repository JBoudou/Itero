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

import { Component, Input, Output, EventEmitter, OnInit, OnDestroy, ElementRef, ViewEncapsulation } from '@angular/core';

import * as d3 from 'd3';

import { PollSubComponent, ServerError } from '../common';
import { CountInfoAnswer, CountInfoEntry } from '../../api';
import { CountsInformationService } from './counts-information.service';

function randomString(length: number): string {
  const alphabet = '0123456789abcdefghijklmnopqrstuvwxyz'
  const len = alphabet.length
  let ret = ''
  for (let i = 0; i < length; i++) {
    ret += alphabet.charAt(Math.floor(Math.random() * len))
  }
  return ret
}

function ticks1235interval(scale: { domain(): number[] }): number {
  let max = d3.max(scale.domain())
  if (max < 3) {
    return 1
  }
  let pow = 0
  while (max > 30) {
    max /= 10
    pow += 1
  }
  const base = max <  7 ? 1 :
               max < 13 ? 2 :
               max < 19 ? 3 :
                          5 ;
  return base * (10 ** pow)
}

export function ticks1235(scale: { domain(): number[] }): number[] {
  const max = d3.max(scale.domain())
  const interval = ticks1235interval(scale)
  const ret = Array<number>(Math.floor(max/interval) + 1)
  for (let i = 0; i <= max; i += interval) ret[i / interval] = i;
  return ret
}


@Component({
  selector: 'app-counts-information',
  templateUrl: './counts-information.component.html',
  styleUrls: ['./counts-information.component.sass'],
  encapsulation: ViewEncapsulation.None,
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
        
        const winner = answer.Result[d3.maxIndex(answer.Result, (d: CountInfoEntry) => d.Count)]
        this.winner.emit(winner.Alternative.Name)
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
    // Geometry constants
    const bar = { height: 22, sep: 3 }
    const size = { width: 400, height: (data.length + 1) * (bar.height + bar.sep) }
    const padding = { left: 3, right: 34, top: bar.height + bar.sep, bottom: 0 }
    const labelPadding = { left: 3, right: 3, bottom: 6 }
    const anim = { duration: 1200, ease: d3.easeCubicInOut }

    // Unique identifier for this graph.
    const uniqId = randomString(7)

    const palette = (i: number) => {
      const values = [ '#602c57', '#f4723c', '#9c365f', '#ffa600', '#d14b55' ]
      return values[i % values.length]
    }

    // Scales //

    const max = d3.max(data, (d: CountInfoEntry) => d.Count)
    const x = d3.scaleLinear()
      .domain([0, max])
      .range([padding.left, size.width - padding.right])

    const maxLabelRight = size.width - (padding.right + labelPadding.right)
    const xClip = d3.scaleLinear()
      .domain([0, x.invert(maxLabelRight), max])
      .range([padding.left, maxLabelRight, maxLabelRight])

    const y = d3.scaleBand()
      .domain(data.map((d: CountInfoEntry) => d.Alternative.Name))
      .range([padding.top, size.height - padding.bottom])
      .paddingInner(bar.sep / (bar.sep + bar.height))

    // SVG //

    const svg = d3.select(this.hostElement).select('svg')
      .attr('viewBox', '0 0 ' + size.width + ' ' + size.height)

    const tickSize = size.height - ((padding.top * 0.7) + padding.bottom)
    const xAxis = svg.append('g')
      .attr('transform', 'translate(0,' + (size.height - padding.bottom) + ')')
      .call(
        d3.axisTop(x)
          .tickValues(ticks1235(x))
          .tickFormat(d3.format('d'))
          .tickSize(tickSize)
      )
      .call(g => g.select('.domain').remove())
      .call(g => g.selectAll('.tick line')
        .attr('stroke-opacity', 0.2)
        .attr('stroke-width', '1.5')
      )

    const labels =
      function <GE extends d3.BaseType,PE extends d3.BaseType,PD>
               (s: d3.Selection<GE,CountInfoEntry,PE,PD>): void
    { s
      .attr('x', x(0) + labelPadding.left)
      .attr('y', (d: CountInfoEntry) => y(d.Alternative.Name))
      .attr('dy', y.bandwidth() - labelPadding.bottom)
      .text((d: CountInfoEntry) => d.Alternative.Name)
    }

    const labelBack = svg.append('g')
    labelBack
      .append('clipPath')
        .attr('id', uniqId + '-lbc')
        .append('rect')
          .attr('x', xClip(0))
          .attr('y', 0)
          .attr('height', size.height)
          .attr('width', xClip(max) - xClip(0))
    labelBack
      .selectAll('text')
      .data(data)
      .join('text')
        .call(labels)
        .attr('fill', '#111')
        .attr('clip-path', 'url(#' + uniqId + '-lbc)')

    const bars =
      function <GE extends d3.BaseType,PE extends d3.BaseType,PD,R>
               (s: d3.Selection<GE,CountInfoEntry,PE,PD>, scale: d3.ScaleContinuousNumeric<R,number,number>): void
    { s
      .attr('x', scale(0))
      .attr('y', (d: CountInfoEntry) => y(d.Alternative.Name))
      .attr('height', y.bandwidth())
      .attr('width', 0)
      .transition()
        .duration(anim.duration)
        .ease(anim.ease)
        .attr('width', (d: CountInfoEntry) => scale(d.Count) - scale(0));
    }
    svg.append('g')
      .selectAll('rect')
      .data(data)
      .join('rect')
        .attr('fill', (_: any, i: number) => palette(i))
        .call(bars, x)

    const labelFront = svg.append('g')
    labelFront
      .selectAll('clipPath')
      .data(data)
      .join('clipPath')
        .attr('id', (_: any, i: number) => uniqId + '-lfc' + i)
        .append('rect')
          .call(bars, xClip)
    labelFront
      .selectAll('text')
      .data(data)
      .join('text')
        .call(labels)
        .attr('fill', (_: any, i: number) => d3.lab(palette(i)).l < 60 ? 'white' : 'black')
        .attr('clip-path', (_:any, i: number) => 'url(#' + uniqId + '-lfc' + i + ')')

    const total = d3.sum(data, (d: CountInfoEntry) => d.Count)
    const percents = svg.append('g')
      .attr('class','percent-label')
      .selectAll('text')
      .data(data)
      .join('text')
        .attr('x', size.width)
        .attr('y', (d: CountInfoEntry) => y(d.Alternative.Name))
        .attr('dy', y.bandwidth() - labelPadding.bottom)
        .attr('text-anchor', 'end')
        .transition()
          .duration(anim.duration)
          .ease(anim.ease)
          .textTween((d: CountInfoEntry) => {
            const num = d3.interpolateRound(0, Math.round(d.Count * 100 / total))
            return (t: number) => num(t) + '%'
          })
          
  }

}
