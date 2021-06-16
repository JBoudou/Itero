import { Component, OnInit, Input } from '@angular/core';

@Component({
  selector: 'app-info-panel',
  templateUrl: './info-panel.component.html',
  styleUrls: ['./info-panel.component.sass']
})
export class InfoPanelComponent implements OnInit {
  // Input information panel content from poll.component.ts
  @Input() pollInfo: any;

  showInfo: boolean = false;

  constructor() { }

  ngOnInit(): void {
  }

}
