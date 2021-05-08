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

import { cloneDeep, every, includes, isArray, some } from 'lodash';

export const getTabelData = (
  data: any[], // 原数据 已排序
  pageination: PaginationOptions // 分页配置
): any[] => {
  let __data = data || [];
  __data = __data.slice(
    (pageination.pageIndex - 1) * pageination.pageSize,
    pageination.pageIndex * pageination.pageSize
  );
  return __data;
};

export const filterTableData = (
  data: any[], // table数据
  filters?: { [key: string]: any }[]
): any[] => {
  if (!isArray(data)) {
    return [];
  }
  let __data = cloneDeep(data || []);
  __data = __data.filter((item: any) => {
    return every(filters, (filter: any) => {
      if (isArray(filter.value) && isArray(item[filter.field])) {
        return some(item[filter.field], (field) =>
          includes(filter.value, field)
        );
      }
      if (isArray(filter.value)) {
        return includes(filter.value, item[filter.field]);
      }
      return includes(item[filter.field], filter.value);
    });
  });
  return __data;
};

export const getShowPagination = (
  total: number,
  minPageSize: number
): boolean => {
  return total > minPageSize;
};

export const filterTabDataByCategory = (
  data: any[], // 原数据 已排序
  pageination: any, // 分页数据
  filters: FilterItem[] // 过滤项
): { data: any; pageination: any; tableData: any } => {
  if (filters?.length) {
    const __data = filterTableData(data, filters);
    const __pageination = {
      ...pageination,
      total: __data.length,
      pageIndex: 1,
    };
    return {
      data,
      tableData: getTabelData(__data, __pageination),
      pageination: __pageination,
    };
  }
  return {
    data,
    tableData: getTabelData(data, pageination),
    pageination: {
      ...pageination,
      total: data.length,
      pageIndex: 1,
    },
  };
};

interface PaginationOptions {
  total: number;
  pageIndex: number;
  pageSize: number;
  pageSizeOptions: number[];
  [key: string]: any;
}

export interface FilterItem {
  field: string;
  value: string | any[];
}
