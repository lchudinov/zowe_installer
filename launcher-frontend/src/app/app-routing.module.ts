import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { CompComponent } from './components/comp/comp.component';

const routes: Routes = [
  {path: 'component/:name', component: CompComponent}
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
