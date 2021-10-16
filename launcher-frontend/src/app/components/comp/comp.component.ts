import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

@Component({
  selector: 'app-comp',
  templateUrl: './comp.component.html',
  styleUrls: ['./comp.component.scss']
})
export class CompComponent implements OnInit {
  name: string | null = null;

  constructor(
    private route: ActivatedRoute
  ) { }

  ngOnInit(): void {
    this.route.paramMap.pipe(map(params => params.get('name'))).subscribe(name => this.name = name);
  }

  start(): void {

  }

  stop(): void {

  }

}
