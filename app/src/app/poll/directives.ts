import { Directive, ViewContainerRef } from '@angular/core';

@Directive({
  selector: '[PollBallot]',
})
export class PollBallotDirective {
  constructor(public viewContainerRef: ViewContainerRef) { }
}

@Directive({
  selector: '[PollInformation]',
})
export class PollInformationDirective {
  constructor(public viewContainerRef: ViewContainerRef) { }
}

export interface PollSubComponent {
  pollSegment: string;
}
