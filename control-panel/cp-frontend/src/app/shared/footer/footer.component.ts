import {Component} from '@angular/core';

@Component({
  selector: 'app-footer',
  templateUrl: './footer.component.html',
  styleUrls: ['./footer.component.scss']
})
export class FooterComponent {
  public portal = 'https://servicecomb.apache.org/';
  public blog = 'http://servicecomb.apache.org/year-archive/';
}
