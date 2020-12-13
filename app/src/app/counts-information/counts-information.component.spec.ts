import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CountsInformationComponent } from './counts-information.component';

describe('CountsInformationComponent', () => {
  let component: CountsInformationComponent;
  let fixture: ComponentFixture<CountsInformationComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CountsInformationComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(CountsInformationComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
