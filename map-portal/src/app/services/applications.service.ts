import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';
import { CreateApplication, DeleteApplication } from '../classes';

@Injectable({
  providedIn: 'root'
})
export class ApplicationsService {
  constructor (private apollo: Apollo) { }

  getAllApplicationsByOrg (orgId: string): Observable<any> {
    const filters = [{ key: 'organization', value: [orgId] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Applications(page: {limit: -1}, filters: $filters) {
            totalCount
            edge {
              nodes {
                id
                name
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

  createApplication (input: CreateApplication): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateApplication!){
          CreateApplication(input: $input){
            id
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  deleteApplication (id: string): Observable<any> {
    const input: DeleteApplication = {
      id: id
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteApplication!){
          DeleteApplication(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }
}
