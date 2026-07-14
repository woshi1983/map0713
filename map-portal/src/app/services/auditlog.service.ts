import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';

@Injectable({
  providedIn: 'root'
})
export class AuditlogService {
  constructor (private apollo: Apollo) { }

  getAllAuditlogs (num: number, limit: number, filters: any[]): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          AuditLogs(page: {number: ${num}, limit: ${limit}}, filters: $filters){
            totalCount
            edge {
              nodes {
                userName
                objectName
                objectType
                action
                organization {name}
                createdAt
              }
            }
          }
        }`,
      errorPolicy: 'all',
      variables: {
        filters: filters
      }
    });
  }
}
