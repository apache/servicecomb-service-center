import {Component} from '@angular/core';
import {Observable, of} from 'rxjs';


import {DaoService} from '../../../../services/dao.service';
import {ServiceModel} from '../../../../models/service.model';
import {map} from 'rxjs/operators';

@Component({
  selector: 'app-tables-page',
  templateUrl: './service-list-page.component.html',
  styleUrls: ['./service-list-page.component.scss']
})
export class ServiceListPageComponent {
  public servicesTableData$: Observable<ServiceModel[]>;

  constructor(private daoService: DaoService) {

    this.servicesTableData$ = this.daoService.app$.pipe(map(a => a.services));
    this.daoService.indexServices().subscribe();
  }
}
