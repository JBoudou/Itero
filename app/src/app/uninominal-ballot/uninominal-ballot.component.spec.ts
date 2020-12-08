import { ComponentFixture, TestBed } from '@angular/core/testing';

import { UninominalBallotComponent } from './uninominal-ballot.component';

describe('UninominalBallotComponent', () => {
  let component: UninominalBallotComponent;
  let fixture: ComponentFixture<UninominalBallotComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ UninominalBallotComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(UninominalBallotComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
