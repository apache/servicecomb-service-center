import {Component, ElementRef, Input, OnInit} from '@angular/core';
import SwaggerUI from 'swagger-ui';

@Component({
  selector: 'app-swagger-ui',
  templateUrl: './swagger-ui.component.html',
  styleUrls: ['./swagger-ui.component.css']
})
export class SwaggerUiComponent implements OnInit {

  @Input() schema: any;

  constructor(private el: ElementRef) {
  }

  ngOnInit(): void {
    console.log(this.schema);
    const ui = SwaggerUI({
      url: 'http://petstore.swagger.io/v2/swagger.json',
      spec: this.schema,
      domNode: this.el.nativeElement.querySelector('.swagger-container'),
      deepLinking: true,
      presets: [
        SwaggerUI.presets.apis
      ],
    });
  }

}
