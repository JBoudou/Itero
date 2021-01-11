import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CreateGeneralComponent } from './create-general.component';

describe('CreateGeneralComponent', () => {
  let component: CreateGeneralComponent;
  let fixture: ComponentFixture<CreateGeneralComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CreateGeneralComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(CreateGeneralComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
