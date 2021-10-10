import { Component, Input, OnInit } from '@angular/core';
import { Observable, timer } from 'rxjs';
import { map, switchMap } from 'rxjs/operators';
import { ApiService } from 'src/app/services/api.service';
import { Log } from 'src/app/shared';

@Component({
  selector: 'app-log',
  templateUrl: './log.component.html',
  styleUrls: ['./log.component.scss']
})
export class LogComponent implements OnInit {

  @Input() comp?: string;
  lines$: Observable<string> | undefined;

  constructor(private api: ApiService) {
  }

  ngOnInit(): void {
    this.lines$ = timer(0, 3000).pipe(
      switchMap(() => this.api.getLog(this.comp)),
      map(lines => lines.join('\n'))
    );
  }

}
