import {Pipe, PipeTransform} from '@angular/core';

@Pipe({
  name: 'shortName'
})
export class ShortNamePipe implements PipeTransform {

  transform(value: string): string {

    const commaIndex: number = value.search('\\s');
    return '' + value[0].toUpperCase() + value[commaIndex + 1].toUpperCase();
  }
}
