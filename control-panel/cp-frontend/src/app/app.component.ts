  import { Component } from '@angular/core';
import {faCalculator} from '@fortawesome/free-solid-svg-icons';



@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss'],

})
export class AppComponent {
  faCalculator = faCalculator;
  title = 'control-panel';

}
