import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DisclosePasswordComponent } from './disclose-password.component';

describe('DisclosePasswordComponent', () => {
  let component: DisclosePasswordComponent;
  let fixture: ComponentFixture<DisclosePasswordComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ DisclosePasswordComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(DisclosePasswordComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
