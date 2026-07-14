import { Injectable } from '@angular/core';
import { Observable, throwError } from 'rxjs';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { catchError } from 'rxjs/operators';
import {Apollo, gql} from "apollo-angular";

@Injectable({
  providedIn: 'root'
})
export class CommonService {
  constructor (private http: HttpClient, private apollo: Apollo) { }

  private handleError = (error: any) => throwError(error);

  getAppVersion (): Observable<any> {
    const preLoginURL = '/api/version';
    const httpHeaders = new HttpHeaders(
      {
        'Content-Type': 'application/json'
      }
    );
    return this.http.get(preLoginURL, {
      headers: httpHeaders,
      responseType: 'json',
      observe: 'response' as 'body'
    }
    ).pipe(
      catchError(this.handleError)
    );
  }

  getConfiguration (): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Metrics {
            localUsersDisabled
          }
        }`,
      errorPolicy: 'all'
    });
  }
}
