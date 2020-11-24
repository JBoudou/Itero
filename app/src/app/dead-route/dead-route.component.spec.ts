import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DeadRouteComponent } from './dead-route.component';

describe('DeadRouteComponent', () => {
  let component: DeadRouteComponent;
  let fixture: ComponentFixture<DeadRouteComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ DeadRouteComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(DeadRouteComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
