import * as yaml from 'js-yaml';
import * as _ from 'lodash';
import {ApiModel} from './api.model';

export class SchemaModel {
  schemaId: string;
  schema: string;
  summary: string;
  schemaObj: any;
  apis: ApiModel[];

  public parseSchema(): void {
    if (this.schema) {
      this.schemaObj = yaml.safeLoad(this.schema);
      const paths = this.schemaObj.paths;
      this.apis = [];
      _.keys(this.schemaObj.paths).forEach(p => {
        _.keys(paths[p]).forEach(m => {
          const method = m;
          const pathArray = _(p).split('/');
          let shortPath = '';
          if (pathArray.size() > 3) {
            shortPath = '.../' + pathArray.takeRight(2).join('/');
          } else {
            shortPath = p;
          }
          const overview = [method + ' ' + p, paths[p][method].description].join('\n');
          this.apis.push({method: method.toUpperCase(), path: p, shortPath, overview} as ApiModel);
        });
      });
    }
  }
}
