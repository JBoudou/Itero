import { TestBed } from '@angular/core/testing';

import { CountsInformationService } from './counts-information.service';

describe('CountsInformationService', () => {
  let service: CountsInformationService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(CountsInformationService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
