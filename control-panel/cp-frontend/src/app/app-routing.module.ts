import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import {AuthorComponent} from './author/author.component';
import {HomeComponent} from './home/home.component';


const routes: Routes = [
  { path: 'author', component: AuthorComponent },
  { path: '',   component: HomeComponent }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
