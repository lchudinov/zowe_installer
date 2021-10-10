import { Component, OnInit } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from 'src/app/services/api.service';
import { Comp } from 'src/app/shared';

@Component({
  selector: 'app-comps',
  templateUrl: './comps.component.html',
  styleUrls: ['./comps.component.scss']
})
export class CompsComponent implements OnInit {
  comps$: Observable<Comp[]>;

  constructor(private api: ApiService) {
    this.comps$ = this.api.getComponents();
  }

  ngOnInit(): void {
  }

}
