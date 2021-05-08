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
import {
  Component,
  EventEmitter,
  Input,
  OnInit,
  Output,
  TemplateRef,
} from '@angular/core';

@Component({
  selector: 'app-action-menu',
  templateUrl: './action-menu.component.html',
  styleUrls: ['./action-menu.component.less'],
})
export class ActionMenuComponent implements OnInit {
  @Input() actions!: ActionItem[];
  @Input() menuText?: string;
  @Input() maxShowNum = 3;
  @Input() onToggle?: (e?: any) => void;
  @Input() moreBtnRef?: TemplateRef<any>;

  @Output() menuClick = new EventEmitter<ActionItem>();

  constructor() {}

  moreActions?: ActionItem[];
  buttonActions?: ActionItem[];

  ngOnInit(): void {
    this.actions = this.actions.filter(
      (item) => item.show || item.show === undefined
    );

    if (this.actions.length > this.maxShowNum) {
      this.buttonActions = this.actions.slice(0, this.maxShowNum - 1);
      this.moreActions = this.actions.slice(this.maxShowNum - 1);
    } else {
      this.buttonActions = this.actions;
    }
  }

  toggle(e?: any): void {
    if (this.onToggle) {
      this.onToggle(e);
    }
  }

  onClick(e: ActionItem): any {
    if (typeof e.click === 'function') {
      e.click(e);
    }
    this.menuClick.emit(e);
  }
}

export interface ActionItem {
  id: string;
  label: string;
  disabled?: boolean;
  show?: boolean;
  tip?: string;
  click?: (e: ActionItem) => void;
  [key: string]: any;
}
