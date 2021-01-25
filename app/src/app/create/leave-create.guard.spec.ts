import { TestBed } from '@angular/core/testing';

import { LeaveCreateGuard } from './leave-create.guard';

describe('LeaveCreateGuard', () => {
  let guard: LeaveCreateGuard;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    guard = TestBed.inject(LeaveCreateGuard);
  });

  it('should be created', () => {
    expect(guard).toBeTruthy();
  });
});
