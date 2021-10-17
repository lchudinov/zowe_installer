import { Injectable } from '@angular/core';
import { environment } from 'src/environments/environment';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Comp, Log } from '../shared';
import { map } from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class ApiService {

  private baseURL = environment.baseURL;

  constructor(private http: HttpClient) { }

  getComponents(): Observable<Comp[]> {
    return this.http.get<Comp[]>(`${this.baseURL}/components`);
  }

  getLog(comp?: string):Observable<Log> {
    const url = this.baseURL + (comp ? `/component/${comp}` : '') + '/log';
    return this.http.get<Log>(url).pipe(map(log => this.filterEscapeSeqs(log)));
  }

  stopComponent(name: string): Observable<void> {
    return this.http.post<void>(`${this.baseURL}/component/${name}/stop`, null);
  }

  startComponent(name: string): Observable<void> {
    return this.http.post<void>(`${this.baseURL}/component/${name}/start`, null);
  }

  private filterEscapeSeqs(lines: string[]): string[] {
    return lines.map(line => line.replace(/[\u001b]\[\d{2}m/g, '').replace(/[\u001b]\[0;39m/g, ''));
  }
}
