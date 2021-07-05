import { TestBed } from '@angular/core/testing';

import { ResponsiveBreakpointService } from './responsive-breakpoint.service';

describe('ResponsiveBreakpointService', () => {
  let service: ResponsiveBreakpointService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(ResponsiveBreakpointService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
