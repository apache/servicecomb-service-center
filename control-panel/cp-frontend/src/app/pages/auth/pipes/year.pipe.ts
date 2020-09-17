import {Pipe, PipeTransform} from '@angular/core';
import {DatePipe} from '@angular/common';

@Pipe({
  name: 'year'
})
export class YearPipe extends DatePipe implements PipeTransform {
  transform(date: Date): string {
    return super.transform(date, 'y');
  }
}
