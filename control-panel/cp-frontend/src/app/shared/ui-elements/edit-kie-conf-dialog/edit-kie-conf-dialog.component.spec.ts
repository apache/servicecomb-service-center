import {async, ComponentFixture, TestBed} from '@angular/core/testing';

import {EditKieConfDialogComponent} from './edit-kie-conf-dialog.component';

describe('EditKieConfDialogComponent', () => {
  let component: EditKieConfDialogComponent;
  let fixture: ComponentFixture<EditKieConfDialogComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [EditKieConfDialogComponent]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(EditKieConfDialogComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
