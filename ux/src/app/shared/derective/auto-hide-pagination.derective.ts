/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// tslint:disable:variable-name
import {
  Directive,
  ElementRef,
  Input,
  OnChanges,
  SimpleChanges,
} from '@angular/core';

@Directive({
  // tslint:disable-next-line:directive-selector
  selector: '[autoHide]',
})
export class AutoHidePaginationDirective implements OnChanges {
  @Input() total!: number;
  @Input() pageSizeOptions: number[] = [5];

  @Input() autoHide = false;

  constructor(private el: ElementRef) {
    this.__display = this.el.nativeElement.style.display;
  }
  private __display = '';

  private hide(): void {
    this.el.nativeElement.style.display = 'none';
  }
  private show(): void {
    this.el.nativeElement.style.display = this.__display;
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (this.autoHide && this.total <= this.pageSizeOptions[0]) {
      this.hide();
    } else {
      this.show();
    }
  }
}
