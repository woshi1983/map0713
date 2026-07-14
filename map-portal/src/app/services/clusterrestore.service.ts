import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';
import { CreateRestore } from '../classes';

@Injectable({
  providedIn: 'root'
})
export class ClusterRestoreService {
  constructor (private apollo: Apollo) { }

  getClusterRestores (filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Restores(filters: $filters) {
            name
            status
            backup { name}
            location { name}
            cluster {id name}
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

  createClusterRestore (name: string, cluster: string, backup: string, namespace: string[], nsMapping: JSON): Observable<any> {
    const input: CreateRestore = {
      name: name,
      spec: {
        includedNamespaces: namespace,
        excludedNamespaces: []
      },
      backup: backup,
      cluster: cluster
    };
    if (nsMapping) {
      input.spec.namespaceMapping = nsMapping;
    }
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateRestore!){
        CreateRestore(input: $input){
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
