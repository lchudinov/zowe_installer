import { Component, Input, OnInit } from '@angular/core';
import { BehaviorSubject, Observable, timer } from 'rxjs';
import { map, mapTo, switchMap, tap } from 'rxjs/operators';
import { ApiService } from 'src/app/services/api.service';
import { LogLevel } from 'src/app/shared';

@Component({
  selector: 'app-log',
  templateUrl: './log.component.html',
  styleUrls: ['./log.component.scss']
})
export class LogComponent implements OnInit {

  @Input() comp?: string;
  lines$: Observable<string> | undefined;
  logLevels: LogLevel[] = ['Error', 'Warning', 'Info', 'Debug', 'Any'];
  logLevel$ = new BehaviorSubject<LogLevel>('Any');

  constructor(private api: ApiService) {
  }

  ngOnInit(): void {
    this.lines$ = this.logLevel$.pipe(
      switchMap(level => timer(0, 3000).pipe(mapTo(level))),
      switchMap(level => this.api.getLog(this.comp, level)),
      map(lines => lines.join('\n'))
    );
  }

  onLevelChange(level: LogLevel): void {
    this.logLevel$.next(level);
  }

}
