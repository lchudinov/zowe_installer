import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatDividerModule } from '@angular/material/divider';
import { MatListModule } from '@angular/material/list';
import {MatSelectModule} from '@angular/material/select';



@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    MatDividerModule,
    MatListModule,
    MatSelectModule,
  ],
  exports: [
    MatDividerModule,
    MatListModule,
    MatSelectModule,
  ]
})
export class MaterialModule { }
