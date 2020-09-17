import {Component, Input, OnInit} from '@angular/core';
import {ApiModel} from '../../../models/api.model';

@Component({
  selector: 'app-api-short',
  templateUrl: './api-short.component.html',
  styleUrls: ['./api-short.component.scss']
})
export class ApiShortComponent implements OnInit {
  @Input() apiModel: ApiModel;

  constructor() {
  }

  ngOnInit(): void {
  }

}
