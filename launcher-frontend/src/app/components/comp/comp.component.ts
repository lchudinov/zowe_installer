import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { ApiService } from 'src/app/services/api.service';

@Component({
  selector: 'app-comp',
  templateUrl: './comp.component.html',
  styleUrls: ['./comp.component.scss']
})
export class CompComponent implements OnInit {
  name: string | null = null;

  constructor(
    private route: ActivatedRoute,
    private api: ApiService,
  ) { }

  ngOnInit(): void {
    this.route.paramMap.pipe(map(params => params.get('name'))).subscribe(name => this.name = name);
  }

  start(): void {
    if (this.name) {
      this.api.startComponent(this.name).subscribe(() => console.log(`start ${this.name}`));
    }
  }

  stop(): void {
    if (this.name) {
      this.api.stopComponent(this.name).subscribe(() => console.log(`stop ${this.name}`));
    }
  }

}
