import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CreateResultComponent } from './create-result.component';

describe('CreateResultComponent', () => {
  let component: CreateResultComponent;
  let fixture: ComponentFixture<CreateResultComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CreateResultComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(CreateResultComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
