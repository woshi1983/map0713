import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';

@Injectable({
  providedIn: 'root'
})
export class ReleaseService {
  constructor (private apollo: Apollo) { }

  getAllReleases (num: number, limit: number, component: string): Observable<any> {
    const filters = [{ key: 'component', value: [component] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Releases(page: {number: ${num}, limit: ${limit}}, filters: $filters) {
            totalCount
            edge {
              nodes {
                id
                name
                num
                createdAt
                lastModified
              }
            }
          }
        }`,
      variables: {
        filters: filters
      },
      errorPolicy: 'all'
    });
  }

  getAllReleasesByComponent (component: string): Observable<any> {
    const filters = [{ key: 'component', value: [component] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Releases(page: {limit: -1}, filters: $filters) {
            edge {
              nodes {
                id
                name
                num
              }
            }
          }
        }`,
      variables: {
        filters: filters
      },
      errorPolicy: 'all'
    });
  }

  getReleaseById (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
      query {
        Release(id: "${id}") {
          manifest
          name
          num
        }
      }`,
      errorPolicy: 'all'
    });
  }
}
