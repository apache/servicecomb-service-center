import {Component, OnInit} from '@angular/core';
import {Observable} from 'rxjs';

import {DaoService} from '../../../../services/dao.service';
import {AppModel} from '../../../../models/app.model';
import {ServiceModel} from '../../../../models/service.model';
import {map} from 'rxjs/operators';

@Component({
  selector: 'app-dashboard-page',
  templateUrl: './service-overview-page.component.html',
  styleUrls: ['./service-overview-page.component.scss']
})
export class ServiceOverviewPageComponent implements OnInit {
  public app$: Observable<AppModel>;
  public services$: Observable<ServiceModel[]>;
  serviceN: number;
  instanceN: number;

  constructor(private daoService: DaoService) {
    this.app$ = this.daoService.app$;
    this.services$ = this.daoService.app$.pipe(map(a => a.services));
  }

  ngOnInit(): void {
    this.daoService.indexServices(true).subscribe(d => {
      this.serviceN = this.daoService.app.servicesN;
      this.instanceN = this.daoService.app.instanceN;
    });
  }
}
