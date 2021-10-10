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
    return this.http.get<Log>(url);
  }
}
