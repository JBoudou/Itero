import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CreateRoundComponent } from './create-round.component';

describe('CreateRoundComponent', () => {
  let component: CreateRoundComponent;
  let fixture: ComponentFixture<CreateRoundComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CreateRoundComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(CreateRoundComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
