import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';

@Injectable({
  providedIn: 'root'
})
export class JobsService {
  constructor (private apollo: Apollo) { }

  getAllJobs (clusterId: string): Observable<any> {
    const filters = [{ key: 'cluster', value: [clusterId] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Jobs(filters: $filters) {
            name
            namespace
            status
            createdAt
            lastModified
          }
        }`,
      variables: {
        filters: filters
      },
      errorPolicy: 'all'
    });
  }

}
