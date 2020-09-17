import {Component, Input, OnInit} from '@angular/core';
import {ServiceModel} from '../../../../models/service.model';

@Component({
  selector: 'app-service-overview',
  templateUrl: './service-overview.component.html',
  styleUrls: ['./service-overview.component.scss']
})
export class ServiceOverviewComponent implements OnInit {
  @Input() service: ServiceModel;

  constructor() {
  }

  ngOnInit(): void {
  }

}
