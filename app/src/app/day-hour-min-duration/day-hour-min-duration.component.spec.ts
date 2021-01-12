import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DayHourMinDurationComponent } from './day-hour-min-duration.component';

describe('DayHourMinDurationComponent', () => {
  let component: DayHourMinDurationComponent;
  let fixture: ComponentFixture<DayHourMinDurationComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ DayHourMinDurationComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(DayHourMinDurationComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
