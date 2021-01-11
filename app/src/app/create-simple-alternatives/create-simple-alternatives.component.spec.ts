import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CreateSimpleAlternativesComponent } from './create-simple-alternatives.component';

describe('CreateSimpleAlternativesComponent', () => {
  let component: CreateSimpleAlternativesComponent;
  let fixture: ComponentFixture<CreateSimpleAlternativesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CreateSimpleAlternativesComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(CreateSimpleAlternativesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
