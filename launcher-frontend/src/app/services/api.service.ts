import { Injectable } from '@angular/core';
import { environment } from 'src/environments/environment';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Comp, Log } from '../shared';

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
    console.log(`get logs ${url}`);
    return this.http.get<Log>(url);
  }

  stopComponent(name: string): Observable<void> {
    return this.http.post<void>(`${this.baseURL}/component/${name}/stop`, null);
  }

  startComponent(name: string): Observable<void> {
    return this.http.post<void>(`${this.baseURL}/component/${name}/start`, null);
  }
}
