import { Component, OnInit } from '@angular/core';

import { CreateService, CREATE_TREE, APP_CREATE_TREE, CreateStepStatus, CreateNextStatus } from './create.service';

@Component({
  selector: 'app-create',
  templateUrl: './create.component.html',
  styleUrls: ['./create.component.sass'],
})
export class CreateComponent implements OnInit {

  canBack: boolean = false;
  canNext: boolean = false;
  isValidate: boolean = false;

  constructor(
    private service: CreateService,
  ) {
    this.service.createStepStatus$.subscribe({
      next: (status: CreateStepStatus) => this.onCreateStepStatus(status),
    });
    this.service.createNextStatus$.subscribe({
      next: (status: CreateNextStatus) => this.onCreateNextStatus(status),
    });
  }

  ngOnInit(): void {
  }

  onBack(): void {
    this.service.back();
  }

  onNext(): void {
    this.service.next();
  }

  private onCreateStepStatus(status: CreateStepStatus): void {
    this.canBack = status.current > 0;
  }

  private onCreateNextStatus(status: CreateNextStatus): void {
    this.canNext = status.validable;
    this.isValidate = status.final;
  }

}
