import {Component, OnInit} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {WebsocketService} from '../service/websocket.service';

@Component({
  selector: 'app-author',
  templateUrl: './author.component.html',
  styleUrls: ['./author.component.scss']
})
export class AuthorComponent implements OnInit {

  constructor(private http: HttpClient, private ws: WebsocketService) {
  }

  ngOnInit(): void {
  }
}
