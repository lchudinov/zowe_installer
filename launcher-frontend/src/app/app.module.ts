import { HttpClientModule } from '@angular/common/http';
import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { CompsComponent } from './components/comps/comps.component';
import { LogComponent } from './components/log/log.component';
import { CompComponent } from './components/comp/comp.component';

@NgModule({
  declarations: [
    AppComponent,
    CompsComponent,
    LogComponent,
    CompComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    HttpClientModule,
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
