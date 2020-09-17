import {Component} from '@angular/core';
import {Observable, of} from 'rxjs';

import * as _ from 'lodash';

import {DaoService} from '../../../../services/dao.service';
import {map} from 'rxjs/operators';
import {InstanceModel} from '../../../../models/instance.model';

@Component({
  selector: 'app-tables-page',
  templateUrl: './instance-list-page.component.html',
  styleUrls: ['./instance-list-page.component.scss']
})
export class InstanceListPageComponent {
  public instancesTableData$: Observable<InstanceModel[]>;

  constructor(private daoService: DaoService) {
    this.instancesTableData$ = this.daoService.app$.pipe(map(a => {
      const instancesArray = _.map(a.services, s => s.instances);
      return _.flatten(instancesArray);
    }));
    this.daoService.indexServices().subscribe();
  }
}
