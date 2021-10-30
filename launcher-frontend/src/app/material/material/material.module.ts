import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatDividerModule } from '@angular/material/divider';
import { MatListModule } from '@angular/material/list';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';




@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    MatDividerModule,
    MatListModule,
    MatSelectModule,
    MatButtonModule,
  ],
  exports: [
    MatDividerModule,
    MatListModule,
    MatSelectModule,
    MatButtonModule
  ]
})
export class MaterialModule { }
