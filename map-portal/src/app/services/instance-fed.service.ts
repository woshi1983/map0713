import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';
import { UpdateInstanceFederation } from '../classes';

@Injectable({
  providedIn: 'root'
})
export class InstanceFedService {
  constructor (private apollo: Apollo) { }

  getInstanceFedById (id: string): Observable<any> {
    const filters = [{ key: 'component', value: [id] }];
    return this.apollo.query<any>({
      query: gql`
      query($filters: [QueryFilter!]) {
        InstanceFederations(page: {limit: -1}, filters: $filters) {
          edge {
            nodes {
              id
              status
              instances {
                id
                namespace
                autoRestart
                status {
                  state
                }
                health {
                  available
                  ready
                }
                release {id name num}
                cluster {id name}
                createdAt
                lastModified
              }
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

  updateInstacneFed (param: UpdateInstanceFederation): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateInstanceFederation!){
          UpdateInstanceFederation(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input: param
      },
      errorPolicy: 'all'
    });
  }
}
