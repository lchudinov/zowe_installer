import { Component, Input, OnInit } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from 'src/app/services/api.service';
import { Log } from 'src/app/shared';

@Component({
  selector: 'app-log',
  templateUrl: './log.component.html',
  styleUrls: ['./log.component.scss']
})
export class LogComponent implements OnInit {

  @Input() comp?: string;
  lines$: Observable<Log> | undefined;

  constructor(private api: ApiService) {
  }

  ngOnInit(): void {
    this.lines$ = this.api.getLog(this.comp);
  }

}
