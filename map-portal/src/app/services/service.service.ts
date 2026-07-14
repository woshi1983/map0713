import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';

@Injectable({
  providedIn: 'root'
})
export class ServiceService {
  constructor (private apollo: Apollo) { }

  // used in ingress
  // get service in namespace
  getApplicationService (clusterId: string, namespace?: string): Observable<any> {
    const filters = [{ key: 'cluster', value: [clusterId] }, {key: 'namespace', value: [namespace]}];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Services(filters: $filters) {
            name
            spec { ports { port }}
          }
        }`,
      variables: {
        filters: filters
      },
      errorPolicy: 'all'
    });
  }
}
