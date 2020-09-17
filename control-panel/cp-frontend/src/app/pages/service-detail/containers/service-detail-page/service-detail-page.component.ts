import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {DaoService} from '../../../../services/dao.service';
import {ServiceModel} from '../../../../models/service.model';
import * as _ from 'lodash';
import {Observable, of} from 'rxjs';
import {InstanceModel} from '../../../../models/instance.model';

@Component({
  selector: 'app-map-page',
  templateUrl: './service-detail-page.component.html',
  styleUrls: ['./service-detail-page.component.scss']
})
export class ServiceDetailPageComponent implements OnInit {
  service: ServiceModel;
  instances$: Observable<InstanceModel[]> = of([]);
  consumers$: Observable<ServiceModel[]> = of([]);
  providers$: Observable<ServiceModel[]> = of([]);

  constructor(private route: ActivatedRoute, private daoService: DaoService) {

  }

  ngOnInit() {
    const serviceId = this.route.snapshot.paramMap.get('serviceId');
    this.service = _.find(this.daoService.app.services, s => s.serviceId === serviceId);
    this.instances$ = of(this.service.instances);
    this.consumers$ = of(this.service.consumers);
    this.providers$ = of(this.service.providers);
  }
}
