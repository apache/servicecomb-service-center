import {ServiceModel} from './service.model';
import * as _ from 'lodash';

export class InstanceModel {
  instanceId: string;
  service: ServiceModel;
  endpoints: string[];
  hostName: string;
  status: string;
  timeStamp: string;
  modTimestamp: string;

  public endpointsOverview(): string {
    if (this.endpoints.length > 1) {
      return _.join(this.endpoints, '\n');
    } else {
      return '';
    }
  }

  public endpointsSimple(): string {
    if (this.endpoints.length > 1) {
      return this.endpoints[0] + ', ...';
    } else {
      return this.endpoints[0];
    }
  }
}
