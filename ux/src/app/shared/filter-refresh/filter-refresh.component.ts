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
  ViewChild,
  ViewEncapsulation,
} from '@angular/core';
import { isArray } from 'lodash';
import { ICategorySearchTagItem } from 'ng-devui';
import { FilterItem } from '../toolFunction/tabel.pagination';

@Component({
  selector: 'app-filter-refresh',
  templateUrl: './filter-refresh.component.html',
  styleUrls: ['./filter-refresh.component.less'],
  encapsulation: ViewEncapsulation.Emulated,
})
export class FilterRefreshComponent implements OnInit {
  @Input() category!: ICategorySearchTagItem[];

  @Output() selectedTagsChange = new EventEmitter();
  @Output() refresh = new EventEmitter();

  @ViewChild('categorySearch') categorySearch: any;

  constructor() {}

  ngOnInit(): void {}

  onSelectedTagsChange(e: any): void {
    const filters: FilterItem[] = e.selectedTags.map((item: any) => {
      return {
        field: item.field,
        value: isArray(item.value.value)
          ? item.value.value.map((j: any) => j.id || j.label)
          : item.value.value,
      };
    });
    this.selectedTagsChange.emit(filters);
  }
  onRefresh(): void {
    this.categorySearch.clearFilter();
    this.refresh.emit();
  }
}
