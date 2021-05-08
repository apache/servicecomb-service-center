/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
import { Component } from '@angular/core';
import { ActivatedRoute, NavigationEnd, Router } from '@angular/router';
import { filter, map } from 'rxjs/operators';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.less'],
})
export class AppComponent {
  constructor(private router: Router, private route: ActivatedRoute) {
    this.router.events
      .pipe(
        filter((event) => event instanceof NavigationEnd),
        map(() => this.route),
        map((activatedRoute: ActivatedRoute) => {
          while (activatedRoute.firstChild) {
            activatedRoute = activatedRoute.firstChild;
          }
          return activatedRoute;
        })
      )
      .subscribe((res) => {
        this.showLeftMenu =
          res.snapshot.data.showLeftMenu === undefined
            ? true
            : res.snapshot.data.showLeftMenu;
      });
  }

  title = 'local-cse';
  showLeftMenu!: boolean;

  menu = [
    {
      title: '服务管理',
      children: [
        { title: '服务列表', link: '/servicelist', linkType: 'routerLink' },
        { title: '实例列表', link: '/instancelist', linkType: 'routerLink' },
      ],
    },
    {
      title: '配置管理',
      children: [{ title: '服务配置', link: '/kie', linkType: 'routerLink' }],
    },
  ];
}
